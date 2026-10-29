/**
 * 共享类型定义 —— 严格对齐 docs/api-contract.md 冻结契约
 */

/** 统一响应包 {code, msg, data}，code=0 成功 */
export interface ApiResult<T = unknown> {
  code: number
  msg: string
  data: T
}

/** 分页结构 {list, total, page, page_size} */
export interface PageResult<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

/** 角色枚举 */
export type RoleType = 'admin' | 'user'

/** 协议格式枚举 */
export type ApiFormat = 'openai' | 'anthropic'

/** 用户基础信息（登录返回 / 通用用户简表） */
export interface UserInfo {
  id: number
  username: string
  role: RoleType
  display_name: string
}

/** 当前登录用户 + 权限标识 */
export interface Profile extends UserInfo {
  permissions: string[]
}

/** 登录入参 / 出参 */
export interface LoginParams {
  username: string
  password: string
}
export interface LoginResult {
  token: string
  user: UserInfo
}

/** 修改密码入参 */
export interface PasswordParams {
  old_password: string
  new_password: string
}

/** 注册入参（契约 8.1，公开接口；禁止携带 role/status，后端固定 role=user、status=1） */
export interface RegisterReq {
  username: string
  password: string
  display_name?: string
}

/** 注册出参（契约 8.1：成功 code=0，不签发 token） */
export interface RegisterResult {
  id: number
  username: string
  role: RoleType
  display_name: string
}

/** 供应商 providers（契约 1.1 / 2.2） */
export interface Provider {
  id: number
  name: string
  api_format: ApiFormat
  base_url: string
  platform_url: string
  balance: number
  max_concurrency: number
  timeout_seconds: number
  status: number // 1 启用 / 0 停用
  remark: string
  created_at?: string
  updated_at?: string
}
export interface ProviderForm {
  name: string
  api_format: ApiFormat
  base_url: string
  platform_url?: string
  balance: number
  max_concurrency: number
  timeout_seconds: number
  status: number
  remark?: string
}

/** 模型 models（契约 1.2 / 2.3） */
export interface ModelItem {
  id: number
  name: string
  provider_id: number
  provider_name?: string
  input_price_cached: number // 缓存命中输入价格（元/百万 token）
  input_price: number // 未命中输入价格
  output_price: number // 输出价格
  context_length: number
  support_thinking: boolean
  support_multimodal: boolean
  status: number
  remark?: string
  created_at?: string
}
export interface ModelForm {
  name: string
  provider_id: number | undefined
  input_price_cached: number
  input_price: number
  output_price: number
  context_length: number
  support_thinking: boolean
  support_multimodal: boolean
  status: number
  remark?: string
}

/** API-KEY（契约 1.3 / 2.4）；列表为掩码，详情含完整 api_key */
export interface ApiKeyItem {
  id: number
  name: string
  provider_id: number
  provider_name?: string
  api_key?: string // 仅详情/编辑详情返回完整值；列表为掩码
  expire_date: string | null
  max_quota: number // 0 表示不限
  used_quota: number
  status: number
  remark?: string
  created_at?: string
}
export interface ApiKeyForm {
  name: string
  provider_id: number | undefined
  api_key?: string // 编辑留空表示不修改
  expire_date: string | null
  max_quota: number
  status: number
  remark?: string
}

/** 用户 users（契约 1.4 / 2.5）；不含 password_hash */
export interface UserItem {
  id: number
  username: string
  role: RoleType
  display_name: string
  status: number
  created_at?: string
}
export interface UserCreateForm {
  username: string
  password: string
  role: RoleType
  display_name: string
  status: number
}
export interface UserUpdateForm {
  role: RoleType
  display_name: string
  status: number
}

/** 我的可用 key（契约 2.4 GET /api/me/keys），内嵌供应商与启用模型 */
export interface MyKeyProvider {
  id: number
  name: string
  api_format: ApiFormat
  base_url: string
}
export interface MyKeyModel {
  id: number
  name: string
  context_length: number
}
export interface MyKeyItem {
  id: number
  name: string
  provider: MyKeyProvider
  models: MyKeyModel[]
  status: number
  expire_date: string | null
  max_quota: number
  used_quota: number
  /** 可用性标记；false 时 unusable_reason 说明原因（前端置灰不可选） */
  usable?: boolean
  unusable_reason?: string
}

/** 对话会话 chat_sessions（契约 1.6 / 2.6） */
export interface ChatSession {
  id: number
  user_id: number
  key_id: number
  model: string
  title: string | null
  created_at?: string
}

/** 对话消息 chat_messages（契约 1.7 / 2.6 + 7.2 扩展） */
export interface ChatMessage {
  id?: number
  session_id?: number
  role: 'user' | 'assistant' | 'system'
  content: string
  prompt_tokens: number
  completion_tokens: number
  cached_tokens: number
  total_tokens: number
  cost: number
  created_at?: string
  /** 思考过程文本（契约 7.2 reasoning_content 映射；无思考模型为空串/undefined） */
  reasoning?: string
  /** 模型快照（历史回放由后端返回；实时对话取当前选中模型） */
  model?: string
  /** 供应商名称（历史回放由后端返回；实时对话取当前 Key 所属供应商） */
  provider_name?: string
  /** 前端折叠态（不入库）：历史默认折叠，流式思考阶段展开、完成后自动折叠 */
  reasoningOpen?: boolean
}

/** SSE 事件载荷（契约 2.7） */
export interface SseUsageEvent {
  type: 'usage'
  prompt_tokens: number
  completion_tokens: number
  cached_tokens: number
  total_tokens: number
  cost: number
}

/** 对话转发请求体 */
export interface CompletionPayload {
  key_id: number
  model: string
  session_id?: number | null
  stream: boolean
  messages: Array<{ role: string; content: string }>
}

/** 统计概览（契约 2.8） */
export interface OverviewAdmin {
  provider_count: number
  model_count: number
  key_count: number
  user_count: number
  today_tokens: number
  today_cost: number
  total_cost: number
}
export interface OverviewUser {
  my_key_count: number
  my_session_count: number
  today_tokens: number
  today_cost: number
  total_cost: number
}
export type OverviewResult = OverviewAdmin & Partial<OverviewUser>

/** 6.1 Top 榜单条目（模型/供应商两个维度；calls 调用次数、tokens Token 合计、cost 费用元） */
export interface RankItem {
  name: string
  calls: number
  tokens: number
  cost: number
}

/** 6.2 近 N 天趋势每日聚合 */
export interface TrendItem {
  date: string
  calls: number
  tokens: number
  cost: number
}

/** 6.3 Token 用量构成（固定三项，顺序固定） */
export interface BreakdownItem {
  name: string
  tokens: number
  cost: number
}
