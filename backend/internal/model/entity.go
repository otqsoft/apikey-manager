// Package model 数据库实体结构（与 docs/api-contract.md 第 1 节表结构一一对应）。
package model

// Provider 供应商（providers 表）
type Provider struct {
	ID             uint    `orm:"id"              json:"id"`
	OwnerID        uint    `orm:"owner_id"        json:"owner_id"`         // 归属用户：创建者；admin 可见全部
	Name           string  `orm:"name"            json:"name"`
	ApiFormat      string  `orm:"api_format"      json:"api_format"`       // openai / anthropic
	BaseURL        string  `orm:"base_url"        json:"base_url"`         // OpenAI 格式已含 /v1；Anthropic 不含
	PlatformURL    string  `orm:"platform_url"    json:"platform_url"`      // 网页后台地址，可空
	Balance        float64 `orm:"balance"         json:"balance"`          // 余额（元，人工维护展示）
	MaxConcurrency int     `orm:"max_concurrency" json:"max_concurrency"`  // 供应商维度最大并发信号量
	TimeoutSeconds int     `orm:"timeout_seconds" json:"timeout_seconds"`  // 上游请求超时（秒）
	Status         int     `orm:"status"         json:"status"`           // 1 启用 / 0 停用
	Remark         string  `orm:"remark"         json:"remark"`
	CreatedAt      string  `orm:"created_at"    json:"created_at"`
	UpdatedAt      string  `orm:"updated_at"    json:"updated_at"`
}

// Model 模型（models 表）
type Model struct {
	ID                 uint   `orm:"id"                   json:"id"`
	OwnerID            uint   `orm:"owner_id"             json:"owner_id"`    // 归属用户：创建者
	Name               string `orm:"name"                 json:"name"`
	ProviderID         uint   `orm:"provider_id"          json:"provider_id"`
	InputPriceCached   float64 `orm:"input_price_cached"  json:"input_price_cached"` // 缓存命中输入价（元/百万 token）
	InputPrice         float64 `orm:"input_price"         json:"input_price"`        // 未命中输入价
	OutputPrice        float64 `orm:"output_price"        json:"output_price"`       // 输出价
	ContextLength      int     `orm:"context_length"      json:"context_length"`
	SupportThinking    bool    `orm:"support_thinking"    json:"support_thinking"`
	SupportMultimodal  bool    `orm:"support_multimodal"  json:"support_multimodal"`
	Status             int     `orm:"status"              json:"status"`
	Remark             string  `orm:"remark"              json:"remark"`
	ProviderName       string  `orm:"-"                   json:"provider_name,omitempty"` // 列表冗余字段
	CreatedAt          string  `orm:"created_at"         json:"created_at"`
	UpdatedAt          string  `orm:"updated_at"         json:"updated_at"`
}

// ApiKey 密钥（api_keys 表）
type ApiKey struct {
	ID          uint    `orm:"id"          json:"id"`
	OwnerID     uint    `orm:"owner_id"    json:"owner_id"`     // 归属用户：创建者；owner 或被分配者可用
	Name        string  `orm:"name"        json:"name"`
	ProviderID  uint    `orm:"provider_id" json:"provider_id"`
	APIKey      string  `orm:"api_key"     json:"api_key"`       // 完整 key：仅详情/转发使用，列表必须掩码
	ExpireDate  string  `orm:"expire_date" json:"expire_date"`  // YYYY-MM-DD，空表示永久
	MaxQuota    float64 `orm:"max_quota"   json:"max_quota"`    // 最大限额（元），0 不限
	UsedQuota   float64 `orm:"used_quota"  json:"used_quota"`   // 已用额度（元）
	Status      int     `orm:"status"      json:"status"`
	Remark      string  `orm:"remark"      json:"remark"`
	CreatedAt   string  `orm:"created_at" json:"created_at"`
	UpdatedAt   string  `orm:"updated_at" json:"updated_at"`
}

// User 用户（users 表）
type User struct {
	ID           uint   `orm:"id"            json:"id"`
	Username     string `orm:"username"      json:"username"`
	PasswordHash string `orm:"password_hash" json:"-"` // 永不对外输出
	Role         string `orm:"role"          json:"role"` // admin / user
	DisplayName  string `orm:"display_name"  json:"display_name"`
	Status       int    `orm:"status"        json:"status"`
	CreatedAt    string `orm:"created_at"   json:"created_at"`
	UpdatedAt    string `orm:"updated_at"   json:"updated_at"`
}

// KeyUser 密钥-用户分配关系（key_user 表）
type KeyUser struct {
	ID     uint `orm:"id"      json:"id"`
	KeyID  uint `orm:"key_id"  json:"key_id"`
	UserID uint `orm:"user_id" json:"user_id"`
}

// ChatSession 对话会话（chat_sessions 表）
type ChatSession struct {
	ID        uint   `orm:"id"         json:"id"`
	UserID    uint   `orm:"user_id"    json:"user_id"`
	KeyID     uint   `orm:"key_id"     json:"key_id"`
	Model     string `orm:"model"      json:"model"`
	Title     string `orm:"title"      json:"title"`
	CreatedAt string `orm:"created_at" json:"created_at"`
}

// ChatMessage 对话消息（chat_messages 表）
type ChatMessage struct {
	ID               uint    `orm:"id"                json:"id"`
	SessionID        uint    `orm:"session_id"        json:"session_id"`
	Role             string  `orm:"role"              json:"role"` // user / assistant / system
	Content          string  `orm:"content"          json:"content"`
	ReasoningContent string  `orm:"reasoning_content" json:"reasoning_content"` // 思考过程全文，无思考模型为空串
	PromptTokens    int     `orm:"prompt_tokens"    json:"prompt_tokens"`
	CompletionTokens int     `orm:"completion_tokens" json:"completion_tokens"`
	CachedTokens    int     `orm:"cached_tokens"    json:"cached_tokens"`
	TotalTokens     int     `orm:"total_tokens"      json:"total_tokens"`
	Cost             float64 `orm:"cost"              json:"cost"` // 仅 assistant 消息计费
	CreatedAt        string  `orm:"created_at"        json:"created_at"`
	// 以下为历史消息接口的冗余展示字段（契约 7.2，非表列）
	Model        string `orm:"-" json:"model"`            // 取自会话快照 chat_sessions.model
	ProviderName string `orm:"-" json:"provider_name"`     // 联 api_keys→providers 取供应商名，取不到为空串
}
