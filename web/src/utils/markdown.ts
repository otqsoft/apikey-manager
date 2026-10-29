/**
 * Markdown 渲染工具（契约 8.3）
 * - marked 解析 + DOMPurify.sanitize 清洗后输出 HTML，禁止未过滤的 v-html
 * - 空串返回空串；解析异常兜底为 HTML 转义后的纯文本
 * - 本轮取舍：不引入 highlight.js，代码块仅做暗色基础样式，无语法高亮
 * - user 消息与 reasoning 折叠区不走本函数（保持纯文本 pre-wrap）
 */
import { marked } from 'marked'
import DOMPurify from 'dompurify'

marked.setOptions({
  gfm: true, // GitHub 风格：表格、删除线、任务列表等
  breaks: true // 单换行渲染为 <br>，对话长文体验更自然
})

/** HTML 转义（异常兜底分支使用，避免注入） */
function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

/**
 * 渲染 Markdown 为安全 HTML 字符串
 * @param text 原始 Markdown 文本；空串/空白返回空串
 */
export function renderMarkdown(text: string | null | undefined): string {
  if (!text || !text.trim()) return ''
  try {
    const rawHtml = marked.parse(text, { async: false }) as string
    // DOMPurify 默认白名单清洗，移除脚本/事件属性等危险内容
    return DOMPurify.sanitize(rawHtml)
  } catch {
    return `<pre class="md-fallback">${escapeHtml(text)}</pre>`
  }
}
