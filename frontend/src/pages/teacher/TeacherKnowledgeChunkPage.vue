<template>
  <div class="course-workspace stack">
    <header class="card workspace-header">
      <h2>资源分块与向量</h2>
      <p class="subtle">{{ summaryText }}</p>
      <div class="action-row">
        <button @click="$router.push(`/teacher/knowledge/${courseId}`)">返回资源列表</button>
      </div>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <section class="card stack">
      <div class="tab-segment">
        <button :class="activeTab === 'chunk' ? 'primary' : ''" @click="activeTab = 'chunk'">分块管理</button>
        <button :class="activeTab === 'embed' ? 'primary' : ''" @click="activeTab = 'embed'">向量嵌入</button>
      </div>
    </section>

    <section v-if="activeTab === 'chunk'" class="card stack">
      <h3>预览分块（未保存）</h3>
      <div class="chunk-preview-layout">
        <div class="chunk-preview-controls stack">
          <label class="checkbox-line">
            <input v-model="clearOnPreview" type="checkbox" />
            预览前清空库内已保存分块（推荐）
          </label>
          <div class="stack">
            <label>chunk_size <input type="number" min="50" v-model.number="chunkSize" /></label>
            <label>overlap <input type="number" min="0" v-model.number="overlap" /></label>
          </div>
          <button :disabled="loading" @click="handlePreview">生成预览</button>
          <button :disabled="loading || previewSegments.length === 0" @click="handleSave">保存预览到库</button>
          <button :disabled="loading" @click="handleConfirm">确认分块</button>
        </div>

        <div class="stack preview-list">
          <p v-if="previewSegments.length === 0" class="subtle">点击「生成预览」后在此编辑或删除草稿块。</p>
          <div class="stack">
            <article v-for="(seg, idx) in previewSegments" :key="seg.key" class="card stack">
              <p class="mono">预览 #{{ idx + 1 }}</p>
              <textarea v-model="seg.content" rows="4" />
              <div class="action-row">
                <button @click="openViewer(`预览分块 #${idx + 1}`, seg.content)">查看全文</button>
                <button class="danger" @click="removePreview(seg.key)">删除</button>
              </div>
            </article>
          </div>
        </div>
      </div>
    </section>

    <section v-if="activeTab === 'chunk'" class="card stack">
      <h3>已保存分块（正式数据）</h3>
      <div v-if="savedChunks.length === 0" class="subtle">暂无已保存分块</div>
      <div class="saved-list">
        <article v-for="row in savedChunks" :key="row.id" class="card stack">
          <p class="mono">
            #{{ row.chunk_index + 1 }} · {{ row.id }} · {{ row.confirmed_at ? '已确认' : '未确认' }} ·
            {{ row.embedded_at ? '已嵌入' : '未嵌入' }}
          </p>
          <template v-if="editingChunkId === row.id">
            <textarea v-model="editingContent" rows="6" />
            <div class="action-row">
              <button @click="editingChunkId = ''">取消</button>
              <button :disabled="loading" @click="saveEditedChunk(row.id)">保存修改</button>
            </div>
          </template>
          <template v-else>
            <pre class="mono chunk-text">{{ row.content }}</pre>
            <div class="action-row">
              <button @click="startEdit(row.id, row.content)">编辑</button>
              <button @click="openViewer(`已保存分块 #${row.chunk_index + 1}`, row.content)">查看全文</button>
              <button class="danger" @click="removeSaved(row.id)">删除</button>
            </div>
          </template>
        </article>
      </div>
    </section>

    <section v-if="activeTab === 'embed'" class="card stack">
      <h3>向量嵌入</h3>
      <label>
        嵌入模型
        <select v-model="embeddingModelId">
          <option value="" disabled>请选择 embedding 模型</option>
          <option v-for="m in embeddingModels" :key="m.id" :value="String(m.id)">{{ m.name }} ({{ m.model_id }})</option>
        </select>
      </label>
      <div class="action-row">
        <button class="primary" :disabled="embedding || !embeddingModelId" @click="handleEmbed">
          {{ embedding ? '嵌入中...' : '开始嵌入' }}
        </button>
      </div>
    </section>

    <div v-if="viewer" class="modal-mask">
      <div class="modal-panel card stack">
        <h3>{{ viewer.title }}</h3>
        <pre class="mono viewer-body">{{ viewer.content }}</pre>
        <div class="action-row">
          <button @click="viewer = null">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  chunkPreview,
  saveKnowledgeChunks,
  confirmKnowledgeChunks,
  listKnowledgeChunks,
  deleteKnowledgeChunk,
  updateKnowledgeChunk,
  clearKnowledgeChunks,
  embedResource,
  type SavedChunkRow,
} from '@/services/knowledge'
import { listAIModels, type AIModelItem } from '@/services/aiModels'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const courseId = String(route.params.courseId || '')
const resourceId = String(route.params.resourceId || '')

const activeTab = ref<'chunk' | 'embed'>('chunk')
const chunkSize = ref(1000)
const overlap = ref(200)
const clearOnPreview = ref(true)
const previewSegments = ref<Array<{ key: string; content: string; char_start?: number; char_end?: number }>>([])
const savedChunks = ref<SavedChunkRow[]>([])
const embeddingModels = ref<AIModelItem[]>([])
const embeddingModelId = ref('')
const editingChunkId = ref('')
const editingContent = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)
const embedding = ref(false)
const viewer = ref<{ title: string; content: string } | null>(null)

const summaryText = computed(() => {
  if (savedChunks.value.length === 0) return `资源 ${resourceId} 尚未保存分块`
  const total = savedChunks.value.length
  const confirmed = savedChunks.value.filter((x) => !!x.confirmed_at).length
  const embedded = savedChunks.value.filter((x) => !!x.embedded_at).length
  return `已保存 ${total} 条 · 已确认 ${confirmed} 条 · 已嵌入 ${embedded} 条`
})

function openViewer(title: string, content: string) {
  viewer.value = { title, content }
}

// 加载已保存分块作为“正式数据”来源，供确认、编辑、删除和嵌入流程使用。
async function loadSavedChunks() {
  try {
    const out = await listKnowledgeChunks(auth.token ?? '', resourceId)
    savedChunks.value = out.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载分块失败'
  }
}

// 仅在内存中生成可编辑预览，不直接落库，避免误写入。
async function handlePreview() {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    if (clearOnPreview.value) {
      await clearKnowledgeChunks(auth.token ?? '', resourceId)
    }
    const out = await chunkPreview(auth.token ?? '', resourceId, chunkSize.value, overlap.value, false)
    previewSegments.value = (out.segments || []).map((s) => ({
      key: `${s.index}-${Math.random()}`,
      content: s.content,
      char_start: s.char_start,
      char_end: s.char_end,
    }))
    await loadSavedChunks()
    success.value = `已生成 ${previewSegments.value.length} 条预览分块`
  } catch (err) {
    error.value = err instanceof Error ? err.message : '预览失败'
  } finally {
    loading.value = false
  }
}

function removePreview(key: string) {
  previewSegments.value = previewSegments.value.filter((x) => x.key !== key)
}

// 将预览分块批量写入数据库，形成后续确认与嵌入的正式输入。
async function handleSave() {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await saveKnowledgeChunks(
      auth.token ?? '',
      resourceId,
      previewSegments.value.map((s) => ({
        content: s.content,
        char_start: s.char_start,
        char_end: s.char_end,
      })),
    )
    previewSegments.value = []
    await loadSavedChunks()
    success.value = '已保存预览分块'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    loading.value = false
  }
}

// 确认后分块进入可嵌入状态，后续仅对确认分块进行向量化。
async function handleConfirm() {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await confirmKnowledgeChunks(auth.token ?? '', resourceId)
    await loadSavedChunks()
    success.value = '分块已确认'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '确认失败'
  } finally {
    loading.value = false
  }
}

function startEdit(chunkId: string, content: string) {
  editingChunkId.value = chunkId
  editingContent.value = content
}

async function saveEditedChunk(chunkId: string) {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await updateKnowledgeChunk(auth.token ?? '', resourceId, chunkId, { content: editingContent.value })
    editingChunkId.value = ''
    await loadSavedChunks()
    success.value = '分块已更新'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '更新失败'
  } finally {
    loading.value = false
  }
}

async function removeSaved(chunkId: string) {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await deleteKnowledgeChunk(auth.token ?? '', resourceId, chunkId)
    if (editingChunkId.value === chunkId) editingChunkId.value = ''
    await loadSavedChunks()
    success.value = '分块已删除'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败'
  } finally {
    loading.value = false
  }
}

// 执行资源级嵌入任务，使用教师选择的 embedding 模型。
async function handleEmbed() {
  if (!embeddingModelId.value) return
  embedding.value = true
  error.value = ''
  success.value = ''
  try {
    await embedResource(auth.token ?? '', resourceId, embeddingModelId.value)
    await loadSavedChunks()
    success.value = '向量嵌入完成'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '嵌入失败'
  } finally {
    embedding.value = false
  }
}

async function loadEmbeddingModels() {
  try {
    const out = await listAIModels(auth.token ?? '')
    embeddingModels.value = (out.items || []).filter((m) => m.model_type === 'embedding')
    if (embeddingModels.value.length > 0) {
      embeddingModelId.value = String(embeddingModels.value[0].id)
    }
  } catch {
    // 模型列表失败不阻断分块流程。
  }
}

onMounted(async () => {
  await Promise.all([loadSavedChunks(), loadEmbeddingModels()])
})
</script>

<style scoped>
.tab-segment {
  display: inline-flex;
  gap: 4px;
  padding: 3px;
  border-radius: 10px;
  border: 1px solid #dbe4f0;
  background: #f8fafc;
}

.tab-segment button {
  border-radius: 8px;
  min-height: 34px;
}

.chunk-preview-layout {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 12px;
  min-height: 0;
}

.chunk-preview-controls {
  align-self: start;
}

.preview-list {
  min-height: 0;
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 4px;
}

.saved-list {
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 4px;
}

.chunk-text {
  white-space: pre-wrap;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 10px;
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
  width: min(900px, calc(100vw - 32px));
}
.viewer-body {
  max-height: 60vh;
  overflow: auto;
  white-space: pre-wrap;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 12px;
}
.checkbox-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #475569;
}

.card h3 {
  margin: 0;
  font-size: 18px;
}

.card > .action-row {
  align-items: center;
}
.success {
  background: #ecfdf5;
  border-color: #86efac;
  color: #166534;
}

@media (max-width: 960px) {
  .chunk-preview-layout {
    grid-template-columns: 1fr;
  }
}
</style>
