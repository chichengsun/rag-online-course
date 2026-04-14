/**
 * 知识库服务层
 * 提供知识资源的分块管理、嵌入向量等操作
 */
import { request } from './api'
import { getAccessToken } from './auth'
import type {
  KnowledgeChunk,
  ChunkPreviewReq,
  ChunkPreviewResp,
  SaveKnowledgeChunksReq,
  EmbedResourceReq,
  EmbedResourceResp,
  ConfirmKnowledgeChunksResp,
  ListKnowledgeChunksResp,
  UpdateKnowledgeChunkReq,
  ListKnowledgeResourcesResp,
} from '../types'

/**
 * 获取知识资源列表（分页参数与后端 ListKnowledgeResources 一致）。
 */
export async function getKnowledgeResources(
  courseId: number,
  params: { page?: number; page_size?: number } = {},
): Promise<ListKnowledgeResourcesResp> {
  const token = getAccessToken()
  const q = new URLSearchParams()
  if (params.page != null) q.set('page', String(params.page))
  if (params.page_size != null) q.set('page_size', String(params.page_size))
  const suffix = q.toString() ? `?${q}` : ''
  return request<ListKnowledgeResourcesResp>(`/teacher/courses/${courseId}/knowledge/resources${suffix}`, {
    method: 'GET',
    token,
  })
}

/**
 * 预览资源分块
 * @param resourceId 资源ID
 * @param data 分块预览请求参数
 * @returns 分块预览片段列表
 */
export async function chunkPreview(
  resourceId: number,
  data: ChunkPreviewReq,
): Promise<ChunkPreviewResp> {
  const token = getAccessToken()
  return request<ChunkPreviewResp>(`/teacher/resources/${resourceId}/knowledge/chunk-preview`, {
    method: 'POST',
    token,
    body: data,
  })
}

/**
 * 保存资源分块
 * @param resourceId 资源ID
 * @param data 保存分块请求参数
 */
export async function saveChunks(
  resourceId: number,
  data: SaveKnowledgeChunksReq,
): Promise<void> {
  const token = getAccessToken()
  return request<void>(`/teacher/resources/${resourceId}/knowledge/chunks`, {
    method: 'PUT',
    token,
    body: data,
  })
}

/**
 * 确认资源分块
 * @param resourceId 资源ID
 */
export async function confirmChunks(resourceId: number): Promise<void> {
  const token = getAccessToken()
  return request<ConfirmKnowledgeChunksResp>(`/teacher/resources/${resourceId}/knowledge/chunks/confirm`, {
    method: 'POST',
    token,
  }).then(() => undefined)
}

/**
 * 获取资源的分块列表
 * @param resourceId 资源ID
 * @returns 分块列表
 */
/** 将列表接口返回的 map 行规范为前端使用的 KnowledgeChunk（id 与后端 id::text 对齐为 string）。 */
function normalizeKnowledgeChunk(row: Record<string, unknown>): KnowledgeChunk {
  const meta = row.metadata_json ?? row.metadata
  return {
    id: String(row.id ?? ''),
    resource_id: row.resource_id != null ? Number(row.resource_id) : 0,
    course_id: row.course_id != null ? Number(row.course_id) : 0,
    chunk_index: Number(row.chunk_index ?? 0),
    content: String(row.content ?? ''),
    char_start: row.char_start != null && row.char_start !== '' ? Number(row.char_start) : null,
    char_end: row.char_end != null && row.char_end !== '' ? Number(row.char_end) : null,
    metadata: typeof meta === 'object' && meta !== null && !Array.isArray(meta) ? (meta as Record<string, unknown>) : {},
    is_embedded: row.embedded_at != null && String(row.embedded_at) !== '',
    is_confirmed: row.confirmed_at != null && String(row.confirmed_at) !== '',
    created_at: String(row.created_at ?? ''),
    updated_at: String(row.updated_at ?? ''),
  }
}

export async function getChunks(resourceId: number): Promise<KnowledgeChunk[]> {
  const token = getAccessToken()
  const data = await request<ListKnowledgeChunksResp>(`/teacher/resources/${resourceId}/knowledge/chunks`, {
    method: 'GET',
    token,
  })
  const items = Array.isArray(data.items) ? data.items : []
  return items.map((row) => normalizeKnowledgeChunk(row as Record<string, unknown>))
}

/**
 * 嵌入资源（生成向量嵌入）
 * @param resourceId 资源ID
 * @param data 嵌入请求参数
 */
export async function embedResource(
  resourceId: number,
  data: EmbedResourceReq,
): Promise<void> {
  const token = getAccessToken()
  return request<EmbedResourceResp>(`/teacher/resources/${resourceId}/knowledge/embed`, {
    method: 'POST',
    token,
    body: data,
  }).then(() => undefined)
}

/**
 * 删除分块
 * @param resourceId 资源ID
 * @param chunkId 分块ID
 */
export async function deleteChunk(
  resourceId: number,
  chunkId: number | string,
): Promise<void> {
  const token = getAccessToken()
  return request<void>(`/teacher/resources/${resourceId}/knowledge/chunks/${chunkId}`, {
    method: 'DELETE',
    token,
  })
}

/**
 * 更新分块
 * @param resourceId 资源ID
 * @param chunkId 分块ID
 * @param data 更新分块请求参数
 * @returns 更新后的分块
 */
export async function updateChunk(
  resourceId: number,
  chunkId: number | string,
  data: UpdateKnowledgeChunkReq,
): Promise<void> {
  const token = getAccessToken()
  await request<unknown>(`/teacher/resources/${resourceId}/knowledge/chunks/${chunkId}`, {
    method: 'PATCH',
    token,
    body: data,
  })
}
