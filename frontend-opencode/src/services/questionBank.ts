import { request } from './api'

export type QuestionBankItem = {
  id: string
  course_id: string
  question_type: QuestionType
  stem: string
  reference_answer: string
  source_file_name?: string
  created_at?: string
  updated_at?: string
}

export type QuestionBankListResult = {
  items: QuestionBankItem[]
  total: number
  page: number
  page_size: number
}

type QuestionBankListResp = {
  items: QuestionBankItem[]
  total: number
  page: number
  page_size: number
}

export type QuestionBankPayload = {
  question_type: QuestionType
  stem: string
  reference_answer: string
}

export type QuestionType =
  | 'single_choice'
  | 'multiple_choice'
  | 'true_false'
  | 'short_answer'
  | 'fill_blank'

export const QUESTION_TYPE_OPTIONS: Array<{ value: QuestionType; label: string }> = [
  { value: 'single_choice', label: '单选题' },
  { value: 'multiple_choice', label: '多选题' },
  { value: 'true_false', label: '判断题' },
  { value: 'short_answer', label: '简答题' },
  { value: 'fill_blank', label: '填空题' },
]

export const QUESTION_TYPE_LABELS: Record<QuestionType, string> = QUESTION_TYPE_OPTIONS.reduce(
  (acc, item) => {
    acc[item.value] = item.label
    return acc
  },
  {} as Record<QuestionType, string>,
)

export type QuestionBankListQuery = {
  page?: number
  page_size?: number
  keyword?: string
  question_type?: QuestionType | ''
}

/** 分页查询已入库题目（服务端 keyword / question_type 筛选）。 */
export async function listQuestionBank(
  token: string,
  courseId: number,
  q: QuestionBankListQuery = {},
): Promise<QuestionBankListResult> {
  const sp = new URLSearchParams()
  if (q.page != null && q.page > 0) sp.set('page', String(q.page))
  if (q.page_size != null && q.page_size > 0) sp.set('page_size', String(q.page_size))
  if (q.keyword?.trim()) sp.set('keyword', q.keyword.trim())
  if (q.question_type) sp.set('question_type', q.question_type)
  const qs = sp.toString()
  const path = `/teacher/courses/${courseId}/question-bank${qs ? `?${qs}` : ''}`
  const data = await request<QuestionBankListResp>(path, { method: 'GET', token })
  return {
    items: data.items || [],
    total: Number(data.total ?? 0),
    page: Number(data.page ?? 1),
    page_size: Number(data.page_size ?? 20),
  }
}

export async function createQuestion(
  token: string,
  courseId: number,
  payload: QuestionBankPayload,
): Promise<{ id: string }> {
  return request<{ id: string }>(`/teacher/courses/${courseId}/question-bank`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function updateQuestion(token: string, itemId: number, payload: QuestionBankPayload): Promise<void> {
  await request<void>(`/teacher/question-bank/items/${itemId}`, {
    method: 'PUT',
    token,
    body: payload,
  })
}

export async function deleteQuestion(token: string, itemId: number): Promise<void> {
  await request<void>(`/teacher/question-bank/items/${itemId}`, {
    method: 'DELETE',
    token,
  })
}

type ParseImportResp = { questions: QuestionBankPayload[] }

/** 上传文本并由模型解析为草稿题目（不入库）。 */
export async function parseImportFile(token: string, courseId: number, file: File): Promise<QuestionBankPayload[]> {
  const form = new FormData()
  form.append('file', file)
  const resp = await fetch(`/api/v1/teacher/courses/${courseId}/question-bank/import/parse`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    body: form,
  })
  const json = await resp.json()
  if (!resp.ok) {
    throw new Error(json?.message || `解析失败（${resp.status}）`)
  }
  const data = json?.data as ParseImportResp | undefined
  return Array.isArray(data?.questions) ? data!.questions : []
}

type ConfirmImportResp = { created_count: number; items: QuestionBankItem[] }

/** 将确认后的题目批量写入题库。 */
export async function confirmImportBatch(
  token: string,
  courseId: number,
  body: { source_file_name?: string; questions: QuestionBankPayload[] },
): Promise<{ created_count: number }> {
  const data = await request<ConfirmImportResp>(`/teacher/courses/${courseId}/question-bank/import/confirm`, {
    method: 'POST',
    token,
    body: {
      source_file_name: body.source_file_name ?? '',
      questions: body.questions,
    },
  })
  return { created_count: Number(data?.created_count ?? 0) }
}
