/**
 * 把后端编译产物同步到 Electron extraResources 预留位 web/resources/
 *
 * 产物布局（打包后落到安装目录 resources/backend/ 下，主进程以该目录为工作目录启动）：
 *   web/resources/backend.exe                 后端可执行文件（Windows）
 *   web/resources/manifest/config/config.yaml 后端配置（GoFrame 默认查找路径）
 *
 * 用法：npm run sync:backend
 *      已存在的 config.yaml 不会被覆盖（避免冲掉本地改过的配置），加 --force 强制覆盖。
 */

const fs = require('fs')
const path = require('path')

const ROOT = path.dirname(__dirname)
const BACKEND_BIN = path.join(ROOT, '..', 'backend', 'bin')
const RESOURCES = path.join(ROOT, 'resources')

const force = process.argv.includes('--force')
const isWin = process.platform === 'win32'

function copy(src, dst, { overwrite = true } = {}) {
  if (!fs.existsSync(src)) return false
  fs.mkdirSync(path.dirname(dst), { recursive: true })
  if (fs.existsSync(dst) && !overwrite) return 'skipped'
  fs.copyFileSync(src, dst)
  return true
}

function main() {
  if (!fs.existsSync(BACKEND_BIN)) {
    console.error('未找到后端产物目录：' + BACKEND_BIN)
    console.error('请先在 backend/ 下编译：GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api-key.exe .')
    process.exit(1)
  }

  const srcBin = isWin ? path.join(BACKEND_BIN, 'api-key.exe') : path.join(BACKEND_BIN, 'api-key')
  const dstBin = path.join(RESOURCES, isWin ? 'backend.exe' : 'backend')

  if (!fs.existsSync(srcBin)) {
    console.error('未找到后端可执行文件：' + srcBin)
    process.exit(1)
  }

  console.log('复制到: ' + dstBin)
  copy(srcBin, dstBin)
  console.log('  ' + srcBin + '  ->  ' + dstBin)

  // 配置文件：缺省从 backend/bin/manifest/config/config.yaml 带入
  const srcCfg = path.join(BACKEND_BIN, 'manifest', 'config', 'config.yaml')
  const dstCfg = path.join(RESOURCES, 'manifest', 'config', 'config.yaml')
  if (fs.existsSync(srcCfg)) {
    const r = copy(srcCfg, dstCfg, { overwrite: force })
    console.log(r === 'skipped' ? '配置已存在，跳过（如需覆盖加 --force）: ' + dstCfg : '  ' + srcCfg + '  ->  ' + dstCfg)
  } else {
    console.warn('未找到配置文件，后端启动后会按内置默认值运行: ' + srcCfg)
  }

  console.log('\n完成，可以执行 npm run build:win 打包了。')
}

main()
