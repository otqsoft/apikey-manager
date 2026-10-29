package logic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"apikey-manager/internal/adapter"
	"apikey-manager/internal/adapter/anthropic"
	"apikey-manager/internal/adapter/openai"
	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/pkg/response"
	"apikey-manager/internal/pkg/semaphore"
	"apikey-manager/internal/pkg/util"
	"apikey-manager/internal/service"
)

func init() {
	service.Chat = &chatService{}
}

type chatService struct{}

// CreateSession 创建会话（先校验 key 可用性）
func (s *chatService) CreateSession(ctx context.Context, userID uint, req *model.CreateSessionReq) (*model.ChatSession, error) {
	k, _, err := service.Key.CheckAvailable(ctx, req.KeyID)
	if err != nil {
		return nil, err
	}
	id, err := g.DB().Model("chat_sessions").InsertAndGetId(g.Map{
		"user_id": userID,
		"key_id":  k.ID,
		"model":   req.Model,
		"title":   "",
	})
	if err != nil {
		return nil, err
	}
	return &model.ChatSession{ID: uint(id), UserID: userID, KeyID: k.ID, Model: req.Model}, nil
}

// ListSessions 会话分页；admin 全部，user 仅本人；按创建时间倒序
func (s *chatService) ListSessions(ctx context.Context, userID uint, isAdmin bool, page, size int) (*model.PageResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	m := g.DB().Model("chat_sessions s").
		LeftJoin("api_keys k", "k.id = s.key_id").
		LeftJoin("providers p", "p.id = k.provider_id")
	if !isAdmin {
		m = m.Where("s.user_id", userID)
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	records, err := m.Fields("s.*, k.name as key_name, p.name as provider_name").
		Page(page, size).OrderDesc("s.created_at").OrderDesc("s.id").All()
	if err != nil {
		return nil, err
	}
	return &model.PageResult{List: safeList(records), Total: total, Page: page, PageSize: size}, nil
}

// ListMessages 消息正序；越权 403
func (s *chatService) ListMessages(ctx context.Context, sessionID, userID uint, isAdmin bool) ([]*model.ChatMessage, error) {
	var sess model.ChatSession
	if err := g.DB().Model("chat_sessions").Where("id", sessionID).Scan(&sess); err != nil {
		return nil, scanErr(err, "会话不存在")
	}
	if sess.ID == 0 {
		return nil, errs.NotFound("会话不存在")
	}
	if !isAdmin && sess.UserID != userID {
		return nil, errs.Forbidden("无权查看该会话消息")
	}
	list := make([]*model.ChatMessage, 0)
	if err := g.DB().Model("chat_messages").Where("session_id", sessionID).OrderAsc("id").Scan(&list); err != nil {
		return nil, err
	}

	// P4.2 冗余展示字段：model 取会话快照；provider_name 联 api_keys→providers，取不到为空串
	providerName := ""
	if sess.KeyID > 0 {
		rec, qerr := g.DB().Model("api_keys k").
			LeftJoin("providers p", "p.id = k.provider_id").
			Where("k.id", sess.KeyID).
			Fields("p.name").One()
		if qerr == nil && rec != nil {
			providerName = rec["name"].String()
		}
	}
	for _, m := range list {
		m.Model = sess.Model
		m.ProviderName = providerName
	}
	return list, nil
}

// DeleteSession 仅本人或 admin；级联删除消息
func (s *chatService) DeleteSession(ctx context.Context, sessionID, userID uint, isAdmin bool) error {
	var sess model.ChatSession
	if err := g.DB().Model("chat_sessions").Where("id", sessionID).Scan(&sess); err != nil {
		return scanErr(err, "会话不存在")
	}
	if sess.ID == 0 {
		return errs.NotFound("会话不存在")
	}
	if !isAdmin && sess.UserID != userID {
		return errs.Forbidden("无权删除该会话")
	}
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Model("chat_messages").Where("session_id", sessionID).Delete(); err != nil {
		return err
	}
	if _, err := tx.Model("chat_sessions").Where("id", sessionID).Delete(); err != nil {
		return err
	}
	return tx.Commit()
}

// Complete 对话转发主流程：可用性校验 → 权限校验 → 信号量限流 → 超时控制 → 双协议转发 → 落库计费
func (s *chatService) Complete(r *ghttp.Request, userID uint, req *model.ChatCompletionReq) error {
	ctx := r.Context()

	// 1. 可用性校验
	k, p, err := service.Key.CheckAvailable(ctx, req.KeyID)
	if err != nil {
		return err
	}
	// 2. 数据归属：非 admin 需为密钥 owner 或被分配用户
	isAdmin := middlewareIsAdmin(userID, ctx)
	if !isAdmin {
		owned, err := g.DB().Model("api_keys").Where("id", k.ID).Where("owner_id", userID).Count()
		if err != nil {
			return err
		}
		if owned == 0 {
			cnt, err := g.DB().Model("key_user").Where("key_id", k.ID).Where("user_id", userID).Count()
			if err != nil {
				return err
			}
			if cnt == 0 {
				return errs.Forbidden("无权使用该密钥")
			}
		}
	}

	// 3. 会话确定：未传 session_id 则自动创建
	sessionID, err := s.ensureSession(ctx, userID, req)
	if err != nil {
		return err
	}

	// 4. 组装归一化消息 + 选取适配器
	msgs := make([]adapter.Message, 0, len(req.Messages))
	lastUserContent := ""
	for _, m := range req.Messages {
		msgs = append(msgs, adapter.Message{Role: m.Role, Content: m.Content})
		if m.Role == "user" {
			lastUserContent = m.Content
		}
	}
	var ad adapter.Adapter
	switch p.ApiFormat {
	case "anthropic":
		ad = anthropic.New(p.BaseURL, k.APIKey)
	default:
		ad = openai.New(p.BaseURL, k.APIKey)
	}

	// 5. 供应商维度超时控制
	timeout := p.TimeoutSeconds
	if timeout <= 0 {
		timeout = 60
	}
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 6. 供应商维度并发信号量限流
	release, err := semaphore.Acquire(callCtx, p.ID, p.MaxConcurrency)
	if err != nil {
		return errs.BadRequest("当前并发已满，请稍后重试")
	}
	defer release()

	// 7. 组装上游请求
	url, header, reqBody := ad.BuildRequest(req.Model, msgs, req.Stream)
	bodyBytes, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	for key, values := range header {
		for _, v := range values {
			httpReq.Header.Set(key, v)
		}
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		// 上游连接失败：落用户消息并补会话标题，避免会话空白；不累计费用
		s.recordUserMessage(ctx, sessionID, lastUserContent)
		g.Log().Errorf(ctx, "上游请求失败: %v", err)
		if req.Stream {
			sseRespondError(r, normalizeErrorText("无法连接上游模型服务："+err.Error()))
			return nil
		}
		return errs.BadRequest("无法连接上游模型服务，请检查供应商地址与网络")
	}
	defer resp.Body.Close()

	// 上游非 2xx：解析并归一错误文本（压缩空白、截断 200 字，避免 HTML 错误体刷屏）
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		upstreamMsg := simplifyUpstreamError(raw)
		s.recordUserMessage(ctx, sessionID, lastUserContent)
		g.Log().Warningf(ctx, "上游返回状态码 %d: %s", resp.StatusCode, upstreamMsg)
		if req.Stream {
			sseRespondError(r, upstreamMsg)
			return nil
		}
		return errs.BadRequest("上游模型服务返回错误: " + upstreamMsg)
	}

	if req.Stream {
		// P4.1 上游不支持 SSE（Content-Type 非 text/event-stream）时，按一次性 JSON 解析并补发标准事件，避免挂起报错
		if !isEventStream(resp.Header) {
			return s.handleStreamFallback(r, callCtx, ad, resp, k.ID, p.ID, req.Model, sessionID, lastUserContent)
		}
		return s.handleStream(r, callCtx, ad, resp, k.ID, p.ID, req.Model, sessionID, lastUserContent)
	}
	return s.handleNonStream(r, callCtx, ad, resp, k.ID, p.ID, req.Model, sessionID, lastUserContent)
}

// isEventStream 判断上游响应是否为 SSE 流（大小写不敏感，兼容带 charset 的 Content-Type）
func isEventStream(header http.Header) bool {
	ct := strings.ToLower(header.Get("Content-Type"))
	return strings.Contains(ct, "text/event-stream")
}

// ensureSession 确定会话：已传则校验归属，未传则按 key+model 创建
func (s *chatService) ensureSession(ctx context.Context, userID uint, req *model.ChatCompletionReq) (uint, error) {
	if req.SessionID > 0 {
		var sess model.ChatSession
		if err := g.DB().Model("chat_sessions").Where("id", req.SessionID).Scan(&sess); err != nil {
			return 0, scanErr(err, "会话不存在")
		}
		if sess.ID == 0 {
			return 0, errs.NotFound("会话不存在")
		}
		isAdmin := middlewareIsAdmin(userID, ctx)
		if !isAdmin && sess.UserID != userID {
			return 0, errs.Forbidden("无权使用该会话")
		}
		return sess.ID, nil
	}
	id, err := g.DB().Model("chat_sessions").InsertAndGetId(g.Map{
		"user_id": userID,
		"key_id":  req.KeyID,
		"model":   req.Model,
		"title":   "",
	})
	return uint(id), err
}

// handleNonStream 非流式：一次性解析、计费落库、返回归一化结果
func (s *chatService) handleNonStream(r *ghttp.Request, ctx context.Context, ad adapter.Adapter, resp *http.Response,
	keyID, providerID uint, modelName string, sessionID uint, lastUserContent string) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	content, reasoning, usage, err := ad.ParseNonStream(raw)
	if err != nil {
		g.Log().Errorf(ctx, "非流式响应解析失败: %v, body=%s", err, truncate(string(raw), 500))
		return errs.BadRequest("上游响应解析失败")
	}
	s.recordUserMessage(ctx, sessionID, lastUserContent)
	cost := calcCost(ctx, providerID, modelName, usage)
	messageID, err := s.persistAssistantMessage(ctx, sessionID, content, reasoning, usage, cost)
	if err != nil {
		return err
	}
	s.afterMessage(ctx, keyID, cost)
	response.Success(r, model.ChatNonStreamResp{
		Content:   content,
		Reasoning: reasoning,
		Usage: model.ChatUsage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			CachedTokens:     usage.CachedTokens,
			TotalTokens:      usage.TotalTokens,
		},
		Cost:      cost,
		MessageID: messageID,
	})
	return nil
}

// sseHeaders 统一设置 SSE 响应头（正常流式与非 SSE 降级共用）
func sseHeaders(r *ghttp.Request) {
	header := r.Response.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
}

// handleStream 流式：逐行转发 SSE 事件；正文 delta 与思考 reasoning 独立累加，结束统一落库计费
func (s *chatService) handleStream(r *ghttp.Request, ctx context.Context, ad adapter.Adapter, resp *http.Response,
	keyID, providerID uint, modelName string, sessionID uint, lastUserContent string) error {
	sseHeaders(r)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	var content, reasoning strings.Builder
	var lastUsage *adapter.Usage
	for scanner.Scan() {
		part, perr := ad.StreamParse(scanner.Text())
		if perr != nil || part == nil {
			continue
		}
		if part.Usage != nil {
			lastUsage = part.Usage
		}
		// 思考增量独立推送 reasoning 事件（无思考模型时该分支不触发）
		if part.Reasoning != "" {
			reasoning.WriteString(part.Reasoning)
			sseWrite(r, g.Map{"type": "reasoning", "content": part.Reasoning})
		}
		if part.Content != "" {
			content.WriteString(part.Content)
			sseWrite(r, g.Map{"type": "delta", "content": part.Content})
		}
		if part.Done {
			break
		}
	}
	userRecorded := false
	if err := scanner.Err(); err != nil {
		g.Log().Errorf(ctx, "流式读取异常: %v", err)
		// 读取异常分支同样落用户消息并补会话标题，保证失败对话有用户记录
		s.recordUserMessage(ctx, sessionID, lastUserContent)
		userRecorded = true
	}

	usage := adapter.Usage{}
	if lastUsage != nil {
		usage = *lastUsage
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	cost := calcCost(ctx, providerID, modelName, usage)

	// 落库：用户消息 + assistant 消息（异常分支已记录则不重复写）
	if !userRecorded {
		s.recordUserMessage(ctx, sessionID, lastUserContent)
	}
	messageID, err := s.persistAssistantMessage(ctx, sessionID, content.String(), reasoning.String(), usage, cost)
	if err != nil {
		g.Log().Errorf(ctx, "assistant 消息落库失败: %v", err)
	}
	s.afterMessage(ctx, keyID, cost)

	// usage 事件 + done 事件
	sseWrite(r, g.Map{
		"type":               "usage",
		"prompt_tokens":      usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
		"cached_tokens":      usage.CachedTokens,
		"total_tokens":       usage.TotalTokens,
		"cost":               cost,
	})
	sseWrite(r, g.Map{"type": "done", "session_id": sessionID, "message_id": messageID})
	r.Response.Flush()
	return nil
}

// fallbackChunkInterval 非 SSE 降级逐片下发间隔（契约 8.2 要求 12~30ms，取 20ms 折中，兼顾观感与总耗时）
const fallbackChunkInterval = 20 * time.Millisecond

// markdownListRe Markdown 列表行：-/*/+ 或 数字. / 数字) 开头
var markdownListRe = regexp.MustCompile(`^\s*(?:[-*+]\s+|\d+[.)]\s+)`)

// handleStreamFallback 非 SSE 降级（契约 8.2 重点修复）：请求声明 stream=true 但上游返回一次性 JSON
// （Content-Type 非 text/event-stream）时，后端把 reasoning 与正文按 Markdown 结构切片后逐片模拟流式下发：
// 顺序 reasoning* → delta* → usage → done；片间 20ms 间隔且受供应商超时 ctx 约束（取消立即退出）；
// 全部片发完后再一次性落库（用户消息 + assistant 完整正文/reasoning）并累计 quota，不逐片写库。
func (s *chatService) handleStreamFallback(r *ghttp.Request, ctx context.Context, ad adapter.Adapter, resp *http.Response,
	keyID, providerID uint, modelName string, sessionID uint, lastUserContent string) error {
	sseHeaders(r)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		sseRespondError(r, "上游响应读取失败")
		return nil
	}
	content, reasoning, usage, err := ad.ParseNonStream(raw)
	if err != nil {
		g.Log().Errorf(ctx, "非 SSE 降级响应解析失败: %v, body=%s", err, truncate(string(raw), 500))
		sseRespondError(r, "上游响应解析失败")
		return nil
	}

	// 1) 先逐片下发思考过程，再逐片下发正文；任一步被 ctx 取消则立即退出，不再落库/发 usage
	if !emitFallbackChunks(r, ctx, "reasoning", chunkReasoning(reasoning)) {
		return nil
	}
	if !emitFallbackChunks(r, ctx, "delta", chunkMarkdown(content)) {
		return nil
	}

	// 2) 全部片发完后一次性落库：用户消息 + assistant 完整正文/reasoning，再按公式计费累计 quota
	s.recordUserMessage(ctx, sessionID, lastUserContent)
	cost := calcCost(ctx, providerID, modelName, usage)
	messageID, err := s.persistAssistantMessage(ctx, sessionID, content, reasoning, usage, cost)
	if err != nil {
		g.Log().Errorf(ctx, "非 SSE 降级 assistant 消息落库失败: %v", err)
	}
	s.afterMessage(ctx, keyID, cost)

	// 3) usage + done 收尾事件
	sseWrite(r, g.Map{
		"type":               "usage",
		"prompt_tokens":      usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
		"cached_tokens":      usage.CachedTokens,
		"total_tokens":       usage.TotalTokens,
		"cost":               cost,
	})
	sseWrite(r, g.Map{"type": "done", "session_id": sessionID, "message_id": messageID})
	r.Response.Flush()
	return nil
}

// emitFallbackChunks 逐片写入 SSE 事件并立即 Flush；片间等待 fallbackChunkInterval，
// 等待期间若 ctx.Done()（供应商 timeout_seconds 超时或客户端断开）则返回 false 立即退出。
func emitFallbackChunks(r *ghttp.Request, ctx context.Context, eventType string, chunks []string) bool {
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		select {
		case <-ctx.Done():
			g.Log().Warningf(ctx, "非 SSE 降级流式被取消（超时或断开），事件=%s", eventType)
			return false
		case <-time.After(fallbackChunkInterval):
		}
		sseWrite(r, g.Map{"type": eventType, "content": chunk})
		r.Response.Flush()
	}
	return true
}

// chunkReasoning 思考过程为纯文本，按句末标点与换行切分（。！？!?；;，, 及换行），保持语义完整。
func chunkReasoning(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var chunks []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			chunks = append(chunks, b.String())
			b.Reset()
		}
	}
	for _, r := range text {
		b.WriteRune(r)
		switch r {
		case '。', '！', '？', '!', '?', '；', ';', '，', ',', '\n':
			flush()
		}
	}
	flush()
	return chunks
}

// chunkMarkdown 正文按 Markdown 结构单元切分（契约 8.2：不得破坏代码块/列表项/标题整体）：
//   - ``` 围栏代码块：整块作为一个不可拆单元；
//   - 标题行（# 开头）、列表行（-/*/+、数字.、数字)）：整行作为一个单元；
//   - 其余普通段落：按句末标点/换行切分，保留换行符便于前端纯增量拼接。
func chunkMarkdown(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var chunks []string
	lines := strings.Split(text, "\n")
	inFence := false
	var fence []string
	flushFence := func() {
		if len(fence) > 0 {
			chunks = append(chunks, strings.Join(fence, "\n")+"\n")
			fence = nil
		}
	}
	var para strings.Builder
	flushPara := func() {
		if para.Len() > 0 {
			chunks = append(chunks, para.String())
			para.Reset()
		}
	}
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			flushPara()
			inFence = !inFence
			fence = append(fence, line)
			if !inFence {
				flushFence() // 围栏闭合，整块输出
			}
			continue
		}
		if inFence {
			fence = append(fence, line)
			continue
		}
		switch {
		case trim == "":
			flushPara()
			chunks = append(chunks, "\n")
		case strings.HasPrefix(trim, "#") || markdownListRe.MatchString(trim):
			// 标题 / 列表项整体保留，不拆句
			flushPara()
			chunks = append(chunks, line+"\n")
		default:
			para.WriteString(line)
			para.WriteString("\n")
			if strings.ContainsAny(trim, "。！？!?；;") {
				flushPara()
			}
		}
	}
	flushFence()
	flushPara()
	return chunks
}

// persistUserMessage 落库最新一条用户消息（空内容跳过）
func (s *chatService) persistUserMessage(ctx context.Context, sessionID uint, content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	_, err := g.DB().Model("chat_messages").Insert(g.Map{
		"session_id": sessionID,
		"role":       "user",
		"content":    content,
	})
	return err
}

// persistAssistantMessage 落库 assistant 消息（含思考过程 reasoning_content）并返回消息 ID
func (s *chatService) persistAssistantMessage(ctx context.Context, sessionID uint, content, reasoning string, usage adapter.Usage, cost float64) (uint, error) {
	id, err := g.DB().Model("chat_messages").InsertAndGetId(g.Map{
		"session_id":        sessionID,
		"role":              "assistant",
		"content":           content,
		"reasoning_content": reasoning,
		"prompt_tokens":     usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
		"cached_tokens":     usage.CachedTokens,
		"total_tokens":      usage.TotalTokens,
		"cost":              cost,
	})
	return uint(id), err
}

// recordUserMessage 落库用户消息并在会话无标题时生成标题；成功/失败分支共用，不累计费用。
func (s *chatService) recordUserMessage(ctx context.Context, sessionID uint, content string) {
	_ = s.persistUserMessage(ctx, sessionID, content)
	s.updateSessionTitle(ctx, sessionID, content)
}

// updateSessionTitle 首条用户消息前 20 字自动生成会话标题；已有标题不覆盖；空内容跳过。
func (s *chatService) updateSessionTitle(ctx context.Context, sessionID uint, userContent string) {
	if strings.TrimSpace(userContent) == "" {
		return
	}
	var sess model.ChatSession
	if err := g.DB().Model("chat_sessions").Where("id", sessionID).Scan(&sess); err != nil || sess.ID == 0 {
		return
	}
	if sess.Title != "" {
		return
	}
	_, _ = g.DB().Model("chat_sessions").Where("id", sessionID).
		Update(g.Map{"title": util.TruncateRunes(userContent, 20)})
}

// afterMessage：仅在 assistant 消息产生费用（cost>0）时累计进 key.used_quota；
// 会话标题由 recordUserMessage/updateSessionTitle 统一负责，失败分支不累计费用。
func (s *chatService) afterMessage(ctx context.Context, keyID uint, cost float64) {
	if cost > 0 {
		_, _ = g.DB().Model("api_keys").Where("id", keyID).Increment("used_quota", cost)
	}
}

// calcCost 契约费用公式：
// cost = (未命中缓存输入token)/1e6*input_price + 命中缓存输入token/1e6*input_price_cached
//        + 输出token/1e6*output_price；模型缺失/停用按 0 计。
func calcCost(ctx context.Context, providerID uint, modelName string, u adapter.Usage) float64 {
	record, err := g.DB().Model("models").Where("provider_id", providerID).Where("name", modelName).One()
	if err != nil || record == nil {
		return 0
	}
	inputPrice := record["input_price"].Float64()
	cachedPrice := record["input_price_cached"].Float64()
	outputPrice := record["output_price"].Float64()
	regular := u.PromptTokens - u.CachedTokens
	if regular < 0 {
		regular = 0
	}
	cost := float64(regular)/1e6*inputPrice +
		float64(u.CachedTokens)/1e6*cachedPrice +
		float64(u.CompletionTokens)/1e6*outputPrice
	// 保留 6 位小数（元）
	return math.Round(cost*1e6) / 1e6
}

// sseWrite 写出一条 SSE data 事件并立即 flush
func sseWrite(r *ghttp.Request, v interface{}) {
	b, _ := json.Marshal(v)
	r.Response.Write("data: " + string(b) + "\n\n")
	r.Response.Flush()
}

// sseRespondError 流式错误事件；错误文本统一经 normalizeErrorText 压缩截断
func sseRespondError(r *ghttp.Request, msg string) {
	header := r.Response.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	sseWrite(r, g.Map{"type": "error", "msg": normalizeErrorText(msg)})
	r.Response.Flush()
}

// normalizeErrorText 非流式与流式 error 共用的错误文本归一：
// 压缩连续空白/换行/制表符为单个空格，截断到 200 字符以内，避免 HTML 错误体刷屏。
func normalizeErrorText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) > 200 {
		r = r[:200]
	}
	return strings.TrimSpace(string(r))
}

// simplifyUpstreamError 从上游错误体中提取可读文本（兼容 OpenAI/Anthropic 错误结构），
// 提取后统一走 normalizeErrorText 压缩截断。
func simplifyUpstreamError(raw []byte) string {
	if len(raw) == 0 {
		return "上游服务无响应内容"
	}
	var obj struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
		Message string `json:"message"`
		Type    string `json:"type"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		if obj.Error.Message != "" {
			return normalizeErrorText(obj.Error.Message)
		}
		if obj.Message != "" {
			return normalizeErrorText(obj.Message)
		}
	}
	return normalizeErrorText(string(raw))
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// middlewareIsAdmin 避免 logic 反向依赖 middleware 包造成循环引用，这里直接查角色。
func middlewareIsAdmin(userID uint, ctx context.Context) bool {
	role, err := g.DB().Model("users").Where("id", userID).Value("role")
	if err != nil {
		return false
	}
	return role.String() == "admin"
}
