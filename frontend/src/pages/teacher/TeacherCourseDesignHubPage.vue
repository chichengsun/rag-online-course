<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>课程设计</h2>
      <p class="subtle">选择课程后进入 AI 大纲草案生成与应用</p>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>

    <section class="card stack">
      <div v-if="loading" class="subtle">加载中...</div>
      <div v-else-if="courses.length === 0" class="subtle">暂无课程，请先创建课程</div>
      <div v-else class="course-list">
        <article v-for="course in courses" :key="course.id" class="course-item">
          <div>
            <h4>{{ course.title }}</h4>
            <p class="subtle">{{ course.description || '暂无描述' }}</p>
            <p class="mono">状态：{{ course.status }} · 更新时间：{{ formatDate(course.updated_at) }}</p>
          </div>
          <div class="action-row">
            <button class="primary" @click="$router.push(`/teacher/course-design/${course.id}`)">进入课程设计</button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listTeacherCourses, type TeacherCourseItem } from '@/services/course'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const courses = ref<TeacherCourseItem[]>([])
const loading = ref(false)
const error = ref('')

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleDateString('zh-CN') : '-'
}

// 课程设计入口页仅负责课程选择，不承载具体草案编辑逻辑。
async function loadCourses() {
  loading.value = true
  error.value = ''
  try {
    const data = await listTeacherCourses(auth.token ?? '', 1, 100, '', '', 'updated_at', 'desc')
    courses.value = data.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载课程失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => void loadCourses())
</script>

<style scoped>
.workspace-header {
  padding-bottom: 16px;
}

.course-item {
  align-items: center;
}

.course-item h4 {
  margin-bottom: 4px;
  font-size: 16px;
}

.course-item .mono {
  margin-top: 4px;
}

.course-item .action-row {
  align-items: center;
}
</style>
