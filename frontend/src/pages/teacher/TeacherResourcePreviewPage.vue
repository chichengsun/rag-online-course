<template>
  <div class="resource-preview-shell">
    <div class="resource-preview-top">
      <div class="action-row">
        <button @click="$router.back()">返回</button>
      </div>
      <h1>资源预览</h1>
    </div>

    <section class="card resource-meta-card">
      <div class="resource-preview-header">
        <div>
          <h2>{{ detail?.title || '资源预览' }}</h2>
          <p class="subtle">类型：{{ detail?.mime_type || detail?.resource_type || '未知' }}</p>
        </div>
        <div class="action-row">
          <a v-if="detail?.object_url" class="button-link primary-link" :href="detail.object_url" target="_blank" rel="noreferrer">新窗口打开</a>
        </div>
      </div>
    </section>

    <section class="card resource-preview-page">
      <p v-if="loading" class="subtle">加载中...</p>
      <p v-if="error" class="alert error">{{ error }}</p>

      <DocumentPreview
        v-if="!loading && !error && previewUrl && isIframeType"
        :url="previewUrl"
        :file-name="detail?.title"
        :mime-type="detail?.mime_type"
        :resource-type="detail?.resource_type"
      />

      <video
        v-if="!loading && !error && previewUrl && isVideo"
        class="resource-preview-frame"
        controls
        :src="previewUrl"
      />

      <audio
        v-if="!loading && !error && previewUrl && isAudio"
        class="resource-audio-frame"
        controls
        :src="previewUrl"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getTeacherResourceDetail } from '@/services/course'
import { useAuthStore } from '@/stores/auth'
import DocumentPreview from '@/components/DocumentPreview.vue'

type Detail = {
  title?: string
  mime_type?: string
  resource_type?: string
  object_url?: string
}

const route = useRoute()
const auth = useAuthStore()
const token = computed(() => auth.token ?? '')
const resourceId = computed(() => String(route.query.resource_id || route.query.resourceId || ''))

const detail = ref<Detail | null>(null)
const previewUrl = ref('')
const loading = ref(false)
const error = ref('')

const isVideo = computed(() => detail.value?.resource_type === 'video')
const isAudio = computed(() => detail.value?.resource_type === 'audio')
const isIframeType = computed(() => !isVideo.value && !isAudio.value)

async function load() {
  if (!resourceId.value) return
  loading.value = true
  error.value = ''
  try {
    detail.value = await getTeacherResourceDetail(token.value, resourceId.value)
    previewUrl.value = detail.value?.object_url || ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : '预览失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<style scoped>
.resource-preview-shell {
  display: flex;
  min-height: 0;
  height: 100%;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
}

.resource-preview-top h1 {
  margin: 8px 0 0;
  font-size: 28px;
  line-height: 1.2;
}

.resource-meta-card {
  flex: 0 0 auto;
}

.resource-preview-page {
  min-height: 0;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow: hidden;
}

.resource-preview-header h2 {
  margin: 0;
  font-size: 20px;
  line-height: 1.2;
}

.resource-preview-header .action-row {
  align-items: center;
}

.resource-preview-frame,
.resource-audio-frame,
.resource-text-viewer {
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
  border-radius: 12px;
}

.resource-preview-frame {
  background: #000;
}
</style>
