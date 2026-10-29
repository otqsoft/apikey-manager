/**
 * 对话模块（契约 2.6 / 2.7 / 7.2 / 7.3）
 * - 会话、消息列表、删除走 axios
 * - POST /api/chat/completions 流式（SSE）走 fetch + ReadableStream，不能用 axios
 * - SSE 事件：delta（正文增量）/ reasoning（思考增量）/ usage / done / error
 */
import { http } from '@/utils/request'
import type { ChatMessage, ChatSession, CompletionPayload, PageResult } from '@/types'

/** GET /api/chat/sessions 会话列表（admin 全部；user 仅本人，按创建时间倒序） */
export const getSessionPage = (params: { page?: number; page_size?: number } = {}) =>
  http.get<PageResult<ChatSession>>('/chat/sessions', { params: { page: 1, page_size: 100, ...params } })

/** POST /api/chat/sessions 新建会话 {key_id, model} */
export const createSession = (data: { key_id: number; model: string }) =>
  http.post<ChatSession>('/chat/sessions', data)

/**
 * GET /api/chat/sessions/:id/messages 消息列表正序；越权 403
 * 契约 7.2：后端返回 reasoning_content，前端映射到 reasoning；model/provider_name 透传
 */
export const getSessionMessages = async (id: number): Promise<ChatMessage[]> => {
  const res = await http.get<Array<ChatMessage & { reasoning_content?: string }>>(`/chat/sessions/${id}/messages`)
  return (res || []).map((m) => ({
    ...m,
    reasoning: m.reasoning ?? m.reasoning_content ?? '',
    reasoningOpen: false // 历史回放默认折叠
  }))
}

/** DELETE /api/chat/sessions/:id 删除会话（仅本人或 admin） */
export const deleteSession = (id: number) => http.delete(`/chat/sessions/${id}`)

/** SSE 事件统一处理器 */
export interface StreamHandlers {
  onDelta: (content: string) => void
  /** 契约 7.3：思考过程增量（无思考模型时不发该事件） */
  onReasoning: (content: string) => void
  onUsage: (usage: {
    prompt_tokens: number
    completion_tokens: number
    cached_tokens: number
    total_tokens: number
    cost: number
  }) => void
  onDone: (done: { session_id: number; message_id: number }) => void
  onError: (msg: string) => void
}

/** 解析单条 SSE data 行并分发事件 */
function dispatchLine(raw: string, handlers: StreamHandlers) {
  const line = raw.trim()
  if (!line || !line.startsWith('data:')) return
  const payload = line.slice(5).trim()
  if (!payload || payload === '[DONE]') return
  try {
    const evt = JSON.parse(payload)
    switch (evt.type) {
      case 'delta':
        handlers.onDelta(evt.content || '')
        break
      case 'reasoning':
        handlers.onReasoning(evt.content || '')
        break
      case 'usage':
        handlers.onUsage({
          prompt_tokens: evt.prompt_tokens || 0,
          completion_tokens: evt.completion_tokens || 0,
          cached_tokens: evt.cached_tokens || 0,
          total_tokens: evt.total_tokens || 0,
          cost: Number(evt.cost || 0)
        })
        break
      case 'done':
        handlers.onDone({ session_id: evt.session_id, message_id: evt.message_id })
        break
      case 'error':
        handlers.onError(evt.msg || '上游模型返回错误')
        break
      default:
        break
    }
  } catch {
    // 忽略无法解析的残行
  }
}

/**
 * POST /api/chat/completions（恒 stream=true）
 * 逐行解析 `data: {...}`；兼容后端非 SSE 降级补发的 delta/usage/done 序列；
 * 读取结束后兜底解析无换行结尾的残行，保证状态收尾、不卡死 loading。
 * signal 用于「停止生成」：中断读取，已收到的增量内容保留。
 */
export async function streamCompletions(payload: CompletionPayload, handlers: StreamHandlers, signal?: AbortSignal) {
  const token = localStorage.getItem('ak_token') || ''
  const resp = await fetch('/api/chat/completions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify(payload),
    signal
  })

  if (!resp.ok || !resp.body) {
    let errText = ''
    try {
      errText = await resp.text()
    } catch {
      /* 忽略读取异常 */
    }
    let serverMsg = `请求失败（HTTP ${resp.status}）`
    try {
      const parsed = JSON.parse(errText)
      serverMsg = parsed?.msg || serverMsg
    } catch {
      /* 非 JSON 错误体 */
    }
    handlers.onError(serverMsg)
    return
  }

  const reader = resp.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      // 兼容 \r\n 行尾（契约 8.2：统一替换为 \n 再切行）
      buffer = buffer.replace(/\r\n/g, '\n')
      let idx = buffer.indexOf('\n')
      while (idx >= 0) {
        dispatchLine(buffer.slice(0, idx), handlers)
        buffer = buffer.slice(idx + 1)
        idx = buffer.indexOf('\n')
      }
    }
    // 兜底：最后一行可能没有换行符
    dispatchLine(buffer, handlers)
  } catch (e) {
    // 主动停止（AbortError）不算错误：保留已收到的增量，正常收尾
    if ((e as Error)?.name !== 'AbortError') throw e
  }
}
