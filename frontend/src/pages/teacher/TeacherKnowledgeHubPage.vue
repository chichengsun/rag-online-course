<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>课程知识库管理</h2>
      <p class="subtle">选择课程后进入资源分块与嵌入管理</p>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>
    <section class="card stack">
      <div class="course-list">
        <article v-for="course in courses" :key="course.id" class="course-item">
          <div>
            <h4>{{ course.title }}</h4>
            <p class="subtle">{{ course.description || '暂无描述' }}</p>
            <p class="mono">状态：{{ course.status }}</p>
          </div>
          <div class="action-row">
            <button class="primary" @click="$router.push(`/teacher/knowledge/${course.id}`)">进入知识库</button>
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
const error = ref('')

async function loadCourses() {
  // 加载教师课程列表，用于进入对应课程的知识库管理页。
  try {
    const data = await listTeacherCourses(auth.token ?? '', 1, 100, '', '', 'updated_at', 'desc')
    courses.value = data.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载课程失败'
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

.course-item .action-row {
  align-items: center;
}
</style>
