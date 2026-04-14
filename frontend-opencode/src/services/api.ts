export type ApiEnvelope<T> = {
  code: number
  message: string
  data?: T
}

type RequestOptions = {
  method?: string
  token?: string | null
  body?: unknown
}

export class ApiError extends Error {
  code: number
  status: number

  constructor(message: string, code = 500, status = 500) {
    super(message)
    this.code = code
    this.status = status
  }
}

/**
 * 将后端 ListXxxResp（形如 { items: T[] }）与历史裸数组统一为 T[]，
 * 避免页面把 data 当成数组调用 .map / .length 导致运行时报错。
 */
export function unwrapListItems<T>(data: T[] | { items?: T[] } | null | undefined): T[] {
  if (data == null) return []
  if (Array.isArray(data)) return data
  const items = (data as { items?: T[] }).items
  return Array.isArray(items) ? items : []
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', token, body } = options

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const resp = await fetch(`/api/v1${path}`, {
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
    throw new ApiError(message, parsed?.code ?? resp.status, resp.status)
  }

  return (parsed?.data ?? ({} as T)) as T
}
