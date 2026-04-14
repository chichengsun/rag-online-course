<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>课程题库管理</h2>
      <p class="subtle">选择课程后进入题库，支持手工编辑与上传文本 AI 解析导题。</p>
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
          </div>
          <div class="action-row">
            <button class="primary" @click="$router.push(`/teacher/question-bank/${course.id}/list`)">进入题库</button>
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

// 课程题库入口只负责课程选择，不承载题目编辑逻辑。
async function loadCourses() {
  loading.value = true
  error.value = ''
  try {
    const out = await listTeacherCourses(auth.token ?? '', 1, 100, '', '', 'updated_at', 'desc')
    courses.value = out.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载课程失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => void loadCourses())
</script>
