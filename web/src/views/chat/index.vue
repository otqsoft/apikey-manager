<template>
  <div class="chat-page">
    <!-- 左侧会话列表 -->
    <aside class="session-panel">
      <div class="session-header">
        <span>会话列表</span>
        <el-button size="small" type="primary" plain circle :icon="Plus" :disabled="streaming" @click="newSession"></el-button>
      </div>
      <el-scrollbar class="session-list">
        <div
          v-for="s in sessions"
          :key="s.id"
          class="session-item"
          :class="{ active: s.id === activeSessionId }"
          @click="selectSession(s)"
        >
          <div class="session-title">{{ s.title || `${s.model} #${s.id}` }}</div>
          <div class="session-meta">{{ formatDateTime(s.created_at) }}</div>
          <el-icon class="session-del" @click.stop="removeSession(s)"><Delete /></el-icon>
        </div>
        <el-empty v-if="!sessions.length" description="暂无会话" :image-size="60" />
      </el-scrollbar>
    </aside>

    <!-- 右侧聊天区 -->
    <section class="chat-main">
      <!-- 顶部选择栏 -->
      <div class="chat-toolbar">
        <el-select
          v-model="selectedKeyId"
          placeholder="选择我的可用 Key"
          filterable
          class="key-select"
          :disabled="streaming"
          @change="onKeyChange"
        >
          <el-option-group v-for="g in groupedKeys" :key="g.name" :label="g.name">
            <el-option
              v-for="k in g.keys"
              :key="k.id"
              :label="`${k.name}（${k.provider.name}）`"
              :value="k.id"
              :disabled="k.usable === false"
            >
              <div class="key-option">
                <span class="key-option-name">{{ k.name }}（{{ k.provider.name }}）</span>
                <span v-if="k.usable === false" class="key-option-reason">{{ k.unusable_reason || '不可用' }}</span>
              </div>
            </el-option>
          </el-option-group>
        </el-select>
        <el-select v-model="selectedModel" placeholder="选择模型" class="model-select" :disabled="streaming">
          <el-option v-for="m in modelOptions" :key="m.id" :label="m.name" :value="m.name" />
        </el-select>
        <div class="cost-badge">
          <el-tag type="success" size="large">累计费用 {{ formatMoney(totalCost) }}</el-tag>
          <el-tag type="info" size="large" style="margin-left: 8px">累计 Token {{ formatInt(totalTokens) }}</el-tag>
        </div>
      </div>

      <!-- 消息区 -->
      <div ref="messageBox" class="message-box">
        <el-empty v-if="!messages.length" description="选择 Key 与模型后开始对话" :image-size="90" />
        <div v-for="(m, idx) in messages" :key="idx" class="msg-row" :class="m.role">
          <span v-if="m.role === 'assistant'" class="bubble-avatar ai">AI</span>
          <div class="bubble">
            <!-- assistant 头部：供应商 · 模型；user 消息不显示 -->
            <div v-if="m.role === 'assistant'" class="msg-meta">
              {{ m.provider_name || '未知供应商' }} · {{ m.model || '—' }}
            </div>
            <!-- 思考过程折叠区：无 reasoning 内容时完全不渲染 -->
            <div v-if="m.role === 'assistant' && m.reasoning" class="reasoning-block">
              <div class="reasoning-head" @click="m.reasoningOpen = !m.reasoningOpen">
                <el-icon class="reasoning-arrow" :class="{ open: m.reasoningOpen }">
                  <ArrowRight />
                </el-icon>
                <span>思考过程</span>
                <span class="reasoning-count">{{ (m.reasoning || '').length }} 字</span>
              </div>
              <div v-show="m.reasoningOpen" class="reasoning-body">{{ m.reasoning }}</div>
            </div>
            <div class="bubble-content">
              <!-- user 消息：保持纯文本 pre-wrap，不做 Markdown 渲染 -->
              <span v-if="m.role === 'user'" class="user-text" v-text="m.content" />
              <!-- assistant：Markdown 解析 + DOMPurify 清洗后 v-html；历史回放复用同一函数 -->
              <template v-else>
                <div v-if="m.content" class="md-body" v-html="renderMarkdown(m.content)" />
                <span
                  v-else-if="streaming && idx === messages.length - 1 && !m.reasoning"
                  class="thinking-text"
                  >正在思考…</span
                >
                <span v-if="streaming && idx === messages.length - 1 && m.content" class="caret" />
              </template>
            </div>
            <div v-if="m.role === 'assistant' && (m.total_tokens || m.cost)" class="usage-line">
              <span>输入 {{ formatInt(m.prompt_tokens) }}（缓存 {{ formatInt(m.cached_tokens) }}）</span>
              <span>输出 {{ formatInt(m.completion_tokens) }}</span>
              <span>合计 {{ formatInt(m.total_tokens) }}</span>
              <span class="usage-cost">费用 {{ formatMoney(m.cost) }}</span>
            </div>
            <div v-if="m.created_at" class="msg-time">{{ formatDateTime(m.created_at) }}</div>
          </div>
          <span v-if="m.role === 'user'" class="bubble-avatar user">我</span>
        </div>
      </div>

      <!-- 输入区：一体化胶囊容器，发送按钮内嵌右下角 -->
      <div class="input-area">
        <div class="input-shell" :class="{ focused: inputFocused }">
          <textarea
            ref="inputEl"
            v-model="inputText"
            class="input-textarea"
            rows="3"
            placeholder="输入问题…"
            :disabled="streaming"
            @focus="inputFocused = true"
            @blur="inputFocused = false"
            @keydown.enter.exact.prevent="onSend"
          />
          <div class="input-actions">
            <span class="input-hint">Enter 发送 · Shift + Enter 换行</span>
            <button v-if="streaming" class="send-btn stop" title="停止生成" @click="onStop">
              <span class="stop-icon" />
            </button>
            <button v-else class="send-btn" :disabled="!inputText.trim()" title="发送" @click="onSend">
              <el-icon :size="17"><Promotion /></el-icon>
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowRight, Delete, Plus, Promotion } from '@element-plus/icons-vue'
import { createSession, deleteSession, getSessionMessages, getSessionPage, streamCompletions } from '@/api/chat'
import { getMyKeys } from '@/api/keys'
import type { ChatMessage, ChatSession, MyKeyItem } from '@/types'
import { formatDateTime, formatInt, formatMoney } from '@/utils/format'
import { renderMarkdown } from '@/utils/markdown'

// ---------------- 可用 Key 与模型 ----------------
const myKeys = ref<MyKeyItem[]>([])
const selectedKeyId = ref<number | null>(null)
const selectedModel = ref('')

const modelOptions = computed(() => myKeys.value.find((k) => k.id === selectedKeyId.value)?.models || [])

/** Key 下拉树形分组：一级供应商，二级该供应商下的 Key 列表 */
const groupedKeys = computed(() => {
  const map = new Map<string, { name: string; keys: MyKeyItem[] }>()
  for (const k of myKeys.value) {
    const name = k.provider?.name || '未知供应商'
    if (!map.has(name)) map.set(name, { name, keys: [] })
    map.get(name)!.keys.push(k)
  }
  return [...map.values()]
})

async function loadMyKeys() {
  const res = await getMyKeys()
  myKeys.value = res || []
  // 默认选中第一个可用 Key（不可用项置灰不可选）
  if (!selectedKeyId.value) {
    const firstUsable = myKeys.value.find((k) => k.usable !== false)
    if (firstUsable) {
      selectedKeyId.value = firstUsable.id
      onKeyChange(firstUsable.id)
    }
  }
}

function onKeyChange(keyId: number | null) {
  if (!keyId) {
    selectedModel.value = ''
    return
  }
  const k = myKeys.value.find((x) => x.id === keyId)
  selectedModel.value = k?.models?.[0]?.name || ''
}

// ---------------- 会话与消息 ----------------
const sessions = ref<ChatSession[]>([])
const activeSessionId = ref<number | null>(null)
const messages = ref<ChatMessage[]>([])
const inputText = ref('')
const inputFocused = ref(false)
const inputEl = ref<HTMLTextAreaElement>()
const streaming = ref(false)
const abortRef = ref<AbortController | null>(null)
const totalCost = ref(0)
const messageBox = ref<HTMLElement>()

const totalTokens = computed(() => messages.value.reduce((sum, m) => sum + Number(m.total_tokens || 0), 0))

async function loadSessions() {
  const res = await getSessionPage({ page: 1, page_size: 100 })
  sessions.value = res.list || []
}

async function selectSession(s: ChatSession) {
  if (streaming.value) return
  activeSessionId.value = s.id
  const list = await getSessionMessages(s.id)
  messages.value = list || []
  totalCost.value = messages.value.reduce((sum, m) => sum + Number(m.cost || 0), 0)
  scrollToBottom()
}

function newSession() {
  if (streaming.value) return
  activeSessionId.value = null
  messages.value = []
  totalCost.value = 0
}

async function removeSession(s: ChatSession) {
  await ElMessageBox.confirm(`确定删除会话「${s.title || s.model}」吗？`, '删除确认', { type: 'warning' })
  await deleteSession(s.id)
  ElMessage.success('会话已删除')
  if (activeSessionId.value === s.id) newSession()
  loadSessions()
}

function scrollToBottom() {
  nextTick(() => {
    const el = messageBox.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

// ---------------- 发送与流式接收 ----------------
async function onSend() {
  const content = inputText.value.trim()
  if (!content) return
  if (streaming.value) return
  if (!selectedKeyId.value) {
    ElMessage.warning('请先选择一个可用 Key')
    return
  }
  if (!selectedModel.value) {
    ElMessage.warning('该供应商下暂无可用模型')
    return
  }

  // 组装对话历史（user / assistant，不含占位态）
  const history = messages.value
    .filter((m) => m.role !== 'system')
    .map((m) => ({ role: m.role, content: m.content }))
  history.push({ role: 'user', content })

  messages.value.push({
    role: 'user',
    content,
    prompt_tokens: 0,
    completion_tokens: 0,
    cached_tokens: 0,
    total_tokens: 0,
    cost: 0
  })
  inputText.value = ''
  scrollToBottom()

  // 首次发送先落会话
  let sessionId = activeSessionId.value
  if (!sessionId) {
    const session = await createSession({ key_id: selectedKeyId.value, model: selectedModel.value })
    sessionId = session.id
    activeSessionId.value = session.id
    await loadSessions()
  }

  // assistant 占位消息：实时对话用当前选中 Key 的供应商名 + 选中模型兜底头部标签
  const currentKey = myKeys.value.find((k) => k.id === selectedKeyId.value)
  messages.value.push({
    role: 'assistant',
    content: '',
    reasoning: '',
    provider_name: currentKey?.provider?.name || '',
    model: selectedModel.value,
    reasoningOpen: true, // 流式思考阶段展开
    prompt_tokens: 0,
    completion_tokens: 0,
    cached_tokens: 0,
    total_tokens: 0,
    cost: 0
  })
  // 关键：必须通过响应式代理引用修改字段。若直接持有原始对象引用做 +=，
  // 变更不触发重渲染 → 流式期间界面静止、结束时一次性全部显示（Vue3 reactive 陷阱）
  const assistant = messages.value[messages.value.length - 1]
  streaming.value = true
  const ac = new AbortController()
  abortRef.value = ac

  try {
    await streamCompletions(
      {
        key_id: selectedKeyId.value,
        model: selectedModel.value,
        session_id: sessionId,
        stream: true, // 恒为流式；后端对不支持流式的上游会降级补发 delta/usage/done
        messages: history
      },
      {
        onDelta: (text) => {
          assistant.content += text
          scrollToBottom()
        },
        // 思考过程独立累加，不混入正文
        onReasoning: (text) => {
          assistant.reasoning += text
          assistant.reasoningOpen = true
          scrollToBottom()
        },
        onUsage: (u) => {
          const prevCost = Number(assistant.cost || 0)
          Object.assign(assistant, u)
          totalCost.value += Number(u.cost || 0) - prevCost
        },
        onDone: async () => {
          // 完成后自动折叠思考区；会话列表刷新失败不阻塞消息展示
          assistant.reasoningOpen = false
          try {
            await loadSessions()
          } catch {
            /* 静默兜底 */
          }
        },
        onError: (msg) => {
          ElMessage.error(msg || '上游模型返回错误')
          if (!assistant.content) assistant.content = '（生成失败，请稍后重试）'
        }
      },
      ac.signal
    )
  } finally {
    streaming.value = false
    abortRef.value = null
    scrollToBottom()
    inputEl.value?.focus()
  }
}

/** 停止生成：中断读取，已收到的增量内容保留展示 */
function onStop() {
  abortRef.value?.abort()
}

onBeforeUnmount(() => {
  abortRef.value?.abort()
})

onMounted(async () => {
  await Promise.all([loadMyKeys(), loadSessions()])
  inputEl.value?.focus()
})
</script>

<style scoped>
.chat-page {
  display: flex;
  height: calc(100vh - 72px);
  background: transparent;
  border-radius: 16px;
  overflow: hidden;
}
.session-panel {
  width: 260px;
  border-right: 1px solid #fff;
  background: #fff;
  display: flex;
  flex-direction: column;
}
.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 14px 10px;
  font-weight: 600;
  font-size: 13px;
  color: #334155;
}
.session-list {
  flex: 1;
  padding: 0 8px 8px;
}
.session-item {
  position: relative;
  padding: 10px 28px 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  margin-bottom: 4px;
  background: transparent;
  transition: background 0.15s ease;
}
.session-item:hover {
  background: #eef3fa;
}
.session-item.active {
  background: #e3edfb;
}
.session-title {
  font-size: 13px;
  color: #1f2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.session-meta {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 3px;
}
.session-del {
  position: absolute;
  right: 8px;
  top: 12px;
  color: #c0c4cc;
  opacity: 0;
  transition: opacity 0.15s ease;
}
.session-item:hover .session-del {
  opacity: 1;
}
.session-del:hover {
  color: #f56c6c;
}
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.chat-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 18px;
  border-bottom: 1px solid #e8eef6;
  background: #fff;
}
.key-select {
  width: 260px;
}
/* Key 下拉项：名称 + 不可用原因 */
.key-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
}
.key-option-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.key-option-reason {
  flex-shrink: 0;
  font-size: 12px;
  color: #f56c6c;
}
.model-select {
  width: 200px;
}
.cost-badge {
  margin-left: auto;
}
.message-box {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  background: #f4f7fc;
}
.msg-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 18px;
}
.msg-row.user {
  justify-content: flex-end;
}
.bubble-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: #fff;
}
.bubble-avatar.ai {
  background: #2563eb;
}
.bubble-avatar.user {
  background: #94a3b8;
}
/* assistant 气泡占满除头像外的可用宽度（白底圆角，无边框无阴影） */
.bubble {
  flex: 1;
  min-width: 0;
  padding: 10px 14px 12px;
  border-radius: 12px;
  background: #fff;
}
/* user 气泡：浅蓝底、宽度随文本内容自适应（上限 72%），不拉伸占满 */
.msg-row.user {
  justify-content: flex-end;
}
.msg-row.user .bubble {
  flex: 0 0 auto;
  width: fit-content;
  max-width: 72%;
  background: #d6e6ff;
  color: #1f2937;
}
.bubble-content {
  line-height: 1.65;
  font-size: 14px;
  min-width: 0;
}
.user-text {
  white-space: pre-wrap;
  word-break: break-word;
}
/* assistant 头部：供应商 · 模型 */
.msg-meta {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 6px;
}
/* 思考过程折叠区：浅灰底与正文区分 */
.reasoning-block {
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  margin-bottom: 8px;
}
.reasoning-head {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  font-size: 12px;
  color: #64748b;
  cursor: pointer;
  user-select: none;
}
.reasoning-arrow {
  transition: transform 0.2s ease;
  font-size: 12px;
}
.reasoning-arrow.open {
  transform: rotate(90deg);
}
.reasoning-count {
  margin-left: auto;
  color: #94a3b8;
}
.reasoning-body {
  padding: 4px 10px 10px;
  font-size: 13px;
  color: #475569;
  line-height: 1.6;
  white-space: pre-wrap;
  border-top: 1px dashed #e2e8f0;
  max-height: 180px;
  overflow-y: auto;
}
.thinking-text {
  color: #94a3b8;
  font-size: 13px;
}
/* 流式输出末尾呼吸光标 */
.caret {
  display: inline-block;
  width: 7px;
  height: 14px;
  margin-left: 3px;
  background: #2563eb;
  border-radius: 2px;
  vertical-align: middle;
  animation: caret-blink 0.9s ease-in-out infinite;
}
@keyframes caret-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.15; }
}

/* ---------- Markdown 排版（assistant 正文 v-html，契约 8.3；本轮无语法高亮） ---------- */
.md-body {
  min-width: 0;
  word-break: break-word;
}
.md-body :deep(h1),
.md-body :deep(h2),
.md-body :deep(h3),
.md-body :deep(h4) {
  margin: 12px 0 8px;
  font-weight: 700;
  line-height: 1.4;
  color: #1f2937;
}
.md-body :deep(h1) { font-size: 18px; }
.md-body :deep(h2) { font-size: 16px; }
.md-body :deep(h3) { font-size: 15px; }
.md-body :deep(h4) { font-size: 14px; }
.md-body :deep(p) {
  margin: 6px 0;
  line-height: 1.7;
}
.md-body :deep(ul),
.md-body :deep(ol) {
  margin: 6px 0;
  padding-left: 22px;
}
.md-body :deep(li) { margin: 2px 0; }
.md-body :deep(blockquote) {
  margin: 8px 0;
  padding: 6px 12px;
  border-left: 3px solid #93c5fd;
  background: #f1f5f9;
  color: #475569;
  border-radius: 0 6px 6px 0;
}
.md-body :deep(code) {
  background: #f1f5f9;
  color: #dc2626;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
/* 代码块：深色 #0f172a 等宽 + 横向滚动，不撑破布局 */
.md-body :deep(pre) {
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 8px;
  padding: 12px 14px;
  margin: 8px 0;
  overflow-x: auto;
}
.md-body :deep(pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
  font-size: 13px;
  line-height: 1.6;
}
.md-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
  font-size: 13px;
}
.md-body :deep(th),
.md-body :deep(td) {
  border: 1px solid #e2e8f0;
  padding: 6px 10px;
}
.md-body :deep(th) {
  background: #f1f5f9;
  font-weight: 600;
}
.md-body :deep(tr:nth-child(even) td) { background: #f8fafc; }
.md-body :deep(a) {
  color: #2563eb;
  text-decoration: none;
}
.md-body :deep(a:hover) { text-decoration: underline; }
.md-body :deep(hr) {
  border: none;
  border-top: 1px solid #e2e8f0;
  margin: 10px 0;
}
.md-body :deep(img) {
  max-width: 100%;
  border-radius: 8px;
}
.md-body :deep(> :first-child) { margin-top: 0; }
.md-body :deep(> :last-child) { margin-bottom: 0; }

.usage-line {
  margin-top: 8px;
  font-size: 12px;
  color: #64748b;
  border-top: 1px dashed #e2e8f0;
  padding-top: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px 14px;
  font-variant-numeric: tabular-nums;
}
.usage-cost {
  color: #2563eb;
  font-weight: 600;
}
.msg-time {
  margin-top: 6px;
  font-size: 11px;
  color: #94a3b8;
  text-align: right;
}
/* ---------- 一体化输入区：容器即输入框，发送按钮内嵌右下角 ---------- */
.input-area {
  padding: 6px 18px 16px;
  background: #f4f7fc;
}
.input-shell {
  max-width: 860px;
  margin: 0 auto;
  border: 1px solid #d9e2ee;
  border-radius: 16px;
  background: #fff;
  padding: 10px 12px 8px;
  transition: border-color 0.15s ease;
}
.input-shell.focused {
  border-color: #3b82f6;
}
.input-textarea {
  display: block;
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  background: transparent;
  font-size: 14px;
  line-height: 1.6;
  color: #1f2937;
  font-family: inherit;
}
.input-textarea::placeholder {
  color: #a8b3c5;
}
.input-textarea:disabled {
  color: #94a3b8;
}
.input-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
}
.input-hint {
  font-size: 11px;
  color: #b0bacb;
  user-select: none;
}
.send-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #2563eb;
  color: #fff;
  transition: background 0.15s ease, transform 0.1s ease;
}
.send-btn:hover:not(:disabled) {
  background: #1d4fc0;
}
.send-btn:active:not(:disabled) {
  transform: scale(0.94);
}
.send-btn:disabled {
  background: #cbd5e1;
  cursor: not-allowed;
}
/* 停止按钮：白底方块图标 */
.send-btn.stop {
  background: #fff;
  border: 1px solid #cbd5e1;
  color: #334155;
}
.send-btn.stop:hover {
  border-color: #f56c6c;
  color: #f56c6c;
}
.stop-icon {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  background: currentColor;
}
@media (prefers-reduced-motion: reduce) {
  .caret {
    animation: none;
  }
  .send-btn,
  .session-item,
  .input-shell {
    transition: none;
  }
}
</style>
