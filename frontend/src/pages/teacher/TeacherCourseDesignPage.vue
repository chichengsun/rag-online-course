<template>
  <div class="course-workspace stack">
    <section v-if="invalidCourseId" class="card stack">
      <h3>无效的课程 ID</h3>
      <div class="action-row">
        <button @click="$router.push('/teacher/course-design')">返回课程设计列表</button>
      </div>
    </section>

    <header v-if="!invalidCourseId" class="card workspace-header">
      <div class="action-row">
        <button @click="$router.push('/teacher/course-design')">返回列表</button>
        <button @click="$router.push(`/teacher/course-content/${courseId}`)">打开课程与内容</button>
      </div>
      <h2>课程设计</h2>
      <p class="subtle">{{ courseTitle || `课程 #${courseId}` }}</p>
    </header>

    <div v-if="!invalidCourseId && error" class="alert error">{{ error }}</div>
    <div v-if="!invalidCourseId && success" class="alert success">{{ success }}</div>

    <section v-if="!invalidCourseId" class="card stack">
      <h3>生成设置</h3>
      <label>
        补充说明（可选）
        <textarea v-model="extraHint" rows="3" placeholder="例如：面向零基础、16 周、实验导向" />
      </label>
      <label>
        问答模型
        <select v-model="qaModelId">
          <option value="" disabled>选择问答模型</option>
          <option v-for="m in qaModels" :key="m.id" :value="String(m.id)">{{ m.name }} ({{ m.model_id }})</option>
        </select>
      </label>
      <p v-if="qaModels.length === 0" class="subtle">暂无可用 QA 模型，请先前往模型管理页面配置</p>
      <div class="action-row">
        <button class="primary" :disabled="invalidCourseId || generating || qaModels.length === 0" @click="handleGenerate">
          {{ generating ? '生成中...' : '生成大纲草案' }}
        </button>
      </div>
    </section>

    <section v-if="!invalidCourseId && chapters.length > 0" class="card stack">
      <h3>大纲预览（可编辑）</h3>
      <p class="subtle">应用后会追加到现有章节之后，不覆盖现有章/节</p>
      <article v-for="(ch, ci) in chapters" :key="ci" class="card stack">
        <label>
          第 {{ ci + 1 }} 章
          <input :value="ch.title" @input="updateChapterTitle(ci, ($event.target as HTMLInputElement).value)" />
        </label>
        <div class="stack">
          <label v-for="(sec, si) in ch.sections" :key="si">
            小节 {{ si + 1 }}
            <input :value="sec.title" @input="updateSectionTitle(ci, si, ($event.target as HTMLInputElement).value)" />
          </label>
          <div class="action-row">
            <button @click="addSection(ci)">添加小节</button>
          </div>
        </div>
      </article>
      <div class="action-row">
        <button class="primary" :disabled="invalidCourseId || applying" @click="handleApply">
          {{ applying ? '应用中...' : '将大纲追加到课程' }}
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { listTeacherCourses } from '@/services/course'
import { listAIModels, type AIModelItem } from '@/services/aiModels'
import { applyOutlineDraft, generateOutlineDraft, type OutlineChapterDraft } from '@/services/courseDesign'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const courseId = Number(route.params.courseId)
const invalidCourseId = !Number.isFinite(courseId) || courseId <= 0

const courseTitle = ref('')
const extraHint = ref('')
const qaModels = ref<AIModelItem[]>([])
const qaModelId = ref('')
const chapters = ref<OutlineChapterDraft[]>([])
const generating = ref(false)
const applying = ref(false)
const error = ref('')
const success = ref('')

// 拉取课程标题与 QA 模型，作为草案生成上下文的基础输入。
async function loadBaseData() {
  if (invalidCourseId) {
    error.value = '课程 ID 无效'
    return
  }
  try {
    const [courseData, modelData] = await Promise.all([
      listTeacherCourses(auth.token ?? '', 1, 100, '', '', 'updated_at', 'desc'),
      listAIModels(auth.token ?? ''),
    ])
    const course = (courseData.items || []).find((c) => Number(c.id) === courseId)
    if (course) courseTitle.value = course.title
    qaModels.value = (modelData.items || []).filter((m) => m.model_type === 'qa')
    if (qaModels.value.length > 0) qaModelId.value = String(qaModels.value[0].id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载基础数据失败'
  }
}

function updateChapterTitle(ci: number, title: string) {
  chapters.value = chapters.value.map((ch, idx) => (idx === ci ? { ...ch, title } : ch))
}

function updateSectionTitle(ci: number, si: number, title: string) {
  chapters.value = chapters.value.map((ch, idx) => {
    if (idx !== ci) return ch
    return {
      ...ch,
      sections: ch.sections.map((sec, sidx) => (sidx === si ? { ...sec, title } : sec)),
    }
  })
}

function addSection(ci: number) {
  chapters.value = chapters.value.map((ch, idx) =>
    idx === ci ? { ...ch, sections: [...ch.sections, { title: '新小节' }] } : ch,
  )
}

// 调用后端生成草案，结果先落在前端可编辑状态中。
async function handleGenerate() {
  if (invalidCourseId) return
  generating.value = true
  error.value = ''
  success.value = ''
  try {
    const out = await generateOutlineDraft(auth.token ?? '', courseId, {
      qa_model_id: qaModelId.value ? Number(qaModelId.value) : undefined,
      extra_hint: extraHint.value.trim() || undefined,
    })
    chapters.value = out.chapters || []
    success.value = '已生成大纲草案，可先编辑再应用'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '生成失败'
  } finally {
    generating.value = false
  }
}

// 仅在用户确认后写入课程，写入成功后跳到课程内容页继续编辑资源。
async function handleApply() {
  if (invalidCourseId || chapters.value.length === 0) return
  applying.value = true
  error.value = ''
  success.value = ''
  try {
    const out = await applyOutlineDraft(auth.token ?? '', courseId, chapters.value)
    success.value = `已追加 ${out.created_chapters} 章、${out.created_sections} 节`
    await router.push(`/teacher/course-content/${courseId}`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '应用失败'
  } finally {
    applying.value = false
  }
}

onMounted(() => void loadBaseData())
</script>

<style scoped>
.workspace-header .action-row {
  margin-bottom: 4px;
}

.card h3 {
  margin: 0;
  font-size: 18px;
}

.card > .subtle {
  margin-top: -2px;
}

.card article.card {
  border-color: #dbe4f0;
  background: #fbfdff;
  box-shadow: none;
}

.card article.card label {
  font-size: 13px;
}

.card article.card .action-row {
  margin-top: 2px;
}

.card article.card input {
  font-size: 14px;
}

.card article.card input:focus {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.18);
}

.success {
  background: #ecfdf5;
  border-color: #86efac;
  color: #166534;
}
</style>
