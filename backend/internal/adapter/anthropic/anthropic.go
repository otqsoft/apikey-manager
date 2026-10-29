// Package anthropic Anthropic Claude 协议适配器。
//
// 协议要点（契约 3.2）：
//   - URL：{base_url}/v1/messages（base_url 不含 /v1，需显式拼接）。
//   - Header：x-api-key: {api_key}；anthropic-version: 2023-06-01；Content-Type。
//   - Body：{model, max_tokens:4096 必填, system:顶层字符串, messages:[{role,content}], stream}；
//     system 角色消息需从 messages 抽出合并为顶层 system 字符串。
//   - 非流式：content 数组中 type=="text" 的 text 拼接；usage.input_tokens / output_tokens；
//     缓存命中取 usage.cache_read_input_tokens。
//   - 流式事件：message_start（input_tokens 初值）、content_block_delta（delta.text 增量）、
//     message_delta（usage.output_tokens 增量）、message_stop（结束）。
package anthropic

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/adapter"
)

// Adapter Anthropic 适配器（流式解析为有状态：跨事件累计 token 用量）
type Adapter struct {
	BaseURL string
	APIKey  string

	inputTokens    int
	cachedTokens   int
	outputTokens   int
}

// New 构造适配器
func New(baseURL, apiKey string) *Adapter {
	return &Adapter{BaseURL: baseURL, APIKey: apiKey}
}

// BuildRequest 组装 Anthropic Messages 请求
func (a *Adapter) BuildRequest(model string, msgs []adapter.Message, stream bool) (string, http.Header, interface{}) {
	url := strings.TrimRight(a.BaseURL, "/") + "/v1/messages"
	header := http.Header{}
	header.Set("x-api-key", a.APIKey)
	header.Set("anthropic-version", "2023-06-01")
	header.Set("Content-Type", "application/json")

	// system 消息抽出合并为顶层字符串；Anthropic messages 仅允许 user/assistant
	systemParts := make([]string, 0)
	outMsgs := make([]adapter.Message, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "system" {
			if strings.TrimSpace(m.Content) != "" {
				systemParts = append(systemParts, strings.TrimSpace(m.Content))
			}
			continue
		}
		outMsgs = append(outMsgs, adapter.Message{Role: m.Role, Content: m.Content})
	}
	body := g.Map{
		"model":      model,
		"max_tokens": 4096, // Anthropic 必填，默认 4096（前端可扩展预留）
		"messages":   outMsgs,
		"stream":     stream,
	}
	if len(systemParts) > 0 {
		body["system"] = strings.Join(systemParts, "\n\n")
	}
	return url, header, body
}

// ------- 非流式响应结构 -------

type antUsage struct {
	InputTokens          int `json:"input_tokens"`
	OutputTokens         int `json:"output_tokens"`
	CacheReadInputTokens int `json:"cache_read_input_tokens"`
}

type antContentBlock struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Thinking  string `json:"thinking"` // type=="thinking" 时的思考过程文本
}

type antError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type antResp struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Content []antContentBlock `json:"content"`
	Usage   antUsage          `json:"usage"`
	Error   *antError         `json:"error"`
}

// ParseNonStream 解析非流式响应；text 块拼为正文，thinking 块拼为思考过程
func (a *Adapter) ParseNonStream(body []byte) (content, reasoning string, usage adapter.Usage, err error) {
	var resp antResp
	if err = json.Unmarshal(body, &resp); err != nil {
		return "", "", adapter.Usage{}, err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return "", "", adapter.Usage{}, nil
	}
	var textSB, thinkSB strings.Builder
	for _, c := range resp.Content {
		switch c.Type {
		case "text":
			textSB.WriteString(c.Text)
		case "thinking":
			thinkSB.WriteString(c.Thinking)
		}
	}
	usage = adapter.Usage{
		PromptTokens:     resp.Usage.InputTokens,
		CompletionTokens: resp.Usage.OutputTokens,
		CachedTokens:     resp.Usage.CacheReadInputTokens,
		TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}
	return textSB.String(), thinkSB.String(), usage, nil
}

// ------- 流式事件结构 -------

type antStreamEvent struct {
	Type    string `json:"type"`
	Message struct {
		Usage antUsage `json:"usage"`
	} `json:"message"`
	Delta struct {
		Type     string   `json:"type"`
		Text     string   `json:"text"`
		Thinking string   `json:"thinking"` // thinking_delta 事件中的思考增量
		Usage    antUsage `json:"usage"`
	} `json:"delta"`
	Usage antUsage `json:"usage"`
}

// StreamParse 逐行解析 Anthropic SSE；有状态累计用量，message_stop 时返回完整用量；
// content_block_delta 按 delta.type 区分正文文本(text_delta)与思考增量(thinking_delta)。
func (a *Adapter) StreamParse(line string) (*adapter.StreamPart, error) {
	part := &adapter.StreamPart{}
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "event:") {
		return part, nil
	}
	if !strings.HasPrefix(line, "data:") {
		return part, nil
	}
	payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if payload == "" {
		return part, nil
	}
	var ev antStreamEvent
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		return part, nil
	}
	switch ev.Type {
	case "message_start":
		a.inputTokens = ev.Message.Usage.InputTokens
		a.cachedTokens = ev.Message.Usage.CacheReadInputTokens
	case "content_block_start":
		// type=="thinking" 仅表示思考块开始，无增量文本，直接忽略
	case "content_block_delta":
		if ev.Delta.Type == "thinking_delta" {
			part.Reasoning = ev.Delta.Thinking
		} else {
			part.Content = ev.Delta.Text
		}
	case "message_delta":
		// 该事件 usage.output_tokens 为累计值
		if ev.Usage.OutputTokens > 0 {
			a.outputTokens = ev.Usage.OutputTokens
		}
	case "message_stop":
		u := adapter.Usage{
			PromptTokens:     a.inputTokens,
			CompletionTokens: a.outputTokens,
			CachedTokens:     a.cachedTokens,
			TotalTokens:      a.inputTokens + a.outputTokens,
		}
		part.Usage = &u
		part.Done = true
	default:
		// 其他事件忽略
	}
	return part, nil
}

// snapshotUsage 供非流式之外的兜底（当前未使用，保留扩展）
func (a *Adapter) snapshotUsage() adapter.Usage {
	return adapter.Usage{
		PromptTokens:     a.inputTokens,
		CompletionTokens: a.outputTokens,
		CachedTokens:     a.cachedTokens,
		TotalTokens:      a.inputTokens + a.outputTokens,
	}
}
