<template>
  <div class="chat-page">
    <header class="chat-header">
      <div class="action-row">
        <button @click="$router.push('/teacher/knowledge/chats')">返回列表</button>
        <h2>会话 {{ sessionId }}</h2>
      </div>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>

    <section class="chat-layout">
      <article ref="messageContainerRef" class="card chat-messages">
        <div v-if="messages.length === 0" class="subtle">开始提问吧</div>
        <div v-for="m in messages" :key="String(m.id)" class="message-row" :class="m.role === 'user' ? 'user' : 'assistant'">
          <div class="message-meta">
            <span>{{ m.role === 'user' ? '你' : '助手' }}</span>
            <span>{{ formatTime(m.created_at) }}</span>
            <span v-if="m.isStreaming" class="typing-flag">正在输入...</span>
          </div>
          <div class="message-bubble">
            <template v-if="m.role === 'assistant'">
              <div class="msg-content markdown-body" v-html="renderAssistantContent(m.content)" />
              <span v-if="m.isStreaming" class="cursor">▋</span>
            </template>
            <pre v-else class="msg-content">{{ m.content }}</pre>
          </div>
          <section v-if="m.references_json && m.references_json.length > 0" class="refs">
            <div class="refs-title">引用内容（{{ m.references_json.length }}）</div>
            <article v-for="(r, idx) in m.references_json" :key="`${m.id}-${idx}`" class="ref-card">
              <button class="ref-head" @click="toggleRef(String(m.id), idx)">
                <span class="ref-chevron">{{ isRefExpanded(String(m.id), idx) ? '▾' : '▸' }}</span>
                <span class="ref-text">[{{ r.citation_no || idx + 1 }}] {{ r.resource_title }} #{{ (r.chunk_index || 0) + 1 }}</span>
              </button>
              <div v-if="isRefExpanded(String(m.id), idx)" class="ref-body">
                <p>{{ r.snippet || '（无摘要片段）' }}</p>
                <button v-if="r.full_content" class="link-btn" @click="openRef(r)">查看完整内容</button>
              </div>
            </article>
          </section>
        </div>
      </article>

      <aside class="chat-side">
        <label>
          TopK
          <input v-model.number="topK" type="number" min="1" max="20" />
        </label>
        <label>
          语义阈值
          <input v-model.number="semanticMinScore" type="number" min="0" max="1" step="0.01" />
        </label>
        <label>
          关键词阈值
          <input v-model.number="keywordMinScore" type="number" min="0" max="1" step="0.01" />
        </label>
        <label>
          问答模型
          <select v-model="qaModelId">
            <option value="">默认模型</option>
            <option v-for="m in qaModels" :key="m.id" :value="String(m.id)">
              {{ m.name }} ({{ m.model_id }})
            </option>
          </select>
        </label>
        <button @click="loadMessages">刷新历史</button>
      </aside>
    </section>

    <footer class="chat-input">
      <textarea
        v-model="question"
        placeholder="输入问题（Enter 发送，Shift+Enter 换行）"
        :disabled="asking"
        @keydown="handleKeyDown"
      />
      <button class="primary" :disabled="asking || !question.trim()" @click="handleAsk">
        {{ asking ? '思考中...' : '发送' }}
      </button>
    </footer>

    <div v-if="refViewer" class="modal-mask">
      <div class="modal-panel card stack">
        <h3>{{ refViewer.title }}</h3>
        <pre class="mono ref-content">{{ refViewer.content }}</pre>
        <div class="action-row">
          <button @click="refViewer = null">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'
import { useAuthStore } from '@/stores/auth'
import { listAIModels, type AIModelItem } from '@/services/aiModels'
import {
  listChatMessages,
  askInSessionStream,
  type ChatMessageItem,
} from '@/services/knowledgeChat'

type MessageRow = ChatMessageItem & { isStreaming?: boolean }

const route = useRoute()
const auth = useAuthStore()
const sessionId = String(route.params.sessionId || '')

const messages = ref<MessageRow[]>([])
const question = ref('')
const asking = ref(false)
const error = ref('')
const qaModels = ref<AIModelItem[]>([])
const qaModelId = ref('')

const topK = ref(8)
const semanticMinScore = ref(0)
const keywordMinScore = ref(0)
const messageContainerRef = ref<HTMLElement | null>(null)
const refViewer = ref<{ title: string; content: string } | null>(null)
const md = new MarkdownIt({ linkify: true, breaks: true, html: false })
const expandedRefKeys = ref<Set<string>>(new Set())

function formatTime(value: string) {
  return value ? new Date(value).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) : '-'
}

// 将模型输出的 Markdown 转为安全 HTML，避免直接注入不可信内容。
function renderAssistantContent(content: string) {
  const html = md.render(content || '')
  return DOMPurify.sanitize(html)
}

function openRef(reference: {
  citation_no?: number
  resource_title?: string
  snippet?: string
  full_content?: string
}) {
  refViewer.value = {
    title: `[${reference.citation_no || 1}] ${reference.resource_title || '引用内容'}`,
    content: reference.full_content || reference.snippet || '',
  }
}

function makeRefKey(messageId: string, idx: number) {
  return `${messageId}-${idx}`
}

function isRefExpanded(messageId: string, idx: number) {
  return expandedRefKeys.value.has(makeRefKey(messageId, idx))
}

// 引用块按消息+索引独立折叠，避免一次展开影响其他消息。
function toggleRef(messageId: string, idx: number) {
  const key = makeRefKey(messageId, idx)
  const next = new Set(expandedRefKeys.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedRefKeys.value = next
}

// 拉取会话历史并归一化引用字段，便于统一渲染。
async function loadMessages() {
  if (!sessionId) return
  error.value = ''
  try {
    const out = await listChatMessages(auth.token ?? '', sessionId, 1, 200)
    messages.value = (out.items || []).map((m) => ({
      ...m,
      references_json: m.references_json || [],
    }))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载历史失败'
  }
}

// 加载 QA 模型供会话参数选择，默认选中第一项。
async function loadQAModels() {
  try {
    const out = await listAIModels(auth.token ?? '')
    qaModels.value = (out.items || []).filter((m) => m.model_type === 'qa')
    if (!qaModelId.value && qaModels.value.length > 0) {
      qaModelId.value = String(qaModels.value[0].id)
    }
  } catch {
    // 模型加载失败不阻塞问答流程。
  }
}

// 发送提问并消费 SSE，实时拼接 assistant 内容。
async function handleAsk() {
  if (!question.value.trim() || asking.value || !sessionId) return
  asking.value = true
  error.value = ''
  const askText = question.value.trim()
  question.value = ''
  const tmpUserId = `tmp-user-${Date.now()}`
  const tmpAssistantId = `tmp-assistant-${Date.now()}`

  messages.value.push(
    {
      id: tmpUserId,
      session_id: sessionId,
      role: 'user',
      content: askText,
      created_at: new Date().toISOString(),
      references_json: [],
    },
    {
      id: tmpAssistantId,
      session_id: sessionId,
      role: 'assistant',
      content: '',
      created_at: new Date().toISOString(),
      references_json: [],
      isStreaming: true,
    },
  )

  try {
    await askInSessionStream(
      auth.token ?? '',
      sessionId,
      askText,
      topK.value,
      qaModelId.value || undefined,
      semanticMinScore.value,
      keywordMinScore.value,
      {
        onToken: (token) => {
          messages.value = messages.value.map((m) =>
            String(m.id) === tmpAssistantId ? { ...m, content: m.content + token } : m,
          )
        },
        onReferences: (refs) => {
          messages.value = messages.value.map((m) =>
            String(m.id) === tmpAssistantId ? { ...m, references_json: refs } : m,
          )
        },
        onDone: () => {
          messages.value = messages.value.map((m) =>
            String(m.id) === tmpAssistantId ? { ...m, isStreaming: false } : m,
          )
        },
        onError: (message) => {
          error.value = message
        },
      },
    )
    await loadMessages()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '提问失败'
    messages.value = messages.value.filter((m) => String(m.id) !== tmpUserId && String(m.id) !== tmpAssistantId)
  } finally {
    asking.value = false
  }
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    void handleAsk()
  }
}

watch(
  messages,
  async () => {
    await nextTick()
    if (messageContainerRef.value) {
      messageContainerRef.value.scrollTop = messageContainerRef.value.scrollHeight
    }
  },
  { deep: true },
)

onMounted(async () => {
  await Promise.all([loadQAModels(), loadMessages()])
})
</script>

<style scoped>
.chat-page {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
}
.chat-header {
  flex: 0 0 auto;
  border-bottom: 1px solid #dbe4f0;
  background: #fff;
  padding: 10px 16px;
}

.chat-header h2 {
  font-size: 18px;
}
.chat-layout {
  min-height: 0;
  flex: 1 1 auto;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 0;
  overflow: hidden;
}
.chat-messages {
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  border: none;
  border-radius: 0;
  background: #f8fafc;
}
.message-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.message-row.user {
  align-items: flex-end;
}
.message-row.assistant {
  align-items: flex-start;
}
.message-meta {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #64748b;
  align-items: center;
}
.typing-flag {
  color: #2563eb;
}
.message-bubble {
  max-width: 82%;
  border-radius: 14px;
  padding: 10px 12px;
  background: #fff;
  border: 1px solid #e5e7eb;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
}
.message-row.user .message-bubble {
  background: #eff6ff;
  color: #0f172a;
  border-color: #bfdbfe;
}
.msg-content {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
}
.markdown-body :deep(p) {
  margin: 0 0 10px;
}
.markdown-body :deep(p:last-child) {
  margin-bottom: 0;
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 0 0 10px;
  padding-left: 20px;
}
.markdown-body :deep(pre) {
  background: #0f172a;
  color: #e2e8f0;
  padding: 10px;
  border-radius: 8px;
  overflow: auto;
}
.markdown-body :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
.markdown-body :deep(a) {
  color: #2563eb;
  text-decoration: underline;
}
.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #cbd5e1;
  padding: 6px 8px;
}
.cursor {
  animation: blink 1s steps(1) infinite;
}
.refs {
  max-width: 82%;
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.refs-title {
  font-size: 12px;
  color: #64748b;
  font-weight: 600;
}
.ref-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  background: #f8fafc;
}
.ref-head {
  width: 100%;
  border: 0;
  background: transparent;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  text-align: left;
  cursor: pointer;
}
.ref-head:hover {
  background: #f1f5f9;
}
.ref-chevron {
  color: #64748b;
}
.ref-text {
  font-size: 13px;
  color: #0f172a;
}
.ref-body {
  border-top: 1px solid #dbe4f0;
  padding: 8px 10px;
}
.ref-body p {
  margin: 0 0 6px;
  font-size: 13px;
  color: #334155;
  white-space: pre-wrap;
}
.link-btn {
  border: 0;
  background: transparent;
  color: #2563eb;
  cursor: pointer;
  padding: 0;
}
.chat-side {
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  border-left: 1px solid #dbe4f0;
  padding: 12px 12px 12px 14px;
  background: #fff;
}

.chat-side > label {
  font-size: 13px;
  color: #334155;
}

.chat-side > button {
  margin-top: 4px;
}
.chat-input {
  flex: 0 0 auto;
  display: flex;
  gap: 12px;
  align-items: flex-end;
  border-top: 1px solid #dbe4f0;
  border-radius: 0;
  padding: 8px 10px;
  background: #fff;
}
.chat-input textarea {
  min-height: 60px;
  max-height: 200px;
  border: none;
  box-shadow: none;
  padding: 8px 6px;
}
.chat-input textarea:focus {
  outline: none;
  box-shadow: none;
}
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  display: grid;
  place-items: center;
  z-index: 40;
}
.modal-panel {
  width: min(820px, calc(100vw - 32px));
}
.ref-content {
  max-height: 55vh;
  overflow: auto;
  white-space: pre-wrap;
}
@keyframes blink {
  50% {
    opacity: 0;
  }
}
@media (max-width: 960px) {
  .chat-page {
    height: auto;
    min-height: 100%;
  }
  .chat-layout {
    grid-template-columns: 1fr;
  }

  .chat-side {
    border-left: none;
    border-top: 1px solid #dbe4f0;
  }
}
</style>
