import { request } from './api'

export type ChatSessionItem = {
  id: string
  course_id: string
  title: string
  created_at: string
  updated_at: string
  message_count: number | string
  last_message_at?: string
}

export type ListChatSessionsResp = {
  page: number
  page_size: number
  total: number | string
  items: ChatSessionItem[]
}

export type ChatMessageItem = {
  id: string
  session_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  references_json?: Array<{
    citation_no?: number
    resource_title?: string
    chunk_index?: number
    snippet?: string
    full_content?: string
  }>
  created_at: string
}

export type ListChatMessagesResp = {
  page: number
  page_size: number
  total: number | string
  items: ChatMessageItem[]
}

export async function createChatSession(token: string, courseId: string, title: string) {
  return request<{ id: string }>(`/teacher/courses/${courseId}/knowledge/chats/sessions`, {
    method: 'POST',
    token,
    body: { title },
  })
}

export async function listChatSessions(token: string, page: number, pageSize: number, courseId?: string) {
  const q = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
  if (courseId) q.set('course_id', courseId)
  return request<ListChatSessionsResp>(`/teacher/knowledge/chats/sessions?${q.toString()}`, {
    method: 'GET',
    token,
  })
}

export async function listChatMessages(token: string, sessionId: string, page = 1, pageSize = 100) {
  const q = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
  return request<ListChatMessagesResp>(`/teacher/knowledge/chats/sessions/${sessionId}/messages?${q.toString()}`, {
    method: 'GET',
    token,
  })
}

export async function updateChatSessionTitle(token: string, sessionId: string, title: string) {
  return request<unknown>(`/teacher/knowledge/chats/sessions/${sessionId}`, {
    method: 'PATCH',
    token,
    body: { title },
  })
}

export async function deleteChatSession(token: string, sessionId: string) {
  return request<unknown>(`/teacher/knowledge/chats/sessions/${sessionId}`, {
    method: 'DELETE',
    token,
  })
}

export async function askInSession(
  token: string,
  sessionId: string,
  question: string,
  topK = 8,
  qaModelId?: string,
  semanticMinScore = 0,
  keywordMinScore = 0,
) {
  return request<{
    session_id: string
    answer: string
    references: Array<{
      citation_no: number
      chunk_id: string
      resource_id: string
      resource_title: string
      chunk_index: number
      score: number
      snippet: string
      full_content: string
    }>
  }>(`/teacher/knowledge/chats/sessions/${sessionId}/ask`, {
    method: 'POST',
    token,
    body: {
      question,
      top_k: topK,
      use_rerank: true,
      qa_model_id: qaModelId ? Number(qaModelId) : undefined,
      semantic_min_score: semanticMinScore,
      keyword_min_score: keywordMinScore,
    },
  })
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export async function askInSessionStream(
  token: string,
  sessionId: string,
  question: string,
  topK: number,
  qaModelId: string | undefined,
  semanticMinScore: number,
  keywordMinScore: number,
  handlers: {
    onToken: (token: string) => void
    onReferences?: (
      refs: Array<{
        citation_no: number
        chunk_id: string
        resource_id: string
        resource_title: string
        chunk_index: number
        score: number
        snippet: string
        full_content: string
      }>,
    ) => void
    onDone?: (data: { session_id: string; user_message_id: string; assistant_message_id: string }) => void
    onError?: (message: string) => void
  },
) {
  const resp = await fetch(`${API_BASE_URL}/teacher/knowledge/chats/sessions/${sessionId}/ask/stream`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      question,
      top_k: topK,
      use_rerank: true,
      qa_model_id: qaModelId ? Number(qaModelId) : undefined,
      semantic_min_score: semanticMinScore,
      keyword_min_score: keywordMinScore,
    }),
  })
  if (!resp.ok) {
    const text = await resp.text()
    throw new Error(text || `请求失败（${resp.status}）`)
  }
  if (!resp.body) throw new Error('stream body is empty')

  const reader = resp.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const parts = buffer.split('\n\n')
    buffer = parts.pop() ?? ''
    for (const block of parts) {
      const lines = block.split('\n')
      let event = 'message'
      let dataText = ''
      for (const line of lines) {
        if (line.startsWith('event:')) event = line.slice(6).trim()
        if (line.startsWith('data:')) dataText += line.slice(5).trim()
      }
      if (!dataText) continue
      let data: any = {}
      try {
        data = JSON.parse(dataText)
      } catch {
        data = { message: dataText }
      }
      if (event === 'token' && data.token) handlers.onToken(data.token)
      if (event === 'references' && Array.isArray(data.references)) handlers.onReferences?.(data.references)
      if (event === 'done') handlers.onDone?.(data)
      if (event === 'error') handlers.onError?.(data.message || '流式问答失败')
    }
  }
}
