后端二进制放置说明（预留位，后端由独立工程产出，本目录不提交可执行文件）：

1. 打包 Electron 时，本目录（web/resources/）会被 electron-builder 配置 extraResources
   拷贝到安装包的 resources/backend/ 目录下。
2. 请把编译好的后端可执行文件命名为 `backend`（Windows 下为 `backend.exe`），
   放入本目录（resources/）后再执行 build:electron。
3. 也可在开发/启动时通过环境变量覆盖路径：
   Windows PowerShell : $env:BACKEND_BIN="D:\path\backend.exe"; npm run electron:dev
   Linux/macOS       : BACKEND_BIN=/opt/akmanager/backend npm run electron:dev
4. 若未放置后端二进制，Electron 窗口仍会正常打开（纯前端模式），
   但接口请求会失败；开发期请先单独启动后端（http://127.0.0.1:8090），
   前端通过 Vite 代理访问。
