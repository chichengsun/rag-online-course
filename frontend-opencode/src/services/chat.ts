/**
 * 知识库对话 HTTP 封装：路径与后端 internal/api/router.go 一致，统一走 Bearer Token + 统一信封。
 */
import { request } from './api'
import { getAccessToken } from './auth'
import type {
  AskInSessionReq,
  AskInSessionResp,
  ChatMessage,
  ChatSession,
  CreateChatSessionReq,
  CreateChatSessionResp,
  ListChatMessagesResp,
  ListChatSessionsResp,
  ReferenceItem,
  UpdateChatSessionReq,
} from '@/types'

function requireToken(): string {
  const t = getAccessToken()
  if (!t) {
    throw new Error('未登录或登录已过期')
  }
  return t
}

/** 将仓储返回的 references_json（对象/字符串/数组）规范为 ReferenceItem[]。 */
function parseMessageReferences(raw: unknown): ReferenceItem[] {
  if (raw == null) return []
  if (Array.isArray(raw)) return raw as ReferenceItem[]
  if (typeof raw === 'string') {
    try {
      const j = JSON.parse(raw) as unknown
      return Array.isArray(j) ? (j as ReferenceItem[]) : []
    } catch {
      return []
    }
  }
  return []
}

/** 创建会话：POST /teacher/courses/:courseId/knowledge/chats/sessions */
export async function createSession(
  courseId: string,
  data: Partial<CreateChatSessionReq> = {},
): Promise<CreateChatSessionResp> {
  const token = requireToken()
  const body: CreateChatSessionReq = {
    title: data.title?.trim() ?? '',
  }
  return request<CreateChatSessionResp>(`/teacher/courses/${courseId}/knowledge/chats/sessions`, {
    method: 'POST',
    token,
    body,
  })
}

/** 分页会话列表：GET /teacher/knowledge/chats/sessions */
export async function getSessions(courseId?: string): Promise<ChatSession[]> {
  const token = requireToken()
  const q = new URLSearchParams({ page: '1', page_size: '100' })
  if (courseId) {
    q.set('course_id', courseId)
  }
  const data = await request<ListChatSessionsResp>(`/teacher/knowledge/chats/sessions?${q.toString()}`, {
    method: 'GET',
    token,
  })
  return data.items ?? []
}

/** 更新会话标题 */
export async function updateSession(sessionId: string, data: UpdateChatSessionReq): Promise<void> {
  const token = requireToken()
  await request<unknown>(`/teacher/knowledge/chats/sessions/${sessionId}`, {
    method: 'PATCH',
    token,
    body: data,
  })
}

/** 删除会话 */
export async function deleteSession(sessionId: string): Promise<void> {
  const token = requireToken()
  await request<unknown>(`/teacher/knowledge/chats/sessions/${sessionId}`, {
    method: 'DELETE',
    token,
  })
}

/** 会话消息列表（首屏拉足一页） */
export async function getMessages(sessionId: string): Promise<ChatMessage[]> {
  const token = requireToken()
  const q = new URLSearchParams({ page: '1', page_size: '200' })
  const data = await request<ListChatMessagesResp>(
    `/teacher/knowledge/chats/sessions/${sessionId}/messages?${q.toString()}`,
    {
      method: 'GET',
      token,
    },
  )
  const items = (data.items ?? []) as unknown[]
  return items.map((row) => {
    const r = row as Record<string, unknown>
    const refs = r.references ?? r.references_json
    return {
      id: String(r.id ?? ''),
      session_id: String(r.session_id ?? sessionId),
      role: r.role as 'user' | 'assistant',
      content: String(r.content ?? ''),
      references: parseMessageReferences(refs),
      created_at: String(r.created_at ?? ''),
    } as ChatMessage
  })
}

/** 非流式提问 */
export async function askInSession(sessionId: string, data: AskInSessionReq): Promise<AskInSessionResp> {
  const token = requireToken()
  return request<AskInSessionResp>(`/teacher/knowledge/chats/sessions/${sessionId}/ask`, {
    method: 'POST',
    token,
    body: data,
  })
}

export type StreamChatCallbacks = {
  onToken?: (token: string) => void
  onReferences?: (references: ReferenceItem[]) => void
  onDone?: (sessionId: string, messageId: string) => void
  onError?: (err: unknown) => void
}

/**
 * 流式问答：POST /teacher/knowledge/chats/sessions/:id/ask/stream（SSE）
 * 返回 Promise，在收到 done 或读尽流后 resolve；此前 `void` 导致调用方立即 loadMessages 冲掉流式 UI。
 */
export function askInSessionStream(
  sessionId: string,
  data: AskInSessionReq,
  callbacks: StreamChatCallbacks,
): Promise<void> {
  const token = getAccessToken()
  if (!token) {
    callbacks.onError?.(new Error('未登录或登录已过期'))
    return Promise.reject(new Error('未登录或登录已过期'))
  }

  const url = `/api/v1/teacher/knowledge/chats/sessions/${sessionId}/ask/stream`

  return new Promise((resolve, reject) => {
    let settled = false
    const finish = (fn: () => void) => {
      if (settled) return
      settled = true
      fn()
    }

    const userOnDone = callbacks.onDone
    const userOnError = callbacks.onError
    const wrapped: StreamChatCallbacks = {
      ...callbacks,
      onDone: (sid, mid) => {
        userOnDone?.(sid, mid)
        finish(() => resolve())
      },
      onError: (err) => {
        userOnError?.(err)
        finish(() =>
          reject(err instanceof Error ? err : new Error(typeof err === 'string' ? err : '流式响应错误')),
        )
      },
    }

    void (async () => {
      try {
        const res = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify(data),
        })

        if (!res.ok) {
          const text = await res.text()
          let msg = text || `流式请求失败（${res.status}）`
          try {
            const j = JSON.parse(text) as { message?: string }
            if (j?.message) msg = j.message
          } catch {
            /* 非 JSON 错误体 */
          }
          wrapped.onError?.(new Error(msg))
          return
        }

        const reader = (res.body as ReadableStream<Uint8Array>).getReader()
        const decoder = new TextDecoder('utf-8')
        let buffer = ''

        const processChunk = (chunk: string) => {
          buffer += chunk
          const parts = buffer.split(/\r?\n\r?\n/)
          buffer = parts.pop() || ''
          for (const part of parts) {
            const lines = part.split(/\r?\n/)
            let eventType: string | null = null
            let dataStr = ''
            for (const rawLine of lines) {
              const line = rawLine.replace(/^\uFEFF/, '').trimEnd()
              if (line.startsWith('event:')) {
                eventType = line.slice(6).trim()
              } else if (line.startsWith('data:')) {
                dataStr += line.slice(5).trim()
              }
            }
            if (!eventType) continue
            try {
              const payload = dataStr ? (JSON.parse(dataStr) as unknown) : null
              if (eventType === 'token' && wrapped.onToken && payload && typeof payload === 'object' && 'token' in payload) {
                wrapped.onToken(String((payload as { token: string }).token))
              } else if (eventType === 'references' && wrapped.onReferences) {
                const raw = payload as { references?: ReferenceItem[] } | ReferenceItem[] | null
                const list = Array.isArray(raw) ? raw : raw?.references
                wrapped.onReferences(Array.isArray(list) ? list : [])
              } else if (eventType === 'done' && wrapped.onDone) {
                const p = payload as Record<string, string> | null
                const sid = p?.session_id ?? p?.sessionId ?? String(sessionId)
                const mid = p?.assistant_message_id ?? p?.messageId ?? ''
                wrapped.onDone(sid, mid)
              } else if (eventType === 'error' && wrapped.onError) {
                const msg =
                  payload && typeof payload === 'object' && payload !== null && 'message' in payload
                    ? String((payload as { message: string }).message)
                    : '流式响应错误'
                wrapped.onError(new Error(msg))
              }
            } catch (e) {
              wrapped.onError?.(e)
            }
          }
        }

        for (;;) {
          const { value, done } = await reader.read()
          if (done) break
          processChunk(decoder.decode(value, { stream: true }))
        }
        processChunk(decoder.decode())
        if (!settled) {
          finish(() => resolve())
        }
      } catch (err) {
        if (!settled) {
          finish(() =>
            reject(err instanceof Error ? err : new Error(String(err))),
          )
        }
      }
    })()
  })
}
