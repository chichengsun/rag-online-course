<template>
  <div class="course-workspace">
    <header class="card workspace-header">
      <h2>章节与资源管理</h2>
      <p class="subtle">课程结构：章节 -> 节 -> 资源</p>
      <div class="action-row">
        <button @click="$router.push('/teacher/courses')">返回课程列表</button>
        <button class="primary" @click="openChapterCreate">新增章节</button>
      </div>
    </header>

    <div v-if="error" class="alert error">{{ error }}</div>

    <section class="card stack" v-if="chapters.length === 0 && !loading">
      <p class="subtle">暂无章节</p>
    </section>

    <section class="card stack chapter-card" v-for="chapter in chapters" :key="chapter.id">
      <div class="chapter-title-row">
        <h3>{{ chapter.sort_order }}. {{ chapter.title }}</h3>
        <div class="action-row">
          <button @click="openSectionCreate(chapter.id)">新增节</button>
          <button class="danger" @click="handleDeleteChapter(chapter.id)">删除章节</button>
        </div>
      </div>

      <div v-for="section in sectionsByChapter[chapter.id] || []" :key="section.id" class="section-box">
        <div class="chapter-title-row">
          <strong>{{ section.sort_order }}. {{ section.title }}</strong>
          <div class="action-row">
            <button @click="openUpload(section.id)">上传资源</button>
            <button class="danger" @click="handleDeleteSection(section.id)">删除节</button>
          </div>
        </div>

        <div class="resource-list">
          <article v-for="res in resourcesBySection[section.id] || []" :key="res.id" class="course-item resource-row">
            <div>
              <h4>{{ res.title }}</h4>
              <p class="subtle">
                {{ res.resource_type }} · {{ prettySize(res.size_bytes) }}
              </p>
              <p v-if="res.ai_summary_status === 'running'" class="subtle">摘要生成中…</p>
              <p v-if="res.ai_summary_status === 'failed'" class="alert error" style="margin-top: 6px">
                {{ res.ai_summary_error || '摘要失败' }}
              </p>
            </div>
            <div class="action-row">
              <button @click="openPreview(res.id)">预览</button>
              <button
                v-if="canSummary(res.resource_type)"
                @click="handleSummary(res.id)"
                :disabled="res.ai_summary_status === 'running'"
              >
                AI 摘要
              </button>
              <button class="danger" @click="handleDeleteResource(res.id)">删除</button>
            </div>
          </article>
        </div>
      </div>
    </section>

    <div v-if="chapterModal" class="modal-mask" @click="chapterModal = false">
      <div class="modal-card" @click.stop>
        <h3>新增章节</h3>
        <label>标题 <input v-model="chapterForm.title" /></label>
        <label>排序 <input type="number" min="1" v-model.number="chapterForm.sort_order" /></label>
        <div class="action-row">
          <button @click="chapterModal = false">取消</button>
          <button class="primary" @click="submitChapter">创建</button>
        </div>
      </div>
    </div>

    <div v-if="sectionModal" class="modal-mask" @click="sectionModal = false">
      <div class="modal-card" @click.stop>
        <h3>新增节</h3>
        <label>标题 <input v-model="sectionForm.title" /></label>
        <label>排序 <input type="number" min="1" v-model.number="sectionForm.sort_order" /></label>
        <div class="action-row">
          <button @click="sectionModal = false">取消</button>
          <button class="primary" @click="submitSection">创建</button>
        </div>
      </div>
    </div>

    <div v-if="uploadModal" class="modal-mask" @click="uploadModal = false">
      <div class="modal-card" @click.stop>
        <h3>上传资源</h3>
        <label>文件 <input type="file" @change="onFileChange" /></label>
        <label>标题 <input v-model="uploadForm.title" /></label>
        <label>排序 <input type="number" min="1" v-model.number="uploadForm.sort_order" /></label>
        <div class="action-row">
          <button @click="uploadModal = false">取消</button>
          <button class="primary" :disabled="!uploadFile" @click="submitUpload">上传</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  createChapter,
  createSection,
  listCourseChapters,
  listCourseSections,
  listSectionResources,
  initUpload,
  confirmResource,
  deleteChapter,
  deleteSection,
  deleteResource,
  summarizeResource,
  type ChapterItem,
  type SectionItem,
  type ResourceItem,
} from '@/services/course'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const token = computed(() => auth.token ?? '')
const courseId = computed(() => String(route.params.courseId ?? ''))

const chapters = ref<ChapterItem[]>([])
const sectionsByChapter = ref<Record<string, SectionItem[]>>({})
const resourcesBySection = ref<Record<string, ResourceItem[]>>({})
const loading = ref(false)
const error = ref('')

const chapterModal = ref(false)
const sectionModal = ref(false)
const uploadModal = ref(false)
const currentChapterId = ref('')
const currentSectionId = ref('')
const uploadFile = ref<File | null>(null)

const chapterForm = reactive({ title: '', sort_order: 1 })
const sectionForm = reactive({ title: '', sort_order: 1 })
const uploadForm = reactive({ title: '', sort_order: 1 })

const DOC_TYPES = new Set(['pdf', 'txt', 'doc', 'docx', 'ppt'])

function canSummary(type: string) {
  return DOC_TYPES.has(type)
}

function prettySize(v?: number) {
  const bytes = Number(v || 0)
  if (!bytes) return '0 B'
  const mb = bytes / 1024 / 1024
  return `${mb.toFixed(2)} MB`
}

function inferType(fileName: string): ResourceItem['resource_type'] {
  const ext = fileName.toLowerCase().split('.').pop() || ''
  if (ext === 'pdf') return 'pdf'
  if (ext === 'doc') return 'doc'
  if (ext === 'docx') return 'docx'
  if (ext === 'ppt' || ext === 'pptx') return 'ppt'
  if (ext === 'txt' || ext === 'md') return 'txt'
  if (['mp3', 'wav', 'ogg', 'm4a', 'aac', 'flac'].includes(ext)) return 'audio'
  return 'video'
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0] || null
  uploadFile.value = file
  if (file) {
    uploadForm.title = file.name.replace(/\.[^.]+$/, '')
  }
}

async function loadCatalog() {
  if (!courseId.value) return
  loading.value = true
  error.value = ''
  try {
    const ch = await listCourseChapters(token.value, courseId.value)
    chapters.value = ch.items || []
    const secMap: Record<string, SectionItem[]> = {}
    const resMap: Record<string, ResourceItem[]> = {}
    await Promise.all(
      chapters.value.map(async (c) => {
        const sec = await listCourseSections(token.value, courseId.value, c.id)
        secMap[c.id] = sec.items || []
        await Promise.all(
          (sec.items || []).map(async (s) => {
            const rs = await listSectionResources(token.value, s.id)
            resMap[s.id] = rs.items || []
          }),
        )
      }),
    )
    sectionsByChapter.value = secMap
    resourcesBySection.value = resMap
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function openChapterCreate() {
  chapterForm.title = ''
  chapterForm.sort_order = chapters.value.length + 1
  chapterModal.value = true
}

async function submitChapter() {
  await createChapter(token.value, courseId.value, chapterForm)
  chapterModal.value = false
  await loadCatalog()
}

function openSectionCreate(chapterId: string) {
  currentChapterId.value = chapterId
  const n = (sectionsByChapter.value[chapterId] || []).length + 1
  sectionForm.title = ''
  sectionForm.sort_order = n
  sectionModal.value = true
}

async function submitSection() {
  await createSection(token.value, courseId.value, currentChapterId.value, sectionForm)
  sectionModal.value = false
  await loadCatalog()
}

function openUpload(sectionId: string) {
  currentSectionId.value = sectionId
  const n = (resourcesBySection.value[sectionId] || []).length + 1
  uploadForm.title = ''
  uploadForm.sort_order = n
  uploadFile.value = null
  uploadModal.value = true
}

async function submitUpload() {
  if (!uploadFile.value) return
  const file = uploadFile.value
  const resourceType = inferType(file.name)
  const init = await initUpload(token.value, currentSectionId.value, {
    course_id: Number(courseId.value),
    file_name: file.name,
    resource_type: resourceType,
  })
  const putResp = await fetch(init.upload_url, {
    method: 'PUT',
    headers: { 'Content-Type': file.type || 'application/octet-stream' },
    body: file,
  })
  if (!putResp.ok) throw new Error(`上传失败（${putResp.status}）`)
  await confirmResource(token.value, currentSectionId.value, {
    title: uploadForm.title,
    resource_type: resourceType,
    sort_order: uploadForm.sort_order,
    object_key: init.object_key,
    mime_type: file.type || 'application/octet-stream',
    size_bytes: file.size,
  })
  uploadModal.value = false
  await loadCatalog()
}

async function handleSummary(resourceId: string) {
  await summarizeResource(token.value, resourceId)
  await loadCatalog()
}

async function handleDeleteChapter(id: string) {
  if (!window.confirm('确定删除章节及其所有资源吗？')) return
  await deleteChapter(token.value, id)
  await loadCatalog()
}

async function handleDeleteSection(id: string) {
  if (!window.confirm('确定删除该节及其资源吗？')) return
  await deleteSection(token.value, id)
  await loadCatalog()
}

async function handleDeleteResource(id: string) {
  if (!window.confirm('确定删除资源吗？')) return
  await deleteResource(token.value, id)
  await loadCatalog()
}

function openPreview(resourceId: string) {
  router.push(`/teacher/resources/preview?resource_id=${resourceId}`)
}

onMounted(() => {
  void loadCatalog()
})
</script>

<style scoped>
.chapter-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.chapter-title-row h3 {
  margin: 0;
  font-size: 18px;
}

.chapter-card {
  border: 1px solid #dbe4f0;
}

.section-box {
  margin-top: 10px;
  border: 1px solid #dbe4f0;
  border-radius: 12px;
  padding: 12px 12px 10px;
  background: #f8fbff;
}

.section-box strong {
  font-size: 15px;
  color: #0f172a;
}

.resource-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.course-item h4 {
  margin: 0 0 4px;
  font-size: 15px;
}

.resource-row {
  align-items: center;
  border: 1px solid #e2e8f0;
  background: #fff;
}

.resource-row .action-row button {
  min-height: 34px;
  padding: 6px 10px;
  font-size: 13px;
}

.course-item .action-row {
  align-items: center;
}
</style>
