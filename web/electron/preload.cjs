/**
 * preload 脚本：通过 contextBridge 向渲染进程暴露只读的应用元信息
 * 渲染进程通过 window.appInfo 访问；不暴露 Node 能力，保持安全边界
 */
const { contextBridge } = require('electron')

contextBridge.exposeInMainWorld('appInfo', {
  name: 'LLM密钥管理台',
  version: '1.0.0',
  platform: process.platform,
  isElectron: true
})
