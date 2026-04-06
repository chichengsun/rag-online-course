export type ApiEnvelope<T> = {
  code: number
  message: string
  data?: T
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

type RequestOptions = {
  method?: string
  token?: string | null
  body?: unknown
}

export class ApiError extends Error {
  code: number

  constructor(message: string, code = 500) {
    super(message)
    this.code = code
  }
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', token, body } = options
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) headers.Authorization = `Bearer ${token}`

  const resp = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  const text = await resp.text()
  let parsed: ApiEnvelope<T> | null = null
  try {
    parsed = text ? (JSON.parse(text) as ApiEnvelope<T>) : null
  } catch {
    parsed = null
  }

  if (!resp.ok) {
    const message = parsed?.message || `请求失败（${resp.status}）`
    throw new ApiError(message, resp.status)
  }
  return (parsed?.data ?? ({} as T)) as T
}
