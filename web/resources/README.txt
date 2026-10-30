后端二进制放置说明（本目录下的可执行文件由脚本生成，不入库）：

一、自动生成（推荐）
  在 web/ 目录执行：
    npm run sync:backend
  脚本会把 ../../backend/bin/ 下的编译产物同步到这里：
    resources/backend.exe                 后端可执行文件（Linux/macOS 为 backend）
    resources/manifest/config/config.yaml 后端配置（GoFrame 默认查找路径）
  已存在的 config.yaml 不会被覆盖，如需更新加 --force。

二、手动放置
  把编译好的后端可执行文件命名为 backend（Windows 下为 backend.exe）放入本目录，
  并把 manifest/config/config.yaml 一并放入，保持 manifest/config/ 层级。

三、打包后的运行方式
  electron-builder 通过 extraResources 把本目录的内容拷贝到安装包的
  resources/backend/ 下；主进程以该目录为工作目录启动 backend.exe，
  配置文件即 resources/backend/manifest/config/config.yaml。

四、接口地址
  - 浏览器开发模式：前端走 Vite dev server 的 /api 代理 -> http://127.0.0.1:8090
  - 桌面端打包运行：页面以 file:// 协议加载，没有代理，前端直连
    http://127.0.0.1:8090/api（可用构建期环境变量 VITE_API_BASE 覆盖）
  - 也可用环境变量指定后端二进制路径：
    Windows PowerShell : $env:BACKEND_BIN="D:\path\backend.exe"; npm run electron:dev
    Linux/macOS       : BACKEND_BIN=/opt/akmanager/backend npm run electron:dev

五、未放置后端二进制时
  Electron 窗口仍会打开（纯前端模式），但登录等接口请求会失败，
  控制台会提示「未找到后端二进制（resources/backend）」。
