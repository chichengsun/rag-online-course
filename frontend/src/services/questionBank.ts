import { request, type ApiEnvelope } from './api'

export type QuestionBankItem = {
  id: string
  course_id: string
  question_type: string
  stem: string
  reference_answer: string
  source_file_name?: string
  created_at?: string
  updated_at?: string
}

export type QuestionBankListResp = {
  items: QuestionBankItem[]
  total: number
  page: number
  page_size: number
}

export type CreateQuestionPayload = {
  question_type: string
  stem: string
  reference_answer: string
}

export type QuestionBankListQuery = {
  page?: number
  page_size?: number
  keyword?: string
  question_type?: string
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export async function listQuestionBank(token: string, courseId: string, q: QuestionBankListQuery = {}) {
  const sp = new URLSearchParams()
  if (q.page != null && q.page > 0) sp.set('page', String(q.page))
  if (q.page_size != null && q.page_size > 0) sp.set('page_size', String(q.page_size))
  if (q.keyword?.trim()) sp.set('keyword', q.keyword.trim())
  if (q.question_type?.trim()) sp.set('question_type', q.question_type.trim())
  const qs = sp.toString()
  const path = `/teacher/courses/${courseId}/question-bank${qs ? `?${qs}` : ''}`
  return request<QuestionBankListResp>(path, { method: 'GET', token })
}

export async function createQuestion(token: string, courseId: string, payload: CreateQuestionPayload) {
  return request<{ id: string }>(`/teacher/courses/${courseId}/question-bank`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function updateQuestion(token: string, itemId: string, payload: CreateQuestionPayload) {
  return request<unknown>(`/teacher/question-bank/items/${itemId}`, {
    method: 'PUT',
    token,
    body: payload,
  })
}

export async function deleteQuestion(token: string, itemId: string) {
  return request<unknown>(`/teacher/question-bank/items/${itemId}`, {
    method: 'DELETE',
    token,
  })
}

type ParseImportData = { questions: CreateQuestionPayload[] }

/** 上传文本并由模型解析为草稿（不入库）。 */
export async function parseImportFile(token: string, courseId: string, file: File, qaModelId?: string) {
  const form = new FormData()
  form.append('file', file)
  if (qaModelId) form.append('qa_model_id', qaModelId)
  const resp = await fetch(`${API_BASE_URL}/teacher/courses/${courseId}/question-bank/import/parse`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    body: form,
  })
  const json = (await resp.json()) as ApiEnvelope<ParseImportData>
  if (!resp.ok) {
    throw new Error(json?.message || `解析失败（${resp.status}）`)
  }
  return (json?.data?.questions ?? []) as CreateQuestionPayload[]
}

/** 确认后将题目批量写入题库。 */
export async function confirmImportBatch(
  token: string,
  courseId: string,
  body: { source_file_name?: string; questions: CreateQuestionPayload[] },
) {
  return request<{ created_count: number }>(`/teacher/courses/${courseId}/question-bank/import/confirm`, {
    method: 'POST',
    token,
    body: {
      source_file_name: body.source_file_name ?? '',
      questions: body.questions,
    },
  })
}
