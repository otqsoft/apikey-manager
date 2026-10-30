/**
 * Electron 主进程（CommonJS）
 * - 开发模式（electron . 且未打包）：加载 http://localhost:5173
 * - 生产打包后：加载 ../dist/index.html，并尝试启动同包内 resources/backend 下的后端二进制
 * - 后端二进制路径可用环境变量 BACKEND_BIN 覆盖；启动失败仅在控制台提示，保持纯前端窗口
 */
const { app, BrowserWindow } = require('electron')
const fs = require('fs')
const path = require('path')
const { spawn } = require('child_process')

const DEV_URL = 'http://localhost:5173'

let mainWindow = null
let backendProcess = null

/** 尝试拉起本地后端二进制，失败不阻塞前端窗口 */
function tryStartBackend() {
  const candidates = []
  if (process.env.BACKEND_BIN) candidates.push(process.env.BACKEND_BIN)
  // 打包后 extraResources 输出到 resources/backend
  candidates.push(path.join(process.resourcesPath || '', 'backend'))
  candidates.push(path.join(app.getAppPath(), 'resources', 'backend'))
  candidates.push(path.join(__dirname, '..', 'resources', 'backend'))

  const bin = candidates.find((p) => {
    if (!p) return false
    try {
      return fs.existsSync(p)
    } catch (_) {
      return false
    }
  })
  if (!bin) {
    console.warn('[electron] 未找到后端二进制（resources/backend），纯前端模式运行')
    return
  }
  try {
    backendProcess = spawn(bin, [], {
      stdio: 'inherit',
      cwd: path.dirname(bin)
    })
    backendProcess.on('error', (err) => {
      console.warn('[electron] 后端二进制启动失败，保持纯前端窗口运行：', err.message)
      backendProcess = null
    })
    backendProcess.on('exit', (code) => {
      console.warn(`[electron] 后端进程退出，code=${code}，前端窗口保持可用`)
    })
  } catch (err) {
    console.warn('[electron] 未找到后端二进制（resources/backend），纯前端模式：', err.message)
  }
}

/** 窗口 / 任务栏图标：优先用打包后 dist 内的副本，未打包时回落到 public 源图 */
function resolveIcon() {
  const candidates = [path.join(__dirname, '..', 'dist', 'favicon.ico'), path.join(__dirname, '..', 'public', 'favicon.ico')]
  return candidates.find((p) => {
    try {
      return fs.existsSync(p)
    } catch (_) {
      return false
    }
  })
}

function createWindow() {
  const icon = resolveIcon()
  mainWindow = new BrowserWindow({
    width: 1440,
    height: 900,
    minWidth: 1100,
    minHeight: 680,
    autoHideMenuBar: true,
    icon,
    webPreferences: {
      preload: path.join(__dirname, 'preload.cjs'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false
    }
  })

  if (!app.isPackaged) {
    // 开发：先等 Vite dev server；若已设置 BACKEND_BIN 也尝试启动后端
    if (process.env.BACKEND_BIN) tryStartBackend()
    mainWindow.loadURL(DEV_URL)
    mainWindow.webContents.openDevTools({ mode: 'detach' })
  } else {
    tryStartBackend()
    mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'))
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

app.whenReady().then(() => {
  createWindow()
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (backendProcess) backendProcess.kill()
  if (process.platform !== 'darwin') app.quit()
})
