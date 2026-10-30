# 多平台大模型 API-KEY 集中管理工具

<div align="center">
	<img src="docs/logo.png" width="260" height="260">
	<p align="center">
    <img src="https://img.shields.io/badge/-Vue3-34495e?logo=vue.j" />
    <img src="https://img.shields.io/badge/-Vite2.7-646cff?logo=vite&logoColor=white" />
    <img src="https://img.shields.io/badge/-TypeScript-blue?logo=typescript&logoColor=white" />
    <img src="https://img.shields.io/badge/-Pinia-yellow?logo=picpay&logoColor=white" />
    <img src="https://img.shields.io/badge/-ESLint-4b32c3?logo=eslint&logoColor=white" />
    <img src="https://img.shields.io/badge/-Axios-008fc7?logo=axios.js&logoColor=white" />
    <img src="https://img.shields.io/badge/-Prettier-ef9421?logo=Prettier&logoColor=white" alt="Prettier">
    <img src="https://img.shields.io/badge/-Less-1D365D?logo=less&logoColor=white" alt="Less">
    <img src="https://pkg.go.dev/badge/github.com/gogf/gf/v2.svg" >
    <img src="https://goreportcard.com/badge/github.com/gogf/gf" >
	</p>
	<p>&nbsp;</p>
</div>

一个面向个人 / 小团队的桌面端「模型 API-KEY 统一管理与对话测试台」：集中维护多个模型供应商（OpenAI 兼容 / Anthropic 两种协议）、模型定价、API-KEY 额度与分配关系，内置对话测试页并自动统计 Token 消耗与费用。后端持有密钥，前端不接触第三方 Key；普通用户只能使用被分配给自己的 Key。

- 后端：Go + GoFrame v2 + gdb（默认 MySQL，可切回 SQLite 纯 Go 驱动）
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

- Go 1.21+（开发机已在 Go 1.23 验证；切回 SQLite 时为纯 Go 驱动，无需 CGO / gcc）
- Node.js 18+（推荐 20/22）、npm
- MySQL 8.0+（当前默认数据库；本机 5.7 已验证。需先建空库、字符集 utf8mb4）

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
| database.default.type | 驱动名：`mysql`（当前默认）或 `sqlite`（纯 Go 无 CGO） |
| database.default.host/port/name/user/pass/charset | MySQL 连接参数（默认 `127.0.0.1:3306`，库名 `apikey_manager`，utf8mb4；需先手动建空库） |
| sqlite 配置注释节 | 切回 SQLite 时取消注释、注释 mysql 节，并在 `main.go` 放开 sqlite 驱动空白导入 |
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
>
> **接口地址**：浏览器开发模式下前端走 Vite dev server 的 `/api` 代理（`127.0.0.1:8090`）；**桌面端打包运行后页面以 `file://` 协议加载、不存在代理，前端直连 `http://127.0.0.1:8090/api`**（见 `src/utils/request.ts` 的 `resolveBaseURL`，可用构建期环境变量 `VITE_API_BASE` 覆盖）。因此桌面端需要后端同机运行在 8090，或按下方 `sync:backend` 把后端二进制打进安装包由主进程自动拉起。
>
> **后端二进制入库（extraResources）**：在 `web/` 目录执行 `npm run sync:backend`，会把 `backend/bin/` 的编译产物同步到 `web/resources/`（可执行文件 + `manifest/config/config.yaml`），打包后落到安装目录的 `resources/backend/`，主进程启动时会拉起它，用户双击桌面端即可直接使用。该目录下的可执行文件与配置不入库（已在 `.gitignore` 中）。

> **图标**：桌面主程序 exe、安装包、卸载程序以及窗口 / 任务栏图标统一使用 `public/favicon.ico`。electron-builder 要求图标源至少 256×256（`public/favicon.ico` 原图为 180×180 单帧，直接使用会报 `ERR_ICON_TOO_SMALL`），因此由 `web/scripts/make-icon.py` 生成包含 16~256 共 7 档尺寸的多分辨率版本 `web/build/icon.ico` 供打包使用。更换 `public/favicon.ico` 后请先执行 `npm run build:icon` 重新生成，再打包。
>
> Electron 运行时体积较大（约 120 MB），首次打包需下载。仓库内 `web/.npmrc` 已配置 `electron_mirror` 指向 npmmirror 国内镜像，避免官方源中断后留下**损坏的 zip 缓存**（表现为打包时报 `ENOENT ... rename 'win-unpacked\electron.exe'`）。

### Web 部署（nginx）

`npm run build:vite` 产出的 `web/dist` 为纯静态文件，直接交给 nginx 托管即可。**但必须同时把 `/api/` 反向代理给后端**，否则接口会全部失败（后端路由挂在 `/api` 前缀下，前端 axios 的 baseURL 也是相对路径 `/api`）。完整示例见 [`docs/nginx.conf.example`](./docs/nginx.conf.example)，关键两段：

```nginx
root /path/to/apikey-manager/web/dist;   # 前端产物
index index.html;

location /api/ {                          # 后端接口（注意 proxy_pass 结尾不要带路径）
    proxy_pass         http://127.0.0.1:8090;
    proxy_buffering    off;               # SSE 流式对话必需，否则内容会攒到最后一次性输出
    proxy_read_timeout 300s;
}
```

> 后端**未开启 CORS**，因此不要把前端指向 `http://其它主机:8090/api` 跨域直连，浏览器会拦成 `Network Error`。若确实需要跨域，可在 nginx 给 `/api/` 加 `add_header Access-Control-Allow-Origin '*';` 并处理 `OPTIONS` 预检。
>
> 部署在子路径（如 `http://host/ak/`）时，构建前设置 `VITE_API_BASE=/ak/api`；或直接编辑 `dist/index.html` 注入 `<script>window.__AK_API_BASE__ = '/ak/api'</script>`（运行时覆盖，无需重新构建）。

## 默认账号与权限

| 账号 | 密码 | 角色 | 权限 |
|---|---|---|---|
| admin | admin123 | admin | 供应商 / 模型 / Key / 用户全部管理；Key 分配；全局统计；全部会话查看；首次登录建议修改密码 |
| （admin 自建） | 由 admin 设置 | user | 仅见「对话测试」与个人统计；只能使用被分配且可用的 Key；仅看本人会话 |

修改密码：右上角下拉 → 修改密码。

## 数据库与初始化说明

- 首次启动自动执行建表：providers / models / api_keys / users / key_user / chat_sessions / chat_messages（字段、索引见 `docs/api-contract.md` 第 1 节）。
- 种子数据仅在对应表为空时写入：内置 admin 账号 + 少量标注「示例」的供应商 / 模型 / Key 记录，便于首次进入看到界面形态；示例 Key 为占位值，不可用于真实调用。
- MySQL 需先建好空库（字符集 utf8mb4，表结构由程序自动创建）；切回 SQLite 只需把 `database.default.type` 改为 `sqlite`、在 `main.go` 放开 sqlite 驱动导入，数据文件按配置路径生成。
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
8. **`build:win` 报 `ENOENT: no such file or directory, rename '...\win-unpacked\electron.exe'`**：Electron 运行时 zip 缓存损坏导致解压只落地部分文件。删除缓存后重打即可：
   ```powershell
   # 定位并校验缓存（正常约 120 MB，若明显偏大或 tar 报 "Damaged Zip archive" 即为损坏）
   Get-ChildItem "$env:LOCALAPPDATA\electron\Cache\electron-v*-win32-x64.zip"
   tar -tf "$env:LOCALAPPDATA\electron\Cache\electron-v33.4.11-win32-x64.zip"
   # 删除损坏的缓存（可先改名备份），再重新打包
   Remove-Item "$env:LOCALAPPDATA\electron\Cache\electron-v33.4.11-win32-x64.zip"
   $env:ELECTRON_MIRROR='https://npmmirror.com/mirrors/electron/'; npm run build:win
   ```
9. **`build:win` 报 `remove '...\win-unpacked\resources\app.asar': The process cannot access the file`**：上一次的 `web/release` 目录被安全软件（联想电脑管家 / 火绒引擎 / Defender）或仍在运行的编辑器占用句柄，electron-builder 清理旧目录失败。关闭占用程序后删除 `web/release` 再打包，或临时换输出目录：
   ```powershell
   npx electron-builder --win nsis --config.directories.output=release2
   ```
10. **安装包缺少后端**：`web/resources/` 默认为空（仅说明文件），打包出的安装包不含后端二进制，桌面端以纯前端模式启动、接口不可用；按 `web/resources/README.txt` 放入 `backend.exe` 后重新打包即可。
11. **更换桌面 / 安装包图标**：electron-builder 要求图标源至少 256×256，直接指向 180×180 的 `public/favicon.ico` 会报 `ERR_ICON_TOO_SMALL`。替换 `public/favicon.ico` 后执行 `npm run build:icon`（需要 Python + Pillow）重新生成 `web/build/icon.ico` 再打包。
12. **桌面端打开后白屏**：Electron 打包后用 `loadFile`（`file://` 协议）加载 `dist/index.html`，若 Vite 未设置相对路径 base，资源引用会是 `/assets/xxx` 并被解析成 `file:///D:/assets/xxx` 导致全部加载失败。本项目 `vite.config.ts` 已设置 `base: './'`，构建产物中的引用应为 `./assets/xxx`；改动构建配置时注意保留。
13. **桌面端登录报网络错误 / 接口失败**：打包后没有 Vite 代理，`/api` 相对地址在 `file://` 下解析异常。前端已内置判断——页面协议为 `file:` 时直连 `http://127.0.0.1:8090/api`；若后端跑在其它端口，构建时用 `VITE_API_BASE=http://127.0.0.1:9000/api npm run build:win` 覆盖。另需确认后端已启动（`npm run sync:backend` 打入后端后由主进程自动拉起）。
14. **dist 放到 nginx 后接口报「网络异常」**：nginx 只托管了静态文件、没有反代 `/api/`。排查方式——浏览器 F12 → Network，看失败请求的完整 URL：请求地址是 `http://站点/api/auth/login` 但返回 404/返回了 index.html，即缺少反代配置；按 [`docs/nginx.conf.example`](./docs/nginx.conf.example) 加 `location /api/ { proxy_pass http://127.0.0.1:8090; proxy_buffering off; }`。若返回 502 则是后端没起；若报 `Network Error` 且无状态码，多半是配了跨域绝对地址而后端没开 CORS。

## 接口文档

完整字段、端点、SSE 事件格式见 [`docs/api-contract.md`](./docs/api-contract.md)。
