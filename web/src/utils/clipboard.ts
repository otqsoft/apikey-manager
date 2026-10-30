/**
 * 复制到剪贴板（兼容 http 站点）
 *
 * navigator.clipboard 只在「安全上下文」下可用（https、localhost、file://），
 * 部署在 http://IP:端口 这类地址上时该 API 为 undefined，直接调用会抛错。
 * 因此提供两级回退：
 *   1. navigator.clipboard.writeText（https / localhost）
 *   2. 临时 textarea + document.execCommand('copy')（http 站点可用）
 *   3. 弹出可编辑输入框，让用户手动复制
 */
import { ElMessage, ElMessageBox } from 'element-plus'

async function writeClipboard(text: string): Promise<boolean> {
  // 1. 现代 API
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch (_) {
    /* 继续走回退方案 */
  }

  // 2. execCommand 回退
  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.setAttribute('readonly', '')
    ta.style.position = 'fixed'
    ta.style.top = '-9999px'
    ta.style.left = '-9999px'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.focus()
    ta.select()
    ta.setSelectionRange(0, text.length)
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch (_) {
    return false
  }
}

/** 复制文本并给出统一的成功/失败提示 */
export async function copyText(text: string, label: string): Promise<void> {
  if (!text) {
    ElMessage.warning(`${label}为空，无可复制内容`)
    return
  }
  if (await writeClipboard(text)) {
    ElMessage.success(`${label}已复制到剪贴板`)
    return
  }
  // 3. 兜底：弹窗让用户手动复制
  try {
    await ElMessageBox.prompt(`当前页面为非安全上下文（http），无法自动写入剪贴板，请手动复制：`, '手动复制', {
      confirmButtonText: '已复制',
      cancelButtonText: '关闭',
      inputType: 'textarea',
      inputValue: text
    })
  } catch (_) {
    /* 用户关闭 */
  }
}

/** 只写入剪贴板、不弹提示，返回是否成功（供需要自定义提示的地方使用） */
export async function copySilently(text: string): Promise<boolean> {
  if (!text) return false
  return await writeClipboard(text)
}
