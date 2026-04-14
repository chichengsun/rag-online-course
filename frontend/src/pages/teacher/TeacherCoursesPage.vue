<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>课程列表管理</h2>
      <p class="subtle">支持搜索、分页、创建和进入课程目录管理。</p>
    </header>
    <div v-if="message" class="alert success">{{ message }}</div>
    <div v-if="error" class="alert error">{{ error }}</div>

    <section class="card stack">
      <div class="list-toolbar">
        <div class="list-toolbar-actions">
          <input v-model="keyword" placeholder="搜索课程标题或描述（支持回车）" @keydown.enter="runSearch" />
          <select v-model="statusFilter">
            <option value="">全部状态</option>
            <option value="draft">draft</option>
            <option value="published">published</option>
            <option value="archived">archived</option>
          </select>
          <select v-model="sortBy">
            <option value="created_at">创建时间</option>
            <option value="updated_at">更新时间</option>
            <option value="title">标题</option>
          </select>
          <select v-model="sortOrder">
            <option value="desc">降序</option>
            <option value="asc">升序</option>
          </select>
          <button @click="runSearch" :disabled="loading">搜索</button>
        </div>
        <button class="primary" @click="showCreateModal = true">新增课程</button>
      </div>

      <div class="course-grid">
        <p v-if="courses.length === 0 && !loading" class="subtle" style="padding: 1rem 0">
          暂无课程。可尝试调整关键词或状态筛选，或点击「新增课程」。
        </p>
        <article v-for="course in courses" :key="course.id" class="course-item">
          <div>
            <h4>{{ course.title }}</h4>
            <p class="subtle">{{ course.description || '暂无描述' }}</p>
            <p class="mono">状态：{{ course.status }}</p>
          </div>
          <div class="action-row">
            <button class="primary" @click="$router.push(`/teacher/knowledge/${course.id}`)">知识库管理</button>
            <button @click="$router.push(`/teacher/course-content/${course.id}`)">进入章节与资源管理</button>
            <button class="danger" @click="removeCourse(course.id, course.title)">删除</button>
          </div>
        </article>
      </div>

      <div class="pagination-bar">
        <button :disabled="page <= 1 || loading" @click="loadCourses(page - 1)">上一页</button>
        <span class="mono">第 {{ page }} / {{ totalPages }} 页（共 {{ total }} 条）</span>
        <button :disabled="page >= totalPages || loading" @click="loadCourses(page + 1)">下一页</button>
      </div>
    </section>

    <div v-if="showCreateModal" class="modal-mask" @click="showCreateModal = false">
      <div class="modal-card" @click.stop>
        <h3>新增课程</h3>
        <label>
          课程标题
          <input v-model="createForm.title" />
        </label>
        <label>
          课程描述
          <textarea rows="4" v-model="createForm.description" />
        </label>
        <div class="action-row">
          <button @click="showCreateModal = false">取消</button>
          <button class="primary" :disabled="!createForm.title || loading" @click="handleCreateCourse">创建</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { createCourse, deleteCourse, listTeacherCourses, type TeacherCourseItem } from '@/services/course'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const accessToken = computed(() => auth.token ?? '')
const message = ref('')
const error = ref('')
const loading = ref(false)
const courses = ref<TeacherCourseItem[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const keyword = ref('')
const statusFilter = ref('')
const sortBy = ref('created_at')
const sortOrder = ref('desc')
const showCreateModal = ref(false)
const createForm = reactive({ title: '', description: '' })

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
let keywordTimer: number | undefined

async function loadCourses(targetPage = page.value) {
  loading.value = true
  error.value = ''
  try {
    const data = await listTeacherCourses(
      accessToken.value,
      targetPage,
      pageSize.value,
      keyword.value,
      statusFilter.value,
      sortBy.value,
      sortOrder.value,
    )
    courses.value = data.items
    total.value = data.total
    page.value = data.page
    pageSize.value = data.page_size
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载课程列表失败'
  } finally {
    loading.value = false
  }
}

async function handleCreateCourse() {
  loading.value = true
  error.value = ''
  try {
    await createCourse(accessToken.value, createForm)
    showCreateModal.value = false
    createForm.title = ''
    createForm.description = ''
    await loadCourses(1)
    message.value = '课程创建成功'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '创建课程失败'
  } finally {
    loading.value = false
  }
}

async function removeCourse(courseId: string, title: string) {
  if (!window.confirm(`确定删除课程「${title}」？此操作不可恢复。`)) return
  try {
    await deleteCourse(accessToken.value, courseId)
    await loadCourses(page.value)
    message.value = '课程已删除'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除课程失败'
  }
}

function runSearch() {
  void loadCourses(1)
}

watch(keyword, () => {
  window.clearTimeout(keywordTimer)
  keywordTimer = window.setTimeout(() => {
    void loadCourses(1)
  }, 300)
})

onMounted(() => {
  void loadCourses(1)
})
</script>

<style scoped>
.list-toolbar-actions > input {
  min-width: 260px;
}

.list-toolbar-actions > select {
  min-width: 120px;
}

.course-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.course-item {
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.course-item .action-row {
  align-items: center;
  border-top: 1px solid #e2e8f0;
  padding-top: 10px;
}

.course-item .action-row button {
  min-height: 34px;
  padding: 6px 10px;
  font-size: 13px;
}

@media (max-width: 920px) {
  .list-toolbar-actions > input,
  .list-toolbar-actions > select {
    min-width: 100%;
  }

  .course-grid {
    grid-template-columns: 1fr;
  }
}
</style>
