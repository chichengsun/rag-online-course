<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>知识库资源</h2>
      <p class="subtle">课程 ID：{{ courseId }} · 仅列出支持解析与分块的文档类型</p>
      <div class="action-row">
        <button @click="$router.push('/teacher/courses')">返回课程列表</button>
      </div>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>
    <section v-if="loading" class="card stack">
      <p class="subtle">加载中...</p>
    </section>
    <section class="card stack">
      <div v-if="!loading && items.length === 0" class="subtle">暂无符合条件的资源，请先在课程内容中上传文档资源</div>
      <div v-else class="resource-list">
        <article v-for="item in items" :key="item.id" class="course-item">
          <div>
            <h4>{{ item.title }}</h4>
            <p class="subtle">
              {{ item.chapter_title }} / {{ item.section_title || '未分节' }} · {{ item.resource_type.toUpperCase() }}
            </p>
            <span :class="['status-pill', getStatus(item)]">{{ getStatusLabel(item) }}</span>
            <p class="mono">
              分块：{{ item.chunk_count }} · 已嵌入：{{ item.embedded_count }} · 字符：{{ item.total_chunk_chars }}
            </p>
          </div>
          <div class="action-row">
            <button class="primary" @click="$router.push(`/teacher/knowledge/${courseId}/chunk/${item.id}`)">
              分块管理
            </button>
          </div>
        </article>
      </div>
      <div v-if="total > 0" class="list-toolbar">
        <span class="subtle">共 {{ total }} 条记录，第 {{ page }} / {{ totalPages }} 页</span>
        <div class="action-row">
          <button :disabled="loading || page <= 1" @click="load(page - 1)">上一页</button>
          <button :disabled="loading || page >= totalPages" @click="load(page + 1)">下一页</button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { listKnowledgeResources, type KnowledgeResourceRow } from '@/services/knowledge'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const courseId = String(route.params.courseId || '')

const items = ref<KnowledgeResourceRow[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(1)

function toNum(value: number | string | undefined) {
  if (value === undefined) return 0
  return typeof value === 'number' ? value : Number(value)
}

function getStatus(item: KnowledgeResourceRow) {
  if (toNum(item.embedded_count) > 0) return 'embedded'
  if (toNum(item.chunk_count) > 0) return 'chunked'
  return 'none'
}

function getStatusLabel(item: KnowledgeResourceRow) {
  const s = getStatus(item)
  if (s === 'embedded') return '已嵌入'
  if (s === 'chunked') return '已分块'
  return '未分块'
}

// 支持分页加载知识库资源，便于大课程下的资源管理。
async function load(targetPage = 1) {
  if (!courseId) return
  loading.value = true
  error.value = ''
  // 按课程维度加载已上传的知识资源与分块/嵌入统计信息。
  try {
    const data = await listKnowledgeResources(auth.token ?? '', courseId, targetPage, pageSize)
    items.value = data.items || []
    total.value = toNum(data.total)
    page.value = targetPage
    totalPages.value = Math.max(1, Math.ceil(total.value / pageSize))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载知识库资源失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => void load())
</script>

<style scoped>
.course-item {
  align-items: center;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.course-item:hover {
  border-color: #93c5fd;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.1);
}

.resource-list {
  display: grid;
  gap: 10px;
}

.course-item h4 {
  font-size: 16px;
  margin-bottom: 4px;
}

.course-item .mono {
  margin-top: 6px;
}

.status-pill {
  display: inline-block;
  margin-top: 4px;
  margin-bottom: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  border: 1px solid transparent;
}
.status-pill.none {
  background: #f1f5f9;
  color: #475569;
  border-color: #cbd5e1;
}
.status-pill.chunked {
  background: #eff6ff;
  color: #1d4ed8;
  border-color: #93c5fd;
}
.status-pill.embedded {
  background: #ecfdf5;
  color: #166534;
  border-color: #86efac;
}

.list-toolbar {
  margin-top: 6px;
  padding-top: 10px;
  border-top: 1px solid #e2e8f0;
}
</style>
