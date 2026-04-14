import { request, unwrapListItems } from './api'
import { getAccessToken } from './auth'
import type {
  AIModelListItem,
  CreateAIModelReq,
  CreateAIModelResp,
  UpdateAIModelReq,
  TestAIModelConnectionReq,
  TestAIModelConnectionResp,
} from '../types'

function requireToken(): string {
  const t = getAccessToken()
  if (!t) throw new Error('未登录或登录已过期')
  return t
}

/**
 * 获取当前教师的 AI 模型列表
 * GET /api/v1/teacher/ai-models
 */
export async function getModels(): Promise<AIModelListItem[]> {
  const token = requireToken()
  const data = await request<AIModelListItem[] | { items: AIModelListItem[] }>('/teacher/ai-models', {
    method: 'GET',
    token,
  })
  return unwrapListItems(data)
}

/**
 * 创建新的 AI 模型配置
 */
export async function createModel(data: CreateAIModelReq): Promise<CreateAIModelResp> {
  const token = requireToken()
  return request<CreateAIModelResp>('/teacher/ai-models', {
    method: 'POST',
    token,
    body: data,
  })
}

/**
 * 更新 AI 模型配置
 */
export async function updateModel(modelId: string, data: UpdateAIModelReq): Promise<unknown> {
  const token = requireToken()
  return request(`/teacher/ai-models/${modelId}`, {
    method: 'PUT',
    token,
    body: data,
  })
}

/**
 * 删除 AI 模型配置
 */
export async function deleteModel(modelId: string): Promise<void> {
  const token = requireToken()
  return request<void>(`/teacher/ai-models/${modelId}`, { method: 'DELETE', token })
}

/**
 * 测试 AI 模型连通性
 */
export async function testConnection(data: TestAIModelConnectionReq): Promise<TestAIModelConnectionResp> {
  const token = requireToken()
  const body: Record<string, unknown> = {
    model_type: data.model_type,
    api_base_url: data.api_base_url,
    model_id: data.model_id,
  }
  if (data.api_key !== undefined && data.api_key !== '') {
    body.api_key = data.api_key
  }
  if (data.existing_model_id !== undefined) {
    body.existing_model_id = data.existing_model_id
  }

  return request<TestAIModelConnectionResp>('/teacher/ai-models/test-connection', {
    method: 'POST',
    token,
    body,
  })
}
