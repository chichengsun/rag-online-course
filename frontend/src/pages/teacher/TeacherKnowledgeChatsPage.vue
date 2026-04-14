<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>对话管理</h2>
      <p class="subtle">管理知识库会话，点击进入详情查看历史并继续问答。</p>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <section class="card stack">
      <div class="list-toolbar">
        <p class="subtle">共 {{ sessions.length }} 个会话</p>
        <button class="primary" :disabled="creating" @click="showCreate = true">新建对话</button>
      </div>

      <div v-if="sessions.length === 0" class="subtle empty-tips">暂无会话</div>
      <div v-else class="chat-grid">
        <article v-for="s in sessions" :key="s.id" class="course-item">
          <div>
            <h4>{{ s.title || '未命名会话' }}</h4>
            <p class="mono">课程ID：{{ s.course_id }}</p>
            <p class="subtle">消息数：{{ s.message_count }}</p>
            <p class="subtle">创建时间：{{ formatDate(s.created_at, false) }}</p>
            <p class="subtle">更新时间：{{ formatDate(s.updated_at, true) }}</p>
          </div>
          <div class="action-row">
            <button class="primary" @click="$router.push(`/teacher/knowledge/chats/${s.id}`)">进入对话</button>
            <button class="danger" @click="openDelete(s)">删除</button>
          </div>
        </article>
      </div>
    </section>

    <div v-if="showCreate" class="modal-mask">
      <div class="modal-panel card stack">
        <h3>新建会话</h3>
        <label>
          选择课程
          <select v-model="selectedCourseId">
            <option value="" disabled>请选择课程</option>
            <option v-for="c in courses" :key="c.id" :value="String(c.id)">{{ c.title }}</option>
          </select>
        </label>
        <label>
          会话标题
          <input v-model="newSessionTitle" type="text" placeholder="可选，默认新对话" />
        </label>
        <div class="action-row">
          <button @click="showCreate = false">取消</button>
          <button class="primary" :disabled="creating || !selectedCourseId" @click="handleCreate">创建</button>
        </div>
      </div>
    </div>

    <div v-if="showDelete" class="modal-mask">
      <div class="modal-panel card stack">
        <h3>确认删除</h3>
        <p class="subtle">确定删除会话「{{ deletingSession?.title || '未命名会话' }}」吗？</p>
        <div class="action-row">
          <button @click="showDelete = false">取消</button>
          <button class="danger" :disabled="deleting" @click="handleDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { listTeacherCourses, type TeacherCourseItem } from '@/services/course'
import {
  listChatSessions,
  createChatSession,
  deleteChatSession,
  type ChatSessionItem,
} from '@/services/knowledgeChat'

const router = useRouter()
const auth = useAuthStore()

const sessions = ref<ChatSessionItem[]>([])
const courses = ref<TeacherCourseItem[]>([])
const selectedCourseId = ref('')
const newSessionTitle = ref('')

const showCreate = ref(false)
const creating = ref(false)
const showDelete = ref(false)
const deleting = ref(false)
const deletingSession = ref<ChatSessionItem | null>(null)
const error = ref('')
const success = ref('')
let successTimer: ReturnType<typeof setTimeout> | null = null

function formatDate(value: string, withTime = true) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    ...(withTime ? { hour: '2-digit', minute: '2-digit' } : {}),
  })
}

function showSuccess(message: string) {
  success.value = message
  if (successTimer) clearTimeout(successTimer)
  successTimer = setTimeout(() => {
    success.value = ''
    successTimer = null
  }, 3000)
}

// 加载课程与会话列表，供新建会话和列表展示使用。
async function load() {
  error.value = ''
  try {
    const [courseData, sessionData] = await Promise.all([
      listTeacherCourses(auth.token ?? '', 1, 100, '', '', 'updated_at', 'desc'),
      listChatSessions(auth.token ?? '', 1, 100),
    ])
    courses.value = courseData.items || []
    sessions.value = sessionData.items || []
    if (!selectedCourseId.value && courses.value.length > 0) {
      selectedCourseId.value = String(courses.value[0].id)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载会话失败'
  }
}

// 创建新会话后直接跳转到详情页继续提问。
async function handleCreate() {
  if (!selectedCourseId.value) return
  creating.value = true
  error.value = ''
  success.value = ''
  try {
    const out = await createChatSession(auth.token ?? '', selectedCourseId.value, newSessionTitle.value.trim() || '新对话')
    showCreate.value = false
    showSuccess('会话创建成功')
    await router.push(`/teacher/knowledge/chats/${out.id}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '创建会话失败'
  } finally {
    creating.value = false
  }
}

function openDelete(session: ChatSessionItem) {
  deletingSession.value = session
  showDelete.value = true
}

// 删除会话后刷新列表，保持页面与后端状态一致。
async function handleDelete() {
  if (!deletingSession.value) return
  deleting.value = true
  error.value = ''
  success.value = ''
  try {
    await deleteChatSession(auth.token ?? '', deletingSession.value.id)
    showDelete.value = false
    deletingSession.value = null
    showSuccess('会话已删除')
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除会话失败'
  } finally {
    deleting.value = false
  }
}

onMounted(() => void load())
onBeforeUnmount(() => {
  if (successTimer) clearTimeout(successTimer)
})
</script>

<style scoped>
.workspace-header {
  padding-bottom: 16px;
}

.list-toolbar {
  align-items: center;
}

.empty-tips {
  padding: 18px 0;
}

.chat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.course-item {
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.course-item h4 {
  font-size: 16px;
  margin-bottom: 4px;
}

.course-item .mono,
.course-item .subtle {
  margin: 2px 0;
}

.course-item .action-row {
  align-items: center;
  border-top: 1px solid #e2e8f0;
  padding-top: 10px;
}

.course-item .action-row > button {
  min-width: 86px;
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
  width: min(560px, calc(100vw - 32px));
}
.success {
  background: #ecfdf5;
  border-color: #86efac;
  color: #166534;
}

@media (max-width: 960px) {
  .chat-grid {
    grid-template-columns: 1fr;
  }
}
</style>
