<template>
  <div class="course-workspace stack">
    <header class="card workspace-header">
      <h2>课程题库</h2>
      <p class="subtle">课程 {{ courseId }}</p>
      <div class="action-row">
        <button @click="$router.push('/teacher/question-bank')">返回课程选择</button>
      </div>
    </header>

    <nav class="card stack" style="flex-direction: row; gap: 0.75rem; flex-wrap: wrap">
      <router-link
        :to="`${base}/list`"
        style="text-decoration: none; padding: 0.35rem 0.75rem; border-radius: 6px"
        :style="{ fontWeight: isListTab ? 700 : 400, opacity: isListTab ? 1 : 0.75 }"
      >
        已入库题目
      </router-link>
      <router-link
        :to="`${base}/maintain`"
        style="text-decoration: none; padding: 0.35rem 0.75rem; border-radius: 6px"
        :style="{ fontWeight: !isListTab ? 700 : 400, opacity: !isListTab ? 1 : 0.75 }"
      >
        新增与 AI 导题
      </router-link>
    </nav>

    <div v-if="error" class="alert error">{{ error }}</div>
    <div v-if="success" class="alert success">{{ success }}</div>

    <!-- 维护页：预览全文、解析草稿、手工新增 -->
    <template v-if="isMaintain">
      <section class="card stack">
        <h3>上传文本 · AI 解析</h3>
        <p class="subtle">解析结果为草稿，编辑后「确认入库」才写入题库。</p>
        <div class="list-toolbar-actions">
          <input type="file" accept=".txt,.md,.csv" @change="onFileChange" />
          <button :disabled="!uploadFile" @click="handlePreviewFull">预览完整文件</button>
          <button class="primary" :disabled="!uploadFile || parsing" @click="handleParse">
            {{ parsing ? '解析中...' : 'AI 解析为草稿' }}
          </button>
        </div>
        <div v-if="previewOpen" class="card stack" style="margin-top: 0.75rem; max-height: 70vh; overflow: auto">
          <p class="subtle">全文预览 · {{ uploadFile?.name }}（{{ filePreviewText.length }} 字符）</p>
          <pre style="white-space: pre-wrap; font-size: 12px">{{ filePreviewText }}</pre>
          <button @click="previewOpen = false">关闭预览</button>
        </div>
      </section>

      <section v-if="drafts.length > 0" class="card stack">
        <h3>解析草稿（{{ drafts.length }}）</h3>
        <article v-for="(row, index) in drafts" :key="row.key" class="course-item" style="margin-bottom: 1rem">
          <p class="subtle">第 {{ index + 1 }} 题</p>
          <label>
            类型
            <input v-model="row.question_type" />
          </label>
          <label>
            题干
            <textarea v-model="row.stem" rows="4" />
          </label>
          <label>
            参考答案
            <textarea v-model="row.reference_answer" rows="2" />
          </label>
          <button class="danger" @click="removeDraft(index)">移除本题</button>
        </article>
        <div class="action-row">
          <button class="primary" :disabled="confirming" @click="handleConfirmDrafts">
            {{ confirming ? '入库中...' : '确认入库' }}
          </button>
          <button @click="drafts = []">清空草稿</button>
        </div>
      </section>

      <section class="card stack">
        <h3>手工新增单题</h3>
        <div class="double-grid">
          <label>
            题目类型
            <input v-model="createForm.question_type" placeholder="single_choice / true_false" />
          </label>
          <label>
            参考答案
            <input v-model="createForm.reference_answer" />
          </label>
        </div>
        <label>
          题干
          <textarea v-model="createForm.stem" rows="4" />
        </label>
        <div class="action-row">
          <button
            class="primary"
            :disabled="creating || !createForm.question_type || !createForm.stem || !createForm.reference_answer"
            @click="handleCreateSubmit"
          >
            {{ creating ? '提交中...' : '新增并前往列表' }}
          </button>
        </div>
      </section>
    </template>

    <!-- 列表页：分页 + 编辑弹层 -->
    <template v-else>
      <section class="card stack">
        <div class="list-toolbar">
          <h3>已入库题目</h3>
        </div>
        <div class="double-grid" style="margin-bottom: 0.75rem">
          <label>
            关键词
            <input v-model="keywordInput" placeholder="题干/答案/类型" @keydown.enter="applyFilters" />
          </label>
          <label>
            题型
            <select v-model="typeFilter" @change="onTypeChange">
              <option value="">全部</option>
              <option value="single_choice">单选</option>
              <option value="multiple_choice">多选</option>
              <option value="true_false">判断</option>
              <option value="short_answer">简答</option>
              <option value="fill_blank">填空</option>
            </select>
          </label>
        </div>
        <div class="action-row">
          <button class="primary" @click="applyFilters">搜索</button>
        </div>
        <div v-if="loading" class="subtle">加载中...</div>
        <div v-else-if="items.length === 0" class="subtle">暂无题目</div>
        <div v-else class="course-list">
          <article v-for="item in items" :key="item.id" class="course-item">
            <div>
              <h4 style="white-space: pre-wrap">{{ item.stem }}</h4>
              <p class="mono">类型：{{ item.question_type }}</p>
              <p class="subtle">参考答案：{{ item.reference_answer }}</p>
              <p v-if="item.source_file_name" class="subtle">来源：{{ item.source_file_name }}</p>
            </div>
            <div class="action-row">
              <button @click="openEdit(item)">编辑</button>
              <button class="danger" :disabled="deletingId === item.id" @click="handleDelete(item.id)">
                {{ deletingId === item.id ? '删除中...' : '删除' }}
              </button>
            </div>
          </article>
        </div>
        <div v-if="!loading && total > 0" class="action-row" style="margin-top: 1rem; justify-content: space-between">
          <span class="subtle">共 {{ total }} 条 · 第 {{ page }} / {{ totalPages }} 页</span>
          <div class="action-row">
            <button :disabled="page <= 1" @click="prevPage">上一页</button>
            <button :disabled="page >= totalPages" @click="nextPage">下一页</button>
          </div>
        </div>
      </section>
    </template>

    <!-- 编辑弹层 -->
    <div
      v-if="editOpen"
      style="position: fixed; inset: 0; background: rgba(0, 0, 0, 0.35); z-index: 50; display: flex; align-items: center; justify-content: center; padding: 1rem"
      @click.self="closeEdit"
    >
      <div class="card stack" style="max-width: 640px; width: 100%; max-height: 90vh; overflow: auto" @click.stop>
        <h3>编辑题目（ID: {{ editingId }}）</h3>
        <label>
          类型
          <input v-model="editForm.question_type" />
        </label>
        <label>
          题干
          <textarea v-model="editForm.stem" rows="5" />
        </label>
        <label>
          参考答案
          <textarea v-model="editForm.reference_answer" rows="3" />
        </label>
        <div class="action-row">
          <button :disabled="savingEdit" @click="closeEdit">取消</button>
          <button
            class="primary"
            :disabled="savingEdit || !editForm.question_type || !editForm.stem || !editForm.reference_answer"
            @click="handleEditSubmit"
          >
            {{ savingEdit ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  createQuestion,
  deleteQuestion,
  parseImportFile,
  confirmImportBatch,
  listQuestionBank,
  updateQuestion,
  type QuestionBankItem,
} from '@/services/questionBank'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const courseId = String(route.params.courseId || '')
const base = `/teacher/question-bank/${courseId}`

const isMaintain = computed(() => route.path.endsWith('/maintain'))
const isListTab = computed(() => route.path.endsWith('/list'))

const loading = ref(false)
const creating = ref(false)
const savingEdit = ref(false)
const parsing = ref(false)
const confirming = ref(false)
const deletingId = ref('')
const items = ref<QuestionBankItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keywordInput = ref('')
const appliedKeyword = ref('')
const typeFilter = ref('')
const uploadFile = ref<File | null>(null)
const previewOpen = ref(false)
const filePreviewText = ref('')
const error = ref('')
const success = ref('')

type DraftRow = { key: string; question_type: string; stem: string; reference_answer: string }
const drafts = ref<DraftRow[]>([])

const editOpen = ref(false)
const editingId = ref('')

const createForm = reactive({
  question_type: '',
  stem: '',
  reference_answer: '',
})

const editForm = reactive({
  question_type: '',
  stem: '',
  reference_answer: '',
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

function newKey() {
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

function setSuccess(msg: string) {
  success.value = msg
  window.setTimeout(() => {
    if (success.value === msg) success.value = ''
  }, 4000)
}

async function loadItems() {
  if (!courseId || isMaintain.value) return
  loading.value = true
  error.value = ''
  try {
    const out = await listQuestionBank(auth.token ?? '', courseId, {
      page: page.value,
      page_size: pageSize.value,
      keyword: appliedKeyword.value || undefined,
      question_type: typeFilter.value || undefined,
    })
    items.value = out.items || []
    total.value = Number(out.total ?? 0)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载题库失败'
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  appliedKeyword.value = keywordInput.value.trim()
  page.value = 1
  void loadItems()
}

function prevPage() {
  if (page.value <= 1) return
  page.value--
  void loadItems()
}

function nextPage() {
  if (page.value >= totalPages.value) return
  page.value++
  void loadItems()
}

function onTypeChange() {
  page.value = 1
  void loadItems()
}

function onFileChange(evt: Event) {
  const input = evt.target as HTMLInputElement
  uploadFile.value = input.files?.[0] || null
}

async function handlePreviewFull() {
  if (!uploadFile.value) return
  try {
    filePreviewText.value = await uploadFile.value.text()
    previewOpen.value = true
  } catch {
    error.value = '读取文件失败'
  }
}

async function handleParse() {
  if (!courseId || !uploadFile.value) return
  parsing.value = true
  error.value = ''
  try {
    const questions = await parseImportFile(auth.token ?? '', courseId, uploadFile.value)
    drafts.value = questions.map((q) => ({
      key: newKey(),
      question_type: q.question_type,
      stem: q.stem,
      reference_answer: q.reference_answer,
    }))
    if (questions.length === 0) {
      error.value = '模型未返回有效题目'
    } else {
      setSuccess(`已解析 ${questions.length} 道题`)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '解析失败'
  } finally {
    parsing.value = false
  }
}

function removeDraft(i: number) {
  drafts.value = drafts.value.filter((_, idx) => idx !== i)
}

async function handleConfirmDrafts() {
  if (!courseId || drafts.value.length === 0) return
  for (const d of drafts.value) {
    if (!d.stem.trim() || !d.reference_answer.trim() || !d.question_type) {
      error.value = '请补全每道题的类型、题干与参考答案'
      return
    }
  }
  confirming.value = true
  error.value = ''
  try {
    await confirmImportBatch(auth.token ?? '', courseId, {
      source_file_name: uploadFile.value?.name ?? '',
      questions: drafts.value.map(({ question_type, stem, reference_answer }) => ({
        question_type,
        stem,
        reference_answer,
      })),
    })
    drafts.value = []
    uploadFile.value = null
    setSuccess('入库成功')
    await router.push(`${base}/list`)
    await loadItems()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '入库失败'
  } finally {
    confirming.value = false
  }
}

async function handleCreateSubmit() {
  if (!courseId) return
  creating.value = true
  error.value = ''
  const payload = {
    question_type: createForm.question_type.trim(),
    stem: createForm.stem.trim(),
    reference_answer: createForm.reference_answer.trim(),
  }
  try {
    await createQuestion(auth.token ?? '', courseId, payload)
    setSuccess('新增成功')
    createForm.question_type = ''
    createForm.stem = ''
    createForm.reference_answer = ''
    await router.push(`${base}/list`)
    await loadItems()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '提交失败'
  } finally {
    creating.value = false
  }
}

function openEdit(item: QuestionBankItem) {
  editingId.value = item.id
  editForm.question_type = item.question_type
  editForm.stem = item.stem
  editForm.reference_answer = item.reference_answer
  editOpen.value = true
}

function closeEdit() {
  editOpen.value = false
  editingId.value = ''
}

async function handleEditSubmit() {
  if (!editingId.value) return
  savingEdit.value = true
  error.value = ''
  const payload = {
    question_type: editForm.question_type.trim(),
    stem: editForm.stem.trim(),
    reference_answer: editForm.reference_answer.trim(),
  }
  try {
    await updateQuestion(auth.token ?? '', editingId.value, payload)
    setSuccess('已保存')
    closeEdit()
    await loadItems()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '提交失败'
  } finally {
    savingEdit.value = false
  }
}

async function handleDelete(id: string) {
  if (!window.confirm('确定删除该题目吗？')) return
  deletingId.value = id
  error.value = ''
  try {
    await deleteQuestion(auth.token ?? '', id)
    setSuccess('删除成功')
    if (editingId.value === id) closeEdit()
    await loadItems()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败'
  } finally {
    deletingId.value = ''
  }
}

watch(
  () => route.path,
  () => {
    if (!isMaintain.value) void loadItems()
  },
)

onMounted(() => {
  if (!isMaintain.value) void loadItems()
})
</script>
