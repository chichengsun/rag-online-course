import { request } from './api'

export type AIModelType = 'qa' | 'embedding' | 'rerank'

export type AIModelItem = {
  id: string
  name: string
  model_type: AIModelType
  api_base_url: string
  model_id: string
  has_api_key: boolean
}

export type CreateAIModelPayload = {
  name: string
  model_type: AIModelType
  api_base_url: string
  model_id: string
  api_key: string
}

export type UpdateAIModelPayload = {
  name: string
  api_base_url: string
  model_id: string
  api_key: string
}

export async function listAIModels(token: string) {
  return request<{ items: AIModelItem[] }>('/teacher/ai-models', { method: 'GET', token })
}

export async function createAIModel(token: string, payload: CreateAIModelPayload) {
  return request<{ id: string }>('/teacher/ai-models', { method: 'POST', token, body: payload })
}

export async function updateAIModel(token: string, modelId: string, payload: UpdateAIModelPayload) {
  return request<unknown>(`/teacher/ai-models/${modelId}`, { method: 'PUT', token, body: payload })
}

export async function deleteAIModel(token: string, modelId: string) {
  return request<unknown>(`/teacher/ai-models/${modelId}`, { method: 'DELETE', token })
}

export type TestAIModelConnectionPayload = {
  model_type: AIModelType
  api_base_url: string
  model_id: string
  api_key?: string
  /** 编辑已保存模型且未改填 API Key 时传入，用于使用库中密钥 */
  existing_model_id?: string
}

export type TestAIModelConnectionResp = {
  ok: boolean
  message: string
  http_status?: number
}

export async function testAIModelConnection(token: string, payload: TestAIModelConnectionPayload) {
  const body: Record<string, unknown> = {
    model_type: payload.model_type,
    api_base_url: payload.api_base_url,
    model_id: payload.model_id,
  }
  if (payload.api_key !== undefined && payload.api_key !== '') body.api_key = payload.api_key
  if (payload.existing_model_id) body.existing_model_id = Number(payload.existing_model_id)
  return request<TestAIModelConnectionResp>('/teacher/ai-models/test-connection', {
    method: 'POST',
    token,
    body,
  })
}
