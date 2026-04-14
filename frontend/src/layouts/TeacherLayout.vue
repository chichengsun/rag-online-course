<template>
  <main class="teacher-app-shell" v-if="isTeacher">
    <aside class="teacher-app-sidebar">
      <div class="brand-block">
        <h1>教师工作台</h1>
        <p>{{ ragDomain ? 'RAG 对话' : '课程管理' }}</p>
      </div>

      <nav class="category-nav">
        <button
          v-for="item in subNav"
          :key="item.key"
          :class="['category-item', isActive(item.to) ? 'active' : '']"
          :disabled="!item.to"
          @click="item.to && $router.push(item.to)"
        >
          <span>{{ item.title }}</span>
          <small>{{ item.desc }}</small>
        </button>
      </nav>

      <div class="sidebar-footer">
        <p class="subtle">
          当前用户：{{ auth.user?.username }}（{{ auth.user?.role }}）
        </p>
        <button class="danger sidebar-logout-btn" @click="handleLogout">退出登录</button>
      </div>
    </aside>

    <div class="teacher-main-shell">
      <header class="teacher-topbar">
        <div class="module-switch" role="tablist" aria-label="教师端主模块">
          <button :class="!ragDomain ? 'active' : ''" @click="$router.push('/teacher/courses')">课程管理</button>
          <button :class="ragDomain ? 'active' : ''" @click="$router.push('/teacher/knowledge/chats')">RAG 对话</button>
        </div>
        <div class="topbar-user">
          <div>{{ auth.user?.username }}</div>
          <small>{{ currentModuleLabel }}</small>
        </div>
      </header>

      <section :class="['teacher-app-main', outletFillHeight ? 'outlet-fill-height' : '']">
        <RouterView />
      </section>
    </div>
  </main>
  <RouterView v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isTeacher = computed(() => auth.user?.role === 'teacher')
const ragDomain = computed(
  () => route.path.startsWith('/teacher/ai-models') || route.path.startsWith('/teacher/knowledge/chats'),
)
const outletFillHeight = computed(
  () =>
    route.path.startsWith('/teacher/knowledge/chats/') ||
    route.path === '/teacher/knowledge/chats' ||
    route.path === '/teacher/resources/preview',
)

const courseSubNav = [
  { key: 'course-list', title: '课程列表管理', desc: '搜索、分页、创建课程', to: '/teacher/courses' },
  { key: 'course-content', title: '章节与资源管理', desc: '章节目录与资源上传维护', to: '/teacher/course-content' },
  { key: 'course-design', title: '课程设计', desc: 'AI 生成大纲草案并应用到课程', to: '/teacher/course-design' },
  { key: 'question-bank', title: '课程题库', desc: '手工维护与AI导题', to: '/teacher/question-bank' },
  { key: 'knowledge', title: '课程知识库管理', desc: '分块、嵌入与可解析资源', to: '/teacher/knowledge' },
]

const ragSubNav = [
  { key: 'chat-management', title: '对话管理', desc: '分页查看会话并进入详情问答', to: '/teacher/knowledge/chats' },
  { key: 'ai-models', title: '模型管理', desc: '问答、嵌入、重排模型配置', to: '/teacher/ai-models' },
]

const subNav = computed(() => (ragDomain.value ? ragSubNav : courseSubNav))
const currentModuleLabel = computed(() => (ragDomain.value ? 'RAG 对话' : '课程管理'))

const isActive = (to?: string) => !!to && (route.path === to || route.path.startsWith(`${to}/`))

function handleLogout() {
  auth.logout()
  router.replace('/auth')
}
</script>

<style scoped>
.sidebar-logout-btn {
  width: 100%;
}

.teacher-main-shell {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.teacher-topbar {
  height: 56px;
  border-bottom: 1px solid #dbe4f0;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}

.module-switch {
  display: inline-flex;
  gap: 3px;
  border: 1px solid #dbe4f0;
  background: #f8fafc;
  border-radius: 9px;
  padding: 3px;
}

.module-switch button {
  border: none;
  background: transparent;
  min-height: 32px;
  padding: 0 12px;
  border-radius: 7px;
}

.module-switch button.active {
  background: #fff;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.12);
  color: #0f172a;
  font-weight: 600;
}

.topbar-user {
  text-align: right;
  font-size: 13px;
  color: #0f172a;
}

.topbar-user small {
  color: #64748b;
}

@media (max-width: 920px) {
  .teacher-topbar {
    padding: 0 12px;
  }

  .topbar-user {
    display: none;
  }
}
</style>
