# API-KEY 集中管理工具 —— 前后端接口契约（冻结版 v1）

> 本文件是后端（GoFrame）与前端（Vue3+Electron）共同遵守的唯一契约。字段名、路径、枚举值、响应格式以本文件为准；实现细节（Go 包结构、Vue 文件组织）各自决定，但不得偏离本契约。

## 0. 全局约定

- 后端服务端口默认 `8090`，所有接口前缀 `/api`；开发期前端通过 Vite 代理 `/api -> http://127.0.0.1:8090`。
- 统一响应包：
  ```json
  { "code": 0, "msg": "ok", "data": {} }
  ```
  - `code = 0` 成功；`401` 未登录/token 失效；`403` 无权限或数据不属于当前用户；`400` 参数错误；`500` 服务端错误。
  - 分页响应 `data = { "list": [], "total": 0, "page": 1, "page_size": 10 }`。
- 鉴权：除 `POST /api/auth/login` 外全部接口请求头 `Authorization: Bearer <token>`。
- 角色：`admin`（管理员）、`user`（普通用户）。
- 时间字段统一 ISO8601 字符串（如 `2026-10-29T10:00:00+08:00`）；金额单位：**人民币元**；价格字段单位：**元 / 百万 token**。
- 布尔值 JSON 中用 `true/false`；状态字段统一 `status: 1` 启用 / `0` 停用。
- API-KEY 明文仅用于转发；列表类接口回显掩码 `前4后4`（如 `sk-ab****wxyz`）；详情/编辑详情接口返回完整值；日志禁止打印完整 key。

## 1. 数据库表结构（gdb 模型自动建表，首次启动执行）

### 1.1 providers 供应商
| 字段 | 类型 | 说明 |
|---|---|---|
| id | 自增主键 | |
| name | string 非空 | 供应商名称，唯一 |
| api_format | string | `openai` / `anthropic` |
| base_url | string | 模型 API 地址：OpenAI 格式如 `https://api.openai.com/v1`；Anthropic 格式如 `https://api.anthropic.com`（后端拼接 `/v1/messages`） |
| platform_url | string | API 平台网址（网页后台），可空 |
| balance | float64 默认 0 | 余额（元，人工维护展示） |
| max_concurrency | int 默认 10 | 最大并发数（按供应商维度信号量限流） |
| timeout_seconds | int 默认 60 | 超时时长（秒） |
| status | int 默认 1 | 1 启用 / 0 停用 |
| remark | string | 备注，可空 |
| created_at / updated_at | time | |

### 1.2 models 模型
| 字段 | 类型 | 说明 |
|---|---|---|
| id | 自增主键 | |
| name | string 非空 | 模型名称（如 gpt-4o、claude-3-5-sonnet） |
| provider_id | int 索引 | 关联 providers.id |
| input_price_cached | float64 默认 0 | 缓存命中输入价格（元/百万 token） |
| input_price | float64 默认 0 | 未命中输入价格（元/百万 token） |
| output_price | float64 默认 0 | 输出价格（元/百万 token） |
| context_length | int 默认 0 | 上下文长度 |
| support_thinking | bool 默认 false | 是否支持思考模式 |
| support_multimodal | bool 默认 false | 是否支持多模态 |
| status | int 默认 1 | 1 启用 / 0 停用 |
| remark | string | |
| created_at / updated_at | time | |

### 1.3 api_keys 密钥
| 字段 | 类型 | 说明 |
|---|---|---|
| id | 自增主键 | |
| name | string 非空 | 名称（备注名） |
| provider_id | int 索引 | 关联 providers.id |
| api_key | string 非空 | 完整 key（本地桌面工具落库存储；列表掩码返回） |
| expire_date | string/date 可空 | 有效日期 `YYYY-MM-DD`，空表示永久 |
| max_quota | float64 默认 0 | 最大限额（元），0 表示不限 |
| used_quota | float64 默认 0 | 已用额度（元，随消息费用累计） |
| status | int 默认 1 | 1 启用 / 0 停用 |
| remark | string | |
| created_at / updated_at | time | |

可用性判定（对话前校验，不满足返回 400 并说明原因）：key.status=1、供应商 status=1、未过期、max_quota=0 或 used_quota<max_quota。

### 1.4 users 用户
| 字段 | 类型 | 说明 |
|---|---|---|
| id | 自增主键 | |
| username | string 唯一非空 | 登录账号 |
| password_hash | string | bcrypt 哈希 |
| role | string 默认 `user` | `admin` / `user` |
| display_name | string | 显示名 |
| status | int 默认 1 | 1 启用 / 0 停用 |
| created_at / updated_at | time | |

### 1.5 key_user 密钥-用户分配（多对多）
| id | 自增主键 |
| key_id | int 索引 |
| user_id | int 索引 |
| 唯一约束 (key_id, user_id) |

### 1.6 chat_sessions 会话
| 字段 | 说明 |
|---|---|
| id | 主键 |
| user_id | 所属用户（索引） |
| key_id | 使用的 key |
| model | 模型名称快照字符串 |
| title | 会话标题（首条用户消息前 20 字自动生成，可空） |
| created_at | |

### 1.7 chat_messages 消息
| 字段 | 说明 |
|---|---|
| id | 主键 |
| session_id | 外键索引 |
| role | `user` / `assistant` / `system` |
| content | 文本内容 |
| prompt_tokens | 输入 token |
| completion_tokens | 输出 token |
| cached_tokens | 命中缓存输入 token（可空默认 0） |
| total_tokens | 合计 token |
| cost | 本条消息折算费用（元，仅 assistant 消息计） |
| created_at | |

费用折算：`cost = prompt_tokens/1e6 * input_price(命中缓存用 input_price_cached，否则 input_price) + completion_tokens/1e6 * output_price`；模型停用/价格缺失时按 0 计并落库。消息落库后把 assistant 消息 cost 累加进 `api_keys.used_quota`。

## 2. 接口清单

### 2.1 认证 auth
| 方法 | 路径 | 权限 | 说明 / 入参 / 出参 |
|---|---|---|---|
| POST | /api/auth/login | 公开 | 入 `{username, password}`；出 `{token, user:{id,username,role,display_name}}`；首次内置 admin/admin123 |
| POST | /api/auth/logout | 登录 | 前端清 token 即可，后端可黑名单占位 |
| GET | /api/auth/profile | 登录 | 当前用户信息 + 权限标识 `permissions: string[]`（admin 含 `*`；user 含 `chat:use`） |
| PUT | /api/auth/password | 登录 | 入 `{old_password, new_password}`，本人修改 |

### 2.2 供应商 providers（admin 全部；普通用户不开放管理端）
| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/providers | 分页 + `keyword` + `status` 筛选 |
| POST | /api/providers | 新建，body 为表字段（id 除外） |
| GET | /api/providers/:id | 详情 |
| PUT | /api/providers/:id | 更新 |
| DELETE | /api/providers/:id | 删除前校验：名下存在模型或 key 时返回 400 |
| PUT | /api/providers/:id/status | 入 `{status}` 启停 |

### 2.3 模型 models（admin）
| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/models | 分页 + `provider_id` + `keyword`；列表可带供应商 `provider_name` 冗余字段 |
| POST | /api/models | |
| GET | /api/models/:id | |
| PUT | /api/models/:id | |
| DELETE | /api/models/:id | |
| PUT | /api/models/:id/status | `{status}` |

### 2.4 API-KEY keys
| 方法 | 路径 | 权限 | 说明 |
|---|---|---|---|
| GET | /api/keys | admin 全部；user 不调用管理列表 | 分页 + `provider_id` + `keyword`；返回掩码 key；含 `used_quota/max_quota` |
| POST | /api/keys | admin | |
| GET | /api/keys/:id | admin / 被分配用户 | 返回完整 api_key |
| PUT | /api/keys/:id | admin | |
| DELETE | /api/keys/:id | admin | 级联删除 key_user 分配 |
| PUT | /api/keys/:id/status | admin | `{status}` |
| GET | /api/keys/:id/users | admin | 返回已分配 user_id 数组及用户简表 |
| PUT | /api/keys/:id/users | admin | 入 `{user_ids: number[]}` 全量覆盖分配关系 |
| GET | /api/me/keys | 登录 | **普通用户核心接口**：仅返回「已分配给我 + 全部可用条件满足」的 key；每条含 `provider{id,name,api_format,base_url}`、`models:[{id,name,context_length,...}]`（该供应商下启用模型） |

### 2.5 用户 users（admin）
| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/users | 分页 + `keyword`；不返回 password_hash |
| POST | /api/users | 入 `{username,password,role,display_name,status}` |
| GET | /api/users/:id | |
| PUT | /api/users/:id | 更新 role/display_name/status，不更新密码 |
| PUT | /api/users/:id/status | `{status}` |
| POST | /api/users/:id/reset-password | 入 `{password}`，admin 重置 |
| DELETE | /api/users/:id | 禁止删除自己；删除后清理 key_user |
| GET | /api/users/:id/keys | 该用户已分配 key 列表（用户管理弹窗用） |

### 2.6 对话 chat
| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/chat/sessions | admin 全部；user 仅本人；分页，按创建时间倒序 |
| POST | /api/chat/sessions | 入 `{key_id, model}`；出会话对象；可顺带不传 messages |
| GET | /api/chat/sessions/:id/messages | 消息列表正序；越权返回 403 |
| DELETE | /api/chat/sessions/:id | 仅本人或 admin |
| POST | /api/chat/completions | 对话转发（见 2.7） |

### 2.7 对话转发 POST /api/chat/completions
入参：
```json
{
  "key_id": 1,
  "model": "gpt-4o",
  "session_id": 12,
  "stream": true,
  "messages": [
    {"role": "system", "content": "可选"},
    {"role": "user", "content": "..."}
  ]
}
```
后端流程：
1. 校验 key 可用性（见 1.3）；按 `key.provider.api_format` 选适配器；取供应商 `timeout_seconds` 做 context 超时；按供应商 `max_concurrency` 信号量限流。
2. 组装对应协议请求（见第 3 节）转发。
3. `stream=false`：一次性返回归一化结果；`stream=true`：**SSE**，`Content-Type: text/event-stream`，事件：
   - `data: {"type":"delta","content":"片段"}` —— 文本增量
   - `data: {"type":"usage","prompt_tokens":..,"completion_tokens":..,"cached_tokens":..,"total_tokens":..,"cost":0.0012}` —— 结束前/结束时用量
   - `data: {"type":"done","session_id":12,"message_id":34}` —— 结束
   - `data: {"type":"error","msg":"..."}` —— 上游错误（状态码/错误体透出简化文本）
4. 落库：user 消息与 assistant 消息（含 usage、cost）；累计 key.used_quota；首条消息自动写会话标题。
非流式成功响应 `data`：
```json
{ "content":"...", "usage":{"prompt_tokens":0,"completion_tokens":0,"cached_tokens":0,"total_tokens":0}, "cost":0.0, "message_id":1 }
```

### 2.8 统计 stats
| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/stats/overview | admin：全局 {provider_count,model_count,key_count,user_count, today_tokens,today_cost,total_cost}；user：本人 {my_key_count, my_session_count, today_tokens,today_cost,total_cost} |

## 3. 双协议适配器要求（独立 internal/adapter/openai、internal/adapter/anthropic 包）

统一 Go 抽象（建议）：
```go
type Message struct { Role string; Content string }
type Usage struct { PromptTokens, CompletionTokens, CachedTokens, TotalTokens int }
type Adapter interface {
    BuildRequest(model string, msgs []Message, stream bool) (url string, header http.Header, body any)
    ParseNonStream(body []byte) (content string, usage Usage, err error)
    StreamParse(line string) (delta string, usage *Usage, done bool, err error) // 逐 SSE data 行解析
}
```

### 3.1 OpenAI 兼容
- URL：`{base_url}/chat/completions`（base_url 已含 `/v1`，禁止重复拼接）
- Header：`Authorization: Bearer {api_key}`，`Content-Type: application/json`
- Body：`{model, messages:[{role,content}], stream, max_tokens?}`；system 角色直接置于 messages 数组。
- 非流式：`choices[0].message.content`；`usage.prompt_tokens / completion_tokens / total_tokens`；缓存 token 取 `usage.prompt_tokens_details.cached_tokens`。
- 流式 SSE：`data: {choices:[{delta:{content}}]}`；最后一行 `data: [DONE]`；用量在末尾 chunk 的 `usage`（需请求时带 `stream_options:{include_usage:true}` 由后端自动加）。

### 3.2 Anthropic
- URL：`{base_url}/v1/messages`
- Header：`x-api-key: {api_key}`，`anthropic-version: 2023-06-01`，`Content-Type: application/json`
- Body：`{model, max_tokens: 4096(必填，前端可扩展预留), system: "system 合并为顶层字符串", messages:[{role,content}], stream}`；system 角色消息从 messages 抽出合并到顶层。
- 非流式：`content[0].type=="text"` 的 `text` 拼接；`usage.input_tokens / output_tokens`；缓存取 `usage.cache_read_input_tokens`。
- 流式事件：`message_start`（含 input_tokens 初值）、`content_block_delta.delta.text`、`message_delta.usage.output_tokens`、`message_stop`（done）。

## 4. 前端路由与菜单（角色过滤）

| 路由 | 菜单名 | admin | user |
|---|---|---|---|
| /login | 登录 | - | - |
| /dashboard | 仪表盘 | ✅ 全局统计 | ✅ 个人统计 |
| /providers | 供应商管理 | ✅ | ❌ 菜单隐藏+路由守卫 |
| /models | 模型管理 | ✅ | ❌ |
| /keys | API-KEY 管理 | ✅（含分配用户） | ❌ |
| /users | 用户管理 | ✅ | ❌ |
| /chat | 对话测试 | ✅ | ✅（仅 /api/me/keys 数据） |

对话页交互：左侧会话列表；顶部选择可用 key → 联动该供应商启用模型下拉；聊天区多轮对话、流式渲染；每条 assistant 消息展示 token 与费用；侧边或顶部显示累计消耗。

## 5. 内置初始数据
- 管理员账号：`admin / admin123`（bcrypt 哈希落库；若 users 表为空则插入；非空不覆盖）。
- 2~3 条示例供应商（标注「示例数据，请勿直接用于生产」）、2 条示例模型、1 条示例 key（掩码占位示例值如 `sk-example-xxxx`），仅在首次初始化且全表为空时插入，便于首次进入看到界面形态。

## 6. 第二轮新增统计接口（v2，挂 JWT 登录组；admin 全局口径，user 仅本人口径）

口径统一：仅统计 `chat_messages.role='assistant'` 的消息；user 口径必须 `chat_messages m JOIN chat_sessions s ON s.id=m.session_id WHERE s.user_id=? AND m.role='assistant'`；admin 口径不加用户过滤。空表 / 无数据返回空数组 `[]`，不得报错。SQLite 与 MySQL 双方言兼容（日期分组仿现有 `todayCondition` 方言分支：SQLite `date(col,'localtime')`，MySQL `DATE(col)`）。

### 6.1 GET /api/stats/rankings?dim=model|provider&metric=calls|tokens|cost&limit=10
- `dim=model`：按 `s.model` 分组；`calls`=assistant 消息条数，`tokens`=`SUM(m.prompt_tokens + m.completion_tokens)`，`cost`=`SUM(m.cost)`。
- `dim=provider`：`JOIN api_keys k ON k.id=s.key_id JOIN providers p ON p.id=k.provider_id`，按 `p.name` 分组（供应商缺失归为「未知供应商」）。
- `metric` 缺省 `calls`；`limit` 缺省 10、上限 50；结果按所选 metric 降序。
- 响应 `data`：
```json
[ { "name": "gpt-4o", "calls": 12, "tokens": 34567, "cost": 0.2345 } ]
```

### 6.2 GET /api/stats/trend?days=30
- 近 N 个自然日每日聚合（days 缺省 30、取值范围 1~90）；无数据日期补 0，保证折线连续；数组按日期升序。
- 响应 `data`：
```json
[ { "date": "2026-10-01", "calls": 5, "tokens": 12345, "cost": 0.34 } ]
```

### 6.3 GET /api/stats/breakdown
- 三类 Token 构成：①缓存命中输入 `cached_tokens`；②未命中输入 `GREATEST(prompt_tokens - cached_tokens, 0)`；③输出 `completion_tokens`。
- `tokens` 为对应 Token 合计；`cost` 口径为 Token 构成展示（可返回 0，代码注释写明「费用精确拆分需按模型单价二次折算，本接口以 Token 构成为准」）。
- 响应 `data`（固定三项，顺序固定）：
```json
[
  { "name": "缓存命中输入", "tokens": 1000, "cost": 0 },
  { "name": "未命中输入",   "tokens": 2000, "cost": 0 },
  { "name": "输出",         "tokens": 800,  "cost": 0 }
]
```

### 6.4 前端类型与调用约定
- `RankItem { name: string; calls: number; tokens: number; cost: number }`
- `TrendItem { date: string; calls: number; tokens: number; cost: number }`
- `BreakdownItem { name: string; tokens: number; cost: number }`
- 三个接口前端分别封装为 `getRankings(params)`、`getTrend(days)`、`getBreakdown()`；admin/user 均调用，数据范围由后端按角色收敛。

## 7. 第三轮修复补充（v3）

### 7.1 Key 创建即分配
- `POST /api/keys` 创建成功后，在同一事务内把创建者（当前登录用户 `user_id`）幂等写入 `key_user(key_id, user_id)`（先 Count 存在性再插入，避免重复）。
- `GET /api/me/keys` 仍只按 `key_user` 关联过滤，**不增加 admin 全量特判**；管理员创建的 Key 因自动分配而自然出现在本人 `/me/keys`。
- 前端「分配用户」批量覆盖弹窗行为保持不变；创建 Key 成功后前端无需再调用分配接口。

### 7.2 消息结构扩展（chat_messages 表 + 历史消息接口）
- 新列 `reasoning_content TEXT NOT NULL DEFAULT ''`（思考过程文本，无思考模型为空串）。
  - 新库：`initialize/db.go` 建表 DDL 直接包含该列。
  - 旧库：建表后执行幂等列迁移——SQLite 查 `sqlite_master`/`PRAGMA table_info(chat_messages)` 判断列存在性，MySQL 查 `information_schema.COLUMNS`；缺失才 `ALTER TABLE chat_messages ADD COLUMN reasoning_content TEXT NOT NULL DEFAULT ''`，存在则跳过，双方言注释清楚。
- `GET /api/chat/sessions/:id/messages` 每条消息在原字段基础上补充：
  - `reasoning_content: string`（落库的思考过程全文）
  - `model: string`（取自会话快照 `chat_sessions.model`）
  - `provider_name: string`（联 `api_keys k → providers p` 取 `p.name`，取不到为空串）

### 7.3 思考内容（reasoning）协议解析与 SSE 事件
- 适配器抽象建议重构为：
  ```go
  type StreamPart struct {
      Content   string   // 正文增量
      Reasoning string   // 思考增量
      Usage     *Usage
      Done      bool
  }
  ParseNonStream(body []byte) (content, reasoning string, usage Usage, err error)
  StreamParse(line string) (*StreamPart, error)
  ```
- OpenAI 兼容：SSE 取 `choices[0].delta.reasoning_content`（DeepSeek 类），兜底别名 `delta.reasoning`；非流式取 `message.reasoning_content`。
- Anthropic：SSE `content_block_delta.delta.type=="thinking_delta"` 取 `delta.thinking`；`content_block_start.type=="thinking"` 仅作思考块开始标记；非流式 `content[]` 中 `type=="thinking"` 项取 `thinking` 文本。
- SSE 新增事件（与 delta 并行、独立累加，无思考模型时不发该事件）：
  ```
  data: {"type":"reasoning","content":"思考增量片段"}
  ```
- 非流式成功响应 `data` 增加 `reasoning: string`。
- **非 SSE 降级**：`stream=true` 但上游响应 `Content-Type` 非 `text/event-stream`（上游不支持流式）时，后端按一次性 JSON 解析，依次补发 `delta`（全文）、`usage`、`done` 事件，不得挂起或报错卡死；`streaming` 状态由前端 finally 统一复位。

### 7.4 前端类型与事件约定
- `ChatMessage` 增加 `reasoning?: string`、`model?: string`、`provider_name?: string`；历史消息接口把 `reasoning_content` 映射到 `reasoning`。
- `streamCompletions` handlers 增加 `onReasoning(content: string)`；UI 在 assistant 气泡正文上方渲染「思考过程」可折叠区块：无内容不渲染；流式思考阶段展开、结束后自动折叠；历史回放带出已保存 reasoning，默认折叠。
- assistant 气泡头部小字标签展示 `provider_name · model`；user 消息不显示该头部标签。

## 8. 第四轮补充（v4）

### 8.1 注册接口（公开，无需 JWT，与 `/api/auth/login` 同处公开组）
- `POST /api/auth/register`
- 请求体：
  ```json
  { "username": "alice", "password": "至少6位", "display_name": "可选昵称" }
  ```
  - `username` 必填，3~32 位字母 / 数字 / 下划线；`password` 必填且至少 6 位；`display_name` 可选；**禁止接收 role / status**，后端固定 `role="user"`、`status=1`。
- 成功 `code=0`，`data = { "id": 12, "username": "alice", "role": "user", "display_name": "..." }`；**不签发 token**，前端提示后切回登录表单并回填用户名。
- 错误码：参数校验失败 `code=400`（msg 指明字段）；用户名已存在 `code=400`，msg「用户名已存在」；开关关闭 `code=400`，msg「注册已关闭」。
- 密码 bcrypt 哈希存储（复用现有 `golang.org/x/crypto/bcrypt`）；响应不回传密码哈希。
- 配置 `app.allow_register`（bool，默认 `true`；配置缺失按 true 兜底；`false` 时注册接口返回「注册已关闭」）。

### 8.2 流式输出行为约定（正式 SSE + 非 SSE 降级）
- 正式 SSE：上游逐行 `StreamParse` → 立即 `sseWrite` + `Flush`；事件类型保持 `delta / reasoning / usage / done / error`；响应头 `X-Accel-Buffering: no` 防代理缓冲；请求体 `stream` 字段在 OpenAI / Anthropic 适配器均透传，不被默认值覆盖。
- 非 SSE 降级（上游 Content-Type 非 `text/event-stream`）：后端把 reasoning 与正文切片后逐片下发——先 reasoning 分片、再正文分片，片间 12~30ms 间隔；切片不得破坏代码块 / 列表项等 Markdown 结构单元；顺序 `reasoning* → delta* → usage → done`；整体受供应商超时 context 约束，context 取消立即退出。
- 前端：`onDelta / onReasoning` 纯增量累加；`onDone` 仅折叠思考区与刷新会话列表，**不得整体重置 content / reasoning**；reader 兼容 `\r\n` 行尾；残行兜底解析保留。

### 8.3 消息渲染与布局约定
- assistant 正文：`marked` 解析 + `DOMPurify.sanitize` 清洗后渲染（禁止未过滤 v-html）；`utils/markdown.ts` 导出 `renderMarkdown(text): string`；空串不渲染、解析异常兜底为转义纯文本；历史回放复用同一函数。
- user 消息与 reasoning 折叠区保持纯文本 `pre-wrap`，不做 Markdown 渲染。
- 布局：assistant 气泡占满消息区可用宽度（白底圆角卡片）；user 气泡靠右、约 60% 宽、蓝底白字。
- Markdown 样式：标题层级、段落行高 1.7、列表缩进、引用块、行内 code、代码块深色 `#0f172a` 等宽 + 横向滚动、表格细线斑马纹、链接主题色、图片最大宽 100%；本轮不引入 highlight.js 语法高亮（仅代码块基础样式）。
