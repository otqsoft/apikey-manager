// Package adapter 大模型上游协议抽象。
//
// 两套协议关键差异（实现注释同样记录在各子包）：
//  1. system 字段位置：OpenAI 把 system 直接放在 messages 数组；Anthropic 要求把 system
//     抽出合并为顶层字符串字段 system。
//  2. max_tokens：Anthropic 必填（默认 4096）；OpenAI 可选。
//  3. 鉴权头：OpenAI 用 Authorization: Bearer；Anthropic 用 x-api-key + anthropic-version。
//  4. usage 字段名：OpenAI prompt_tokens/completion_tokens，缓存 token 在
//     usage.prompt_tokens_details.cached_tokens；Anthropic 为 input_tokens/output_tokens，
//     缓存命中为 usage.cache_read_input_tokens。
//  5. SSE 结构：OpenAI delta 在 choices[0].delta.content，结束标记 data: [DONE]；
//     Anthropic 为 message_start / content_block_delta / message_delta / message_stop 事件序列。
package adapter

import "net/http"

// Message 归一化消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage 归一化 token 用量
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	CachedTokens     int
	TotalTokens      int
}

// StreamPart 单条 SSE 事件归一化结果（契约 7.3）。
// Content 为正文增量；Reasoning 为思考过程增量；Usage 为本次解析到的用量（可为 nil）；Done 表示流结束。
type StreamPart struct {
	Content   string
	Reasoning string
	Usage     *Usage
	Done      bool
}

// Adapter 上游协议适配器抽象
type Adapter interface {
	// BuildRequest 组装请求：返回完整 URL、请求头、请求体
	BuildRequest(model string, msgs []Message, stream bool) (url string, header http.Header, body interface{})
	// ParseNonStream 解析非流式一次性响应；返回正文 content、思考过程 reasoning、用量 usage
	ParseNonStream(body []byte) (content, reasoning string, usage Usage, err error)
	// StreamParse 逐行解析上游 SSE；line 为原始文本行（含 "data: " 前缀），返回归一化片段
	StreamParse(line string) (*StreamPart, error)
}
