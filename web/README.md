# 多平台大模型 API-KEY 集中管理工具（前端 + Electron 壳）

技术栈：Vue 3（组合式 API / `<script setup>`）+ Vite 5 + TypeScript + Element Plus（zh-cn）+ Pinia + Vue Router 4 + axios + ECharts 5 + marked + DOMPurify + Electron 33。
接口契约唯一依据：`../docs/api-contract.md`（冻结版 v1 + 第 6/7/8 节补充）。

## 1. 常用命令

| 命令 | 说明 |
|---|---|
| `npm install` | 安装依赖；若离线环境可设 `ELECTRON_SKIP_BINARY_DOWNLOAD=1` 跳过 Electron 二进制下载 |
| `npm run dev:vite` | 仅启动 Vite 开发服务器（http://localhost:5173），`/api` 代理到 http://127.0.0.1:8090 |
| `npm run electron:dev` | 同时启动 Vite 与 Electron 窗口（concurrently + wait-on） |
| `npm run build:vite` | 仅前端生产构建，产物输出到 `dist/` |
| `npm run build:electron` | 前端构建后调用 electron-builder 按 package.json `build` 段产出安装包 |
| `npm run build:win` / `build:mac` / `build:linux` | 分别产出 win nsis / mac dmg / linux AppImage |

后端独立运行于 `http://127.0.0.1:8090`；前端不直连第三方模型服务商。

## 2. 目录结构

```
web/
├── electron/               # 主进程 main.cjs + preload.cjs
├── resources/             # 打包时 extraResources → resources/backend（后端二进制预留位）
├── src/
│   ├── api/                # 按模块拆分：auth/providers/models/keys/users/chat/stats
│   ├── types/              # 共享 TS 类型，逐字段对齐契约（第 6 节统计 / 第 7 节消息扩展 / 第 8 节注册）
│   ├── components/ChartCard.vue   # 通用图表卡片：空态 + ResizeObserver + 卸载 dispose
│   ├── utils/request.ts    # axios 实例：Bearer 注入、{code,msg,data} 解包、401/403 处理
│   ├── utils/echarts.ts     # ECharts 5.x 按需注册（Bar/Line/Pie/Grid/Tooltip/Legend/Title/Canvas）
│   ├── utils/markdown.ts     # marked 解析 + DOMPurify 清洗；异常兜底转义纯文本（无语法高亮）
│   ├── utils/format.ts      # 金额两位小数 / 上下文 M 展示 / 千分位 / 时间格式化
│   ├── styles/index.css     # 全局主题：主色 #2563eb、卡片/表格/滚动条/字体栈
│   ├── store/user.ts        # Pinia：token + 用户信息 + permissions，localStorage 持久化
│   ├── router/              # 路由表 + 守卫（角色过滤、未登录跳转）
│   ├── layout/             # 整体布局：侧边菜单 + 顶栏下拉 + 修改密码弹窗
│   └── views/              # login / dashboard / providers / models / keys / users / chat
├── vite.config.ts          # port 5173；proxy /api → 127.0.0.1:8090；outDir dist；echarts 分包
└── tsconfig.json
```

## 3. 页面 ↔ 契约对照

| 页面 | 路由 | 角色 | 主要接口 |
|---|---|---|---|
| 登录 / 注册 | /login | 公开 | POST /api/auth/login；POST /api/auth/register（8.1，不签发 token）；PUT /api/auth/password |
| 仪表盘 | /dashboard | admin/user | GET /api/stats/overview；图表：GET /api/stats/rankings、GET /api/stats/trend、GET /api/stats/breakdown（契约第 6 节） |
| 供应商管理 | /providers | admin | GET/POST /api/providers，GET/PUT/DELETE /api/providers/:id，PUT /:id/status |
| 模型管理 | /models | admin | GET/POST /api/models，GET/PUT/DELETE /:id，PUT /:id/status；供应商下拉 GET /api/providers?status=1 |
| API-KEY 管理 | /keys | admin | GET/POST /api/keys，详情取明文复制，PUT /:id/status，GET/PUT /:id/users |
| 用户管理 | /users | admin | GET/POST /api/users，PUT /:id，PUT /:id/status，POST /:id/reset-password，GET /:id/keys；禁止删自己 |
| 对话测试 | /chat | admin/user | GET /api/me/keys；GET/POST /api/chat/sessions；GET /:id/messages；DELETE /:id；POST /api/chat/completions（SSE：delta/reasoning/usage/done/error） |

## 4. 关键实现说明

- 统一响应包 `{code,msg,data}`：`code===0` 解包返回 `data`；`401` 清登录态跳 `/login`；`403/400/500` 统一 ElMessage 提示。
- 列表类接口返回掩码 key，复制明文时调详情接口 `GET /api/keys/:id` 取完整 `api_key`。
- 额度进度条：`max_quota=0` 显示「不限」；`used_quota/max_quota` 超 90% 标红。
- 对话流式：`fetch + ReadableStream` 逐行解析 `data: {...}`；delta 追加正文、reasoning 累加思考过程、usage 回填 token/费用、done 刷新会话列表并自动折叠思考区、error 弹错；读取结束兜底解析残行，`finally` 恒复位 streaming，兼容后端非 SSE 降级补发序列。
- 路由守卫：无 token → /login；token 无档案 → 拉 GET /api/auth/profile；非 admin 访问 admin 路由 → 提示并回 /dashboard。
- Electron 生产模式尝试 `resources/backend` 后端二进制（`BACKEND_BIN` 环境变量可覆盖），启动失败仅控制台告警，保持纯前端窗口。

### 4.1 仪表盘图表（契约第 6 节 v2 统计接口）

| 图表 | 接口 | 说明 |
|---|---|---|
| Top 榜单（横向条形） | `GET /api/stats/rankings?dim=model\|provider&metric=calls\|tokens\|cost&limit=10` | Tab 切换维度（模型/供应商）与指标（调用次数/Token/费用）；y 轴 `inverse` 使 Top1 在最上 |
| 供应商调用占比（环形） | `GET /api/stats/rankings?dim=provider&metric=calls` | 中心显示总调用次数 |
| 用量趋势（柱+折线双轴） | `GET /api/stats/trend?days=7\|30` | 柱=费用（左轴），折线=Token（右轴）；7/30 天切换 |
| Token 用量构成（环形） | `GET /api/stats/breakdown` | 固定三项：缓存命中输入 / 未命中输入 / 输出；中心显示总 Token |

- 四个图表均 `try/catch + Promise.allSettled` 兜底：单接口失败或空数组时 ChartCard 显示「暂无调用数据」空态，页面不白屏、不抛未捕获异常。
- `ChartCard.vue` 内置 ResizeObserver 监听容器尺寸自动 `chart.resize()`，卸载时 `dispose()`；ECharts 按需注册并在 vite `manualChunks` 独立分包。
- admin/user 调用同一组接口，数据范围由后端按角色收敛。

### 4.2 展示单位规范

- 上下文长度：`formatContext(n)` 按百万 Token 展示（128000→0.13M、1000000→1.00M）；表单仍录入原始 Token 数，旁注「录入原始 Token 数，列表按 M 展示」。
- 模型价格：列表三列统一两位小数并标注「元/百万 Token」；表单 `el-input-number` `precision=2、step=0.1、min=0`。
- 金额：`formatMoney` 统一 `¥ + 两位小数`；`formatMoneyCompact` 保留过万 `¥x.xx万` 紧凑格式；Key 额度、对话费用同为两位小数。

### 4.3 主题视觉

- 主色统一商务蓝 `#2563eb`（Element Plus CSS 变量覆盖于 `src/styles/index.css`），中性灰青辅色；侧边栏深色渐变 + 菜单激活左侧高亮条；卡片白底圆角 12px、柔和阴影；数字 `tabular-nums` 等宽。
- 对话页：用户气泡右侧蓝色、助手气泡左侧白色；流式末尾呼吸光标（CSS 动画）；输入栏顶部阴影；用量行 token/费用精致排版。

### 4.4 第三轮修复（契约第 7 节 v3）

- **消息扩展**：`ChatMessage` 增加 `reasoning / model / provider_name`；历史消息接口 `reasoning_content` 在 `api/chat.ts` 统一映射为 `reasoning`。
- **思考过程折叠区**：`onReasoning` 独立累加；无内容不渲染；流式思考阶段展开、`done` 后自动折叠；历史回放默认折叠；标题栏显示字数，箭头随展开旋转；浅灰底与正文气泡区分。
- **assistant 头部标签**：气泡上方小字展示「供应商 · 模型」；实时对话取当前 Key 的 `provider.name` + 选中模型兜底，历史回放用接口返回值；user 消息不显示。
- **流式加固**：请求恒 `stream:true`；`finally` 必复位 `streaming`；解析循环结束兜底处理无换行残行；兼容后端把不支持流式的上游降级为一次性 delta/usage/done 序列，不卡死、不残留 loading。
- **Key 创建即分配（7.1）**：由后端事务写入 `key_user`，前端「分配用户」批量覆盖弹窗行为不变，创建成功后无需再调分配接口。

### 4.5 第四轮增强（契约第 8 节 v4）

**注册（8.1）**
- `POST /api/auth/register` 公开接口；`types` 新增 `RegisterReq / RegisterResult`；登录页改为登录 / 注册双 Tab。
- 注册校验：用户名 3~32 位字母 / 数字 / 下划线；密码 ≥6 位；确认密码一致性校验；昵称可选；后端错误（用户名已存在 / 注册已关闭 / 参数错误）由 axios 拦截器透传 `msg`。
- 注册成功不签发 token（契约 8.1）：提示「注册成功，请登录」后切回登录 Tab 并回填用户名。

**流式加固（8.2）**
- `onDelta / onReasoning` 纯字符串累加；`onDone` 仅折叠思考区与刷新会话列表，不重置 `content / reasoning`；reader 切行前将 `\r\n` 归一化为 `\n`；残行兜底解析保留；`finally` 恒复位 `streaming`，异常 / 降级序列均不卡死。

**Markdown 渲染（8.3）**
- `src/utils/markdown.ts`：`marked` 解析 + `DOMPurify.sanitize` 清洗；空串返回空串；解析异常兜底为 HTML 转义纯文本。
- assistant 正文 `v-html` 渲染（经 sanitize）；user 消息与 reasoning 折叠区保持纯文本 `pre-wrap`；历史回放复用同一 `renderMarkdown`。
- 布局：assistant 气泡占满可用宽度（白底圆角卡片）；user 气泡靠右约 60%、蓝底白字；MD 样式覆盖标题层级、段落行高 1.7、列表缩进、引用块、行内 code、代码块深色 `#0f172a` 等宽 + 横向滚动、表格细线斑马纹、链接主题色、图片 `max-width:100%`。
- **取舍说明**：本轮不引入 `highlight.js`，代码块仅做暗色基础样式、无语法高亮；后续如需高亮可单独增量引入并在 `markdown.ts` 接入 marked renderer。

## 5. 打包注意

1. 安装包图标未内置（electron-builder 使用默认图标），正式发布前请在 `build.win/mac/linux.icon` 补图标路径。
2. 后端可执行文件放入 `web/resources/`（Windows 命名 `backend.exe`），随 extraResources 打入安装包 `resources/backend/`。
3. 沙箱 / CI 环境若无 GitHub 网络，`npm install` 时设 `ELECTRON_SKIP_BINARY_DOWNLOAD=1` 可完成前端构建；`electron-builder` 出包需可下载对应平台打包依赖。
