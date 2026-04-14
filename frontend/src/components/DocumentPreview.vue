<template>
  <div class="doc-preview-root">
    <p v-if="loading" class="subtle">文档加载中...</p>
    <p v-else-if="error" class="alert error">{{ error }}</p>

    <div v-else-if="kind === 'docx'" ref="docxContainerRef" class="docx-container" />

    <VueOfficePptx
      v-else-if="kind === 'pptx' && pptxSrc"
      class="pptx-container"
      :src="pptxSrc"
      @rendered="onRendered"
      @error="onPptxError"
    />

    <div v-else-if="kind === 'sheet'" class="sheet-container" v-html="sheetHtml" />

    <iframe v-else-if="kind === 'iframe' && objectUrl" class="resource-preview-frame" :src="objectUrl" />

    <p v-else class="subtle">暂不支持该文档类型预览，请新窗口打开。</p>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import * as XLSX from 'xlsx'
import VueOfficePptx from '@vue-office/pptx'

const props = defineProps<{
  url: string
  fileName?: string
  mimeType?: string
  resourceType?: string
}>()

const loading = ref(false)
const error = ref('')
const objectUrl = ref('')
const arrayBuffer = ref<ArrayBuffer | null>(null)
const sheetHtml = ref('')
const docxContainerRef = ref<HTMLDivElement | null>(null)
const pptxSrc = ref<ArrayBuffer | null>(null)

const kind = computed(() => {
  const name = (props.fileName || '').toLowerCase()
  const mime = (props.mimeType || '').toLowerCase()
  const type = (props.resourceType || '').toLowerCase()
  if (name.endsWith('.docx') || mime.includes('wordprocessingml') || type === 'docx' || type === 'doc') return 'docx'
  if (name.endsWith('.pptx') || name.endsWith('.ppt') || mime.includes('presentation') || type === 'ppt') return 'pptx'
  if (
    name.endsWith('.xlsx') ||
    name.endsWith('.xls') ||
    name.endsWith('.csv') ||
    mime.includes('spreadsheet') ||
    mime.includes('csv')
  ) {
    return 'sheet'
  }
  return 'iframe'
})

function cleanupObjectUrl() {
  if (objectUrl.value) {
    URL.revokeObjectURL(objectUrl.value)
    objectUrl.value = ''
  }
}

function detectCsvText(buffer: ArrayBuffer) {
  const utf8 = new TextDecoder('utf-8', { fatal: false }).decode(buffer)
  if (!utf8.includes('�')) return utf8
  try {
    return new TextDecoder('gb18030' as any, { fatal: false }).decode(buffer)
  } catch {
    return utf8
  }
}

async function renderDocx() {
  if (!arrayBuffer.value || !docxContainerRef.value) return
  const { renderAsync } = await import('docx-preview')
  docxContainerRef.value.innerHTML = ''
  await renderAsync(arrayBuffer.value, docxContainerRef.value)
}

function renderSheet() {
  if (!arrayBuffer.value) return
  const lowerName = (props.fileName || '').toLowerCase()
  if (lowerName.endsWith('.csv') || (props.mimeType || '').toLowerCase().includes('csv')) {
    const text = detectCsvText(arrayBuffer.value)
    const wb = XLSX.read(text, { type: 'string' })
    const firstSheet = wb.SheetNames[0]
    sheetHtml.value = firstSheet ? XLSX.utils.sheet_to_html(wb.Sheets[firstSheet]) : '<p>空表格</p>'
    return
  }
  const wb = XLSX.read(arrayBuffer.value, { type: 'array' })
  const firstSheet = wb.SheetNames[0]
  sheetHtml.value = firstSheet ? XLSX.utils.sheet_to_html(wb.Sheets[firstSheet]) : '<p>空表格</p>'
}

function onRendered() {
  // noop
}

function onPptxError(e: unknown) {
  error.value = e instanceof Error ? e.message : 'PPT 预览失败'
}

async function load() {
  if (!props.url) return
  loading.value = true
  error.value = ''
  sheetHtml.value = ''
  pptxSrc.value = null
  cleanupObjectUrl()
  try {
    const resp = await fetch(props.url)
    if (!resp.ok) throw new Error(`下载资源失败（${resp.status}）`)
    arrayBuffer.value = await resp.arrayBuffer()
    objectUrl.value = URL.createObjectURL(new Blob([arrayBuffer.value]))

    if (kind.value === 'docx') {
      await nextTick()
      await renderDocx()
    } else if (kind.value === 'pptx') {
      pptxSrc.value = arrayBuffer.value
    } else if (kind.value === 'sheet') {
      renderSheet()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '文档预览失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => void load())
onBeforeUnmount(() => cleanupObjectUrl())
</script>

<style scoped>
.doc-preview-root {
  min-height: 70vh;
}
.docx-container,
.pptx-container,
.sheet-container {
  background: #fff;
  border: 1px solid #dbe4f0;
  border-radius: 12px;
  min-height: 70vh;
  padding: 12px;
}
.sheet-container :deep(table) {
  width: 100%;
  border-collapse: collapse;
}
.sheet-container :deep(th),
.sheet-container :deep(td) {
  border: 1px solid #cbd5e1;
  padding: 6px 8px;
}
</style>
