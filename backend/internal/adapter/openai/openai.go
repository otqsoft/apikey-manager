// Package openai OpenAI 兼容协议适配器。
//
// 协议要点（契约 3.1）：
//   - URL：{base_url}/chat/completions；base_url 已含 /v1，禁止重复拼接 /v1。
//   - Header：Authorization: Bearer {api_key}，Content-Type: application/json。
//   - Body：{model, messages:[{role,content}], stream}；system 消息直接置于 messages 数组；
//     流式时后端自动附加 stream_options.include_usage=true 以获取末尾用量。
//   - 非流式：choices[0].message.content；usage.prompt_tokens / completion_tokens / total_tokens；
//     缓存 token 取 usage.prompt_tokens_details.cached_tokens。
//   - 流式：data: {choices:[{delta:{content}}]}；最后一行为 data: [DONE]；用量在末尾 chunk 的 usage。
package openai

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"apikey-manager/internal/adapter"
)

// Adapter OpenAI 兼容适配器
type Adapter struct {
	BaseURL string
	APIKey  string
}

// New 构造适配器
func New(baseURL, apiKey string) *Adapter {
	return &Adapter{BaseURL: baseURL, APIKey: apiKey}
}

// BuildRequest 组装 OpenAI 兼容请求
func (a *Adapter) BuildRequest(model string, msgs []adapter.Message, stream bool) (string, http.Header, interface{}) {
	// base_url 约定已含 /v1，这里只追加 /chat/completions，避免出现 /v1/v1
	url := strings.TrimRight(a.BaseURL, "/") + "/chat/completions"
	header := http.Header{}
	header.Set("Authorization", "Bearer "+a.APIKey)
	header.Set("Content-Type", "application/json")

	body := g.Map{
		"model":    model,
		"messages": msgs, // system 角色直接留在数组内
		"stream":   stream,
	}
	if stream {
		// 让上游在流式末尾回传 usage 用量
		body["stream_options"] = g.Map{"include_usage": true}
	}
	return url, header, body
}

// ------- 响应结构 -------

type oaiUsageDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type oaiUsage struct {
	PromptTokens     int              `json:"prompt_tokens"`
	CompletionTokens int              `json:"completion_tokens"`
	TotalTokens      int              `json:"total_tokens"`
	PromptTokensDetails oaiUsageDetails `json:"prompt_tokens_details"`
}

type oaiChoiceMessage struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"` // DeepSeek 等思考模型非流式思考文本
}

type oaiChoiceDelta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"` // 主流思考字段
	Reasoning        string `json:"reasoning"`         // 部分兼容实现的兜底别名
}

type oaiChoice struct {
	Message      oaiChoiceMessage `json:"message"`
	Delta        oaiChoiceDelta   `json:"delta"`
	FinishReason string           `json:"finish_reason"`
}

type oaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type oaiResp struct {
	ID      string     `json:"id"`
	Choices []oaiChoice `json:"choices"`
	Usage   oaiUsage   `json:"usage"`
	Error   *oaiError  `json:"error"`
}

// ParseNonStream 解析非流式响应；content 为正文，reasoning 为思考过程（无思考模型为空串）
func (a *Adapter) ParseNonStream(body []byte) (content, reasoning string, usage adapter.Usage, err error) {
	var resp oaiResp
	if err = json.Unmarshal(body, &resp); err != nil {
		return "", "", adapter.Usage{}, err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return "", "", adapter.Usage{}, nil
	}
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
		reasoning = resp.Choices[0].Message.ReasoningContent
	}
	usage = adapter.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		CachedTokens:     resp.Usage.PromptTokensDetails.CachedTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	return content, reasoning, usage, nil
}

// stripData 去掉 SSE "data: " 前缀；空行/注释行返回空串
func stripData(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, ":") {
		return ""
	}
	payload := strings.TrimPrefix(line, "data:")
	return strings.TrimSpace(payload)
}

// StreamParse 逐行解析 OpenAI SSE；正文 delta 与思考 reasoning 分别通过 StreamPart 透出
func (a *Adapter) StreamParse(line string) (*adapter.StreamPart, error) {
	part := &adapter.StreamPart{}
	payload := stripData(line)
	if payload == "" {
		return part, nil
	}
	if payload == "[DONE]" {
		part.Done = true
		return part, nil
	}
	var chunk struct {
		Choices []struct {
			Delta oaiChoiceDelta `json:"delta"`
		} `json:"choices"`
		Usage *oaiUsage `json:"usage"`
	}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		return part, nil // 忽略无法解析的心跳/注释行
	}
	if len(chunk.Choices) > 0 {
		d := chunk.Choices[0].Delta
		part.Content = d.Content
		// 优先 reasoning_content，兼容兜底字段 reasoning
		part.Reasoning = d.ReasoningContent
		if part.Reasoning == "" {
			part.Reasoning = d.Reasoning
		}
	}
	if chunk.Usage != nil {
		u := adapter.Usage{
			PromptTokens:     chunk.Usage.PromptTokens,
			CompletionTokens: chunk.Usage.CompletionTokens,
			CachedTokens:     chunk.Usage.PromptTokensDetails.CachedTokens,
			TotalTokens:      chunk.Usage.TotalTokens,
		}
		if u.TotalTokens == 0 {
			u.TotalTokens = u.PromptTokens + u.CompletionTokens
		}
		part.Usage = &u
	}
	return part, nil
}
