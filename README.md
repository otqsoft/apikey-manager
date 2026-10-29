# 多平台大模型 API-KEY 集中管理工具

一个面向个人 / 小团队的桌面端「模型 API-KEY 统一管理与对话测试台」：集中维护多个模型供应商（OpenAI 兼容 / Anthropic 两种协议）、模型定价、API-KEY 额度与分配关系，内置对话测试页并自动统计 Token 消耗与费用。后端持有密钥，前端不接触第三方 Key；普通用户只能使用被分配给自己的 Key。

- 后端：Go + GoFrame v2 + gdb（默认 SQLite 纯 Go 驱动，可切换 MySQL）
- 前端：Vue3 + Vite + TypeScript + Element Plus + Pinia + Vue Router + axios
- 桌面壳：Electron（electron-builder 打包 Windows nsis / macOS dmg / Linux AppImage）

## 功能特性

| 模块 | 说明 |
|---|---|
| 供应商管理 | 名称、协议格式（OpenAI / Anthropic）、API 地址 base_url、平台网址、余额、最大并发数、超时时长、启停、备注 |
| 模型管理 | 所属供应商、缓存命中输入价、未命中输入价、输出价（元/百万 Token，两位小数）、上下文长度（列表按 M 展示）、是否支持思考模式、是否支持多模态、启停 |
| API-KEY 管理 | 名称、所属供应商、Key 明文（列表掩码回显）、有效日期、最大限额 / 已用额度进度、启停；一个供应商可挂多个 Key；可批量分配给用户 |
| 用户与权限 | admin 管理员（全部管理 + 全局统计 + 查看全部会话）；user 普通用户（仅使用被分配且可用的 Key、仅看本人会话与个人统计）；JWT 登录；公开自助注册默认开通（注册即普通用户，`app.allow_register` 可关闭） |
| 对话测试 | 按用户可用 Key 与模型发起多轮对话；支持 SSE 流式渲染；记录每条消息 Token（含缓存命中 Token）与折算费用；Key 额度自动累计 |
| 对话增强 | 思考过程（reasoning）独立 SSE 事件、气泡内可折叠展示；assistant 消息头部显示「供应商 · 模型」；新建 Key 自动分配给创建者；上游非 SSE 时后端切片模拟逐字流式（约 20ms 片间隔、Markdown 结构单元不拆碎、受供应商超时约束） |
| 消息渲染 | assistant 正文 Markdown 渲染（marked 解析 + DOMPurify 消毒，杜绝未过滤注入；代码块暗色等宽横向滚动；本轮不引入语法高亮依赖）；assistant 气泡占满消息区宽度，user 气泡靠右约 60% 蓝底；user 消息与思考折叠区保持纯文本 |
| 统计概览 | 全局 / 个人维度的供应商数、模型数、Key 数、会话数、今日 Token、今日费用、累计费用 |
| 仪表盘图表 | 调用 Top 榜单（模型 / 供应商 × 调用次数 / Token / 费用切换）、供应商调用占比环形图、7/30 天用量趋势双轴图（柱=费用、线=Token）、Token 用量构成环形图；空数据空态兜底；admin 全局、user 本人口径 |
| 双协议转发 | OpenAI 兼容 `POST {base_url}/chat/completions`（Bearer）与 Anthropic `POST {base_url}/v1/messages`（x-api-key + anthropic-version），协议适配独立成包 |

## 目录结构

```
.
├── backend/                    # GoFrame v2 后端（模块名 apikey-manager）
│   ├── main.go                 # 入口；空白导入 sqlite 驱动，mysql 驱动注释预留
│   ├── manifest/config/config.yaml
│   └── internal/
│       ├── cmd/                # 路由装配与启动命令
│       ├── controller/         # 接口控制器（auth/provider/model/key/user/chat/stats）
│       ├── service/            # 业务接口定义
│       ├── logic/              # 业务实现（含 scan 行扫描辅助）
│       ├── model/               # 实体实体 entity.go + DTO dto.go
│       ├── middleware/         # JWT 鉴权 + 管理员守卫 + 上下文用户
│       ├── adapter/            # 协议公共抽象；openai/ 与 anthropic/ 两套适配
│       ├── initialize/          # SQLite/MySQL 双方言 DDL 自动建表 + 种子数据
│       └── pkg/                # errs / response / jwtx / semaphore / util(掩码、截断)
├── web/                        # Vue3 + Electron 前端
│   ├── electron/               # 主进程 main.cjs + preload.cjs
│   ├── resources/             # 打包时 extraResources 后端二进制预留位
│   ├── src/
│   │   ├── api/                # 按模块拆分的接口封装（含类型）
│   │   ├── types/index.ts      # TS 类型（逐字段对齐后端契约）
│   │   ├── utils/request.ts    # axios 封装（Bearer 注入、统一解包与错误提示）
│   │   ├── store/              # Pinia 用户态（localStorage 持久化）
│   │   ├── router/             # 路由表 + 角色守卫
│   │   ├── layout/             # 侧边菜单 + 顶栏 + 修改密码弹窗
│   │   └── views/              # login/dashboard/providers/models/keys/users/chat
│   ├── vite.config.ts          # 5173 端口；/api 代理到 127.0.0.1:8090；outDir dist
│   └── package.json            # electron-builder 配置（win nsis / mac dmg / linux AppImage）
└── docs/api-contract.md         # 前后端冻结接口契约（表结构 + 端点 + SSE 格式）
```

## 环境要求

- Go 1.21+（开发机已在 Go 1.23 验证；SQLite 为纯 Go 驱动，**无需 CGO / gcc**）
- Node.js 18+（推荐 20/22）、npm
- 可选：MySQL 8.0+（默认使用内置 SQLite，无需安装数据库服务）

## 快速开始

> 开发期为「后端服务 + 前端 dev server」分别启动；Electron 仅在需要桌面窗口调试 / 打包时使用。

### 1. 启动后端

```bash
cd backend
go mod tidy
go run . 
# 默认监听 :8090；首次启动自动建表并写入内置管理员与示例数据
```

配置文件 `backend/manifest/config/config.yaml`：

| 配置项 | 说明 |
|---|---|
| server.address | 服务监听地址，默认 `:8090` |
| database.default.type | 驱动名：`sqlite`（默认，纯 Go 无 CGO）或 `mysql`（GoFrame gdb 配置键为 type） |
| database.default.name | SQLite 为数据库文件路径（默认 `./data/apikey.db`，启动时自动创建目录与表）；MySQL 为库名 |
| mysql 注释节 | host/port/name/user/pass/charset；切 MySQL 时取消注释、注释 sqlite 节，并在 `main.go` 放开 mysql 驱动空白导入 |
| jwt.secret / jwt.expire | JWT 签名密钥与有效期（秒，默认 7 天；生产请修改 secret） |
| default.username / password / displayName | 首次初始化内置管理员（admin / admin123），仅 users 表为空时写入，不覆盖已有账号 |

> 后端也可先编译再运行：`CGO_ENABLED=0 go build -o bin/api .`，然后 `./bin/api -c manifest/config/config.yaml`（兼容写法；框架原生写法为 `./bin/api --gf.gcfg.file=manifest/config/config.yaml`；不传参数时自动搜索 `manifest/config/config.yaml`）。

### 2. 启动前端开发服务

```bash
cd web
npm install
npm run dev:vite     # http://localhost:5173，/api 代理到 127.0.0.1:8090
```

浏览器打开终端提示地址，使用 `admin / admin123` 登录。TypeScript 类型检查：`npx tsc --noEmit`；生产构建：`npm run build:vite`（产物输出 `web/dist/`）。

### 3. Electron 开发调试

```bash
# 先启动后端（步骤 1），另开终端：
cd web
npm install
npm run electron:dev    # concurrently 启动 Vite 与 Electron 窗口（加载 localhost:5173）
```

## 打包发布

### 后端二进制（纯 Go，可交叉编译，无需 CGO）

```bash
cd backend
# Windows
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-windows-amd64.exe .
# macOS
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o bin/api-darwin-arm64 .
# Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-linux-amd64 .
```

运行时同目录放置 `manifest/config/config.yaml`（或通过 `-c` 指定配置路径）；SQLite 数据文件按配置路径生成。

### 前端 / Electron 安装包

```bash
cd web
npm install
npm run build:vite       # 先构建前端静态产物 dist/
# 将后端二进制按 web/resources/README.txt 说明放入 extraResources 预留位后：
npm run build:win        # Windows nsis 安装包
npm run build:mac        # macOS dmg
npm run build:linux      # Linux AppImage
```

> 主进程在生产模式下会尝试启动后端二进制并加载前端 dist；具体资源路径与环境变量覆盖方式见 `web/electron/` 内注释。打包前请先完成后端编译并放入约定目录。

## 默认账号与权限

| 账号 | 密码 | 角色 | 权限 |
|---|---|---|---|
| admin | admin123 | admin | 供应商 / 模型 / Key / 用户全部管理；Key 分配；全局统计；全部会话查看；首次登录建议修改密码 |
| （admin 自建） | 由 admin 设置 | user | 仅见「对话测试」与个人统计；只能使用被分配且可用的 Key；仅看本人会话 |

修改密码：右上角下拉 → 修改密码。

## 数据库与初始化说明

- 首次启动自动执行建表：providers / models / api_keys / users / key_user / chat_sessions / chat_messages（字段、索引见 `docs/api-contract.md` 第 1 节）。
- 种子数据仅在对应表为空时写入：内置 admin 账号 + 少量标注「示例」的供应商 / 模型 / Key 记录，便于首次进入看到界面形态；示例 Key 为占位值，不可用于真实调用。
- SQLite 文件可直接备份 / 删除（删除后重启重新初始化）；切换 MySQL 只需把 `database.default.type` 改为 `mysql`、填写连接参数，并在 `main.go` 放开 mysql 驱动空白导入，表结构由程序自动创建（MySQL 需先建好空库、字符集 utf8mb4）。
- 时间字段以后端本地时间字符串（`YYYY-MM-DD HH:mm:ss`）输出，前端直接展示；数据库可按需调整时区配置。

## 对话转发规则与计费

- OpenAI 协议：`POST {base_url}/chat/completions`，请求头 `Authorization: Bearer {key}`；system 消息置于 messages 数组；用量取 `usage.prompt_tokens / completion_tokens`，缓存命中取 `prompt_tokens_details.cached_tokens`。
- Anthropic 协议：`POST {base_url}/v1/messages`，请求头 `x-api-key` + `anthropic-version: 2023-06-01`；`max_tokens` 必填（默认 4096）；system 合并为顶层字符串；用量取 `input_tokens / output_tokens / cache_read_input_tokens`。
- 单条消息费用 = 输入 Token 按是否命中缓存分别乘缓存命中价 / 未命中价 + 输出 Token 乘输出价（价格单位：元 / 百万 Token）；费用累加到对应 Key 的 `used_quota`；Key 停用、供应商停用、过期、超限额时不可发起对话。
- 并发：按供应商 `max_concurrency` 信号量限流；超时按供应商 `timeout_seconds` 控制。

## 常见问题

1. **端口被占用**：后端 8090、前端 5173 可分别在 `config.yaml` 的 `server.address` 与 `vite.config.ts` 的 `server.port` 修改；前端代理目标随之调整。
2. **CGO 编译报错 / 缺少 gcc**：本项目 SQLite 使用纯 Go 驱动，构建时保持 `CGO_ENABLED=0`；不要引入 mattn/go-sqlite3 等 CGO 驱动。
3. **依赖拉取失败**：Go 配置 `go env -w GOPROXY=https://goproxy.cn,direct`；npm 可切换国内镜像 `npm config set registry https://registry.npmmirror.com`。
4. **对话返回上游错误**：检查供应商 base_url 协议前缀是否正确、Key 是否有效或余额不足、模型是否在该供应商名下且已启用；后端日志会透出上游错误摘要（不打印完整 Key）。
5. **真实 Key 接入**：在「API-KEY 管理」新建记录粘贴真实 Key；在「模型管理」维护对应模型定价后即可在对话页选择使用；请勿在公开仓库提交真实 Key。
6. **MySQL 连接失败**：确认 `database.default.type=mysql`、连接参数正确、MySQL 已创建空库（字符集 utf8mb4），并已在 `main.go` 放开 mysql 驱动导入，表结构由程序自动创建。
7. **无有效上游 Key 时对话报错**：示例数据中的 Key 为占位值；失败请求会优雅提示「无法连接上游模型服务 / 上游模型服务返回错误」并保留用户消息记录，填入真实有效 Key 后即可走通成功链路（含 Token 用量解析与费用累计）。

## 已验证范围与未覆盖项（如实说明）

- **后端已实测通过（23 项接口断言全绿，含自动化冒烟脚本覆盖）**：
  - 编译与启动：`go vet` 无告警、`CGO_ENABLED=0 go build` 静态编译通过（约 32MB，无 CGO 依赖）；启动自动建 7 张表、写入内置 admin 与示例数据；`-c` 指定配置文件生效（库文件落到指定路径）。
  - 认证与权限：错误密码拒绝、无 Token 返回 401、JWT 签发；普通用户访问管理接口统一返回 403；`/me/keys` 仅返回已分配且可用的 Key；删除用户带事务级联清理分配关系 / 会话 / 消息，且禁止删除当前登录账号。
  - CRUD：供应商 / 模型 / Key / 用户列表与增改；供应商名下仍绑定模型或 Key 时拒绝删除；行内启停；Key 列表掩码回显、详情接口取明文复制。
  - 可用性拦截：Key 停用、供应商停用、Key 过期、额度超额四类场景均被拦截并返回明确提示；供应商停用时其下 Key 在 `/me/keys` 中被静默过滤。
  - 对话链路：未分配 Key 越权使用返回 403；上游不可达 / 非 2xx 时非流式返回友好错误、流式下发 SSE `error` 事件；失败对话的用户消息仍落库、会话标题正常生成、Token 与费用记 0 不计费；管理员可见全部会话，普通用户仅见本人会话。
  - 统计：管理员全局口径与普通用户个人口径分别返回对应字段。
- **前端已验证**：`npx tsc --noEmit` 零类型错误；`npm run build:vite` 构建通过；路由守卫与菜单按角色过滤；对话页 SSE 逐字渲染、单条 Token/费用与累计费用 / Token 显示。
- **复核中修复的问题**（交付前已回归通过）：① 用户列表 `COUNT(多列)` 在 SQLite 下报 500，改为计数查询不携带列表字段；② Key 搜索 `OR` 条件补显式括号，避免与供应商筛选 `AND` 组合时优先级穿透；③ 自定义命令未接管 `-c` 参数导致指定配置不生效，已在入口兼容 `-c` 与框架原生 `--gf.gcfg.file`。
- **第二轮（统计图表 + 单位规范 + 视觉主题）已验证**：
  - 后端新增 `GET /api/stats/rankings`、`/stats/trend`、`/stats/breakdown`：`go vet` 与 `CGO_ENABLED=0 go build` 通过；curl 实测空表返回 `[]` / 三项全零、趋势 7 天连续补零、`days=999` 收敛为 90 天、普通用户本人口径返回空数组不报错、无 Token 401；聚合结果经注入测试消息手算核对一致；SQLite 实测、MySQL 方言分支（`DATE()`、`GREATEST()`）代码同构未实测。
  - 前端 `npx tsc --noEmit` 零错误、`npm run build:vite` 通过；echarts 按需注册并独立分包；模型上下文长度按 `M` 两位小数展示、价格统一两位小数并标注「元/百万 Token」、金额统一 `¥x.xx`；四图表均有空态占位与 `ResizeObserver` 自适应；主色统一商务蓝 `#2563eb`，无紫 / 靛蓝配色；既有登录、CRUD、RBAC、对话流式降级、Key 掩码与分配流程无回归。
- **第三轮（显示修复 + 思考链路 + 创建即分配）已验证**：
  - 后端 `go vet` / `CGO_ENABLED=0 go build` 通过；新建 Key 事务内幂等写入 `key_user`，curl 确认 admin 创建的 Key 出现在 `/me/keys` 且 MeList 无 admin 全量特判；历史消息接口返回 `reasoning_content / model / provider_name`；旧库（无 `reasoning_content` 列）启动自动幂等补列（SQLite `PRAGMA` 探测 + ALTER，MySQL `information_schema` 分支代码就位），新库 DDL 直接带列；OpenAI `reasoning_content/reasoning` 与 Anthropic `thinking_delta/thinking` 解析、`reasoning` SSE 事件、非 SSE 降级序列均已实现并通过编译；旧接口与统计接口回归 200、401/403 行为不变。
  - 前端 `npx tsc --noEmit` 零错误、`npm run build:vite` 通过；仪表盘问候语改为函数调用（不再渲染源码文本）；模型弹窗加宽 640px、价格框与数字框统一全宽；assistant 气泡头部标签、思考过程折叠区（无内容隐藏 / 流式展开 / 完成折叠 / 历史默认折叠）落位；未新增运行时依赖。
- **第四轮（注册自助开通 + 真流式加固 + Markdown 渲染）已验证**：
  - 后端 `go vet` / `CGO_ENABLED=0 go build` 通过；`POST /api/auth/register` 实测：正常注册返回 `{id,username,role:"user",display_name}` 且无 token 字段；重名返回「用户名已存在」、短密码 / 非法用户名返回 400；新注册用户可登录但访问 `/providers`、`/keys` 返回 403、`/me/keys` 为空数组；`app.allow_register=false` 时返回「注册已关闭」且不影响已有用户登录；注册逻辑固定 `role=user/status=1`、bcrypt 哈希、响应不含密码哈希；其余接口与统计、降级错误事件回归正常。
  - 非 SSE 降级重写为分片模拟流式：reasoning 按句切、正文按 Markdown 结构切（代码围栏 / 标题 / 列表项整体不拆），约 20ms 片间隔逐片 `reasoning* → delta* → usage → done`，片间响应 `ctx.Done()` 取消；全部片发完才一次性落库计费；正式 SSE 透传、`X-Accel-Buffering: no`、逐行 Flush 保持不变。
  - 前端 `npx tsc --noEmit` 零错误、`npm run build:vite` 通过；新增 `marked ^12.0.2`、`dompurify ^3.4.17`（随对话页路由分包，不拖大首屏）；登录页登录 / 注册双 Tab、确认密码一致性校验、成功切回登录并回填用户名；assistant 正文经 DOMPurify 消毒后 `v-html`，user / reasoning 纯文本；assistant 气泡全宽、user 气泡 60% 靠右；reader 兼容 `\r\n`，`onDone` 不重置消息文本。
- **未覆盖项（需在真实环境补验）**：沙箱网络无法访问真实 `api.openai.com` / `api.anthropic.com`，且示例 Key 为占位值，因此「真实上游逐字 / 分片流式体感、真实思考模型 reasoning 增长、上游 200 成功时的 usage 解析、缓存 Token 识别、费用折算与 `used_quota` 累加、Markdown 模型输出（表格 / 代码块 / 列表）实际渲染效果、Anthropic 成功响应解析」未做端到端实测，代码已完整实现，接入真实有效 Key 即可走通；图表在真实数据下的比例、Tooltip 与轴标签密度未做浏览器人工点验（空态与构建已验证）；Electron 桌面窗口与三平台安装包在无图形界面 / 受限网络的构建沙箱内未实际出包，主进程、preload 与 builder 配置已按规范编写，请在桌面环境按上述命令完成最终打包；安装包图标使用 electron-builder 默认图标，正式发布前请补充；本轮未引入 highlight.js，代码块仅有暗色等宽基础样式、无语法高亮配色，后续可按需增量接入。

## 接口文档

完整字段、端点、SSE 事件格式见 [`docs/api-contract.md`](./docs/api-contract.md)。
