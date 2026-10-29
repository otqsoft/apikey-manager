package model

// ===================== 认证 =====================

// LoginReq 登录入参
type LoginReq struct {
	Username string `json:"username" v:"required#账号不能为空"`
	Password string `json:"password" v:"required#密码不能为空"`
}

// LoginResp 登录出参
type LoginResp struct {
	Token string `json:"token"`
	User  struct {
		ID          uint   `json:"id"`
		Username    string `json:"username"`
		Role        string `json:"role"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
}

// ProfileResp 当前用户信息 + 权限标识
type ProfileResp struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	DisplayName string   `json:"display_name"`
	Status      int      `json:"status"`
	Permissions []string `json:"permissions"`
}

// ChangePasswordReq 修改本人密码
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" v:"required#旧密码不能为空"`
	NewPassword string `json:"new_password" v:"required|min-length:6#新密码不能为空|新密码长度至少6位"`
}

// RegisterReq 公开注册入参（契约 8.1；不提供 role/status 字段，后端固定 user/启用）
type RegisterReq struct {
	Username    string `json:"username" v:"required|regex:^[A-Za-z0-9_]{3,32}$#用户名不能为空|用户名须为3~32位字母、数字或下划线"`
	Password    string `json:"password" v:"required|min-length:6#密码不能为空|密码至少6位"`
	DisplayName string `json:"display_name"`
}

// RegisterResp 注册成功出参（不签发 token、不回传密码哈希）
type RegisterResp struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	DisplayName string `json:"display_name"`
}

// ===================== 供应商 =====================

// ProviderSaveReq 新建/更新供应商入参
type ProviderSaveReq struct {
	Name           string  `json:"name"          v:"required#供应商名称不能为空"`
	ApiFormat      string  `json:"api_format"    v:"required|in:openai,anthropic#协议格式非法"`
	BaseURL        string  `json:"base_url"      v:"required#接口地址不能为空"`
	PlatformURL    string  `json:"platform_url"`
	Balance        float64 `json:"balance"`
	MaxConcurrency int     `json:"max_concurrency"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	Status         int     `json:"status"`
	Remark         string  `json:"remark"`
}

// StatusReq 通用启停入参
type StatusReq struct {
	Status int `json:"status" v:"required|in:0,1#状态值非法（仅 0/1）"`
}

// ===================== 模型 =====================

// ModelSaveReq 新建/更新模型入参
type ModelSaveReq struct {
	Name              string  `json:"name"                v:"required#模型名称不能为空"`
	ProviderID        uint    `json:"provider_id"        v:"required#所属供应商不能为空"`
	InputPriceCached float64 `json:"input_price_cached"`
	InputPrice       float64 `json:"input_price"`
	OutputPrice       float64 `json:"output_price"`
	ContextLength    int     `json:"context_length"`
	SupportThinking   bool    `json:"support_thinking"`
	SupportMultimodal bool   `json:"support_multimodal"`
	Status            int     `json:"status"`
	Remark            string  `json:"remark"`
}

// ===================== 密钥 =====================

// KeySaveReq 新建/更新密钥入参
type KeySaveReq struct {
	Name       string  `json:"name"        v:"required#密钥名称不能为空"`
	ProviderID uint    `json:"provider_id" v:"required#所属供应商不能为空"`
	APIKey     string  `json:"api_key"`
	ExpireDate string  `json:"expire_date"`
	MaxQuota   float64 `json:"max_quota"`
	Status     int     `json:"status"`
	Remark     string  `json:"remark"`
}

// KeyUsersReq 全量覆盖分配用户
type KeyUsersReq struct {
	UserIDs []uint `json:"user_ids"`
}

// MeProvider 个人可用密钥返回中的供应商快照
type MeProvider struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ApiFormat string `json:"api_format"`
	BaseURL   string `json:"base_url"`
}

// MeModel 个人可用密钥返回中的启用模型快照
type MeModel struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	ContextLength     int    `json:"context_length"`
	SupportThinking   bool   `json:"support_thinking"`
	SupportMultimodal bool   `json:"support_multimodal"`
}

// MeKeyItem 普通用户 GET /api/me/keys 返回项（全量返回已分配密钥，不可用项前端置灰）
type MeKeyItem struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	ProviderID     uint       `json:"provider_id"`
	KeyMasked      string     `json:"key_masked"`
	ExpireDate     string     `json:"expire_date"`
	MaxQuota       float64    `json:"max_quota"`
	UsedQuota      float64    `json:"used_quota"`
	Status         int        `json:"status"`
	Remark         string     `json:"remark"`
	Provider       MeProvider `json:"provider"`
	Models         []MeModel  `json:"models"`
	Usable         bool       `json:"usable"`
	UnusableReason string     `json:"unusable_reason"`
}

// ===================== 用户管理 =====================

// UserSaveReq 新建用户入参
type UserSaveReq struct {
	Username    string `json:"username"   v:"required#账号不能为空"`
	Password    string `json:"password"   v:"required|min-length:6#初始密码不能为空且至少6位"`
	Role        string `json:"role"      v:"in:admin,user#角色非法"`
	DisplayName string `json:"display_name"`
	Status      int    `json:"status"`
}

// UserUpdateReq 更新用户（不含密码、不含账号）
type UserUpdateReq struct {
	Role        string `json:"role" v:"in:admin,user#角色非法"`
	DisplayName string `json:"display_name"`
	Status      int    `json:"status"`
}

// ResetPasswordReq 管理员重置密码
type ResetPasswordReq struct {
	Password string `json:"password" v:"required|min-length:6#新密码不能为空且至少6位"`
}

// ===================== 对话 =====================

// ChatMessageItem 转发请求中的单条消息
type ChatMessageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CreateSessionReq 创建会话
type CreateSessionReq struct {
	KeyID uint   `json:"key_id" v:"required#缺少 key_id"`
	Model string `json:"model" v:"required#缺少 model"`
}

// ChatCompletionReq 对话转发入参
type ChatCompletionReq struct {
	SessionID uint              `json:"session_id"`
	KeyID     uint              `json:"key_id" v:"required#缺少 key_id"`
	Model     string            `json:"model" v:"required#缺少 model"`
	Stream    bool              `json:"stream"`
	Messages  []ChatMessageItem `json:"messages" v:"required#消息列表不能为空"`
}

// ChatUsage token 用量归一化结构
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	CachedTokens     int `json:"cached_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatNonStreamResp 非流式成功响应（契约 7.3：reasoning 为思考过程，无思考模型为空串）
type ChatNonStreamResp struct {
	Content   string    `json:"content"`
	Reasoning string    `json:"reasoning"`
	Usage     ChatUsage `json:"usage"`
	Cost      float64   `json:"cost"`
	MessageID uint      `json:"message_id"`
}

// PageResult 统一分页结构
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
