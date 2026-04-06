import { request } from './api'

export type KnowledgeResourceRow = {
  id: string
  chapter_id: string
  chapter_title: string
  title: string
  resource_type: string
  total_chunk_chars: number | string
  chunk_count: number | string
  embedded_count: number | string
}

export type ListKnowledgeResourcesResp = {
  page: number
  page_size: number
  total: number | string
  items: KnowledgeResourceRow[]
}

export type ChunkPreviewSegment = {
  index: number
  content: string
  char_start: number
  char_end: number
}

export type ChunkPreviewResp = {
  segments: ChunkPreviewSegment[]
}

export type ChunkSaveItem = {
  content: string
  char_start?: number
  char_end?: number
}

export async function listKnowledgeResources(token: string, courseId: string, page: number, pageSize: number) {
  const q = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
  return request<ListKnowledgeResourcesResp>(`/teacher/courses/${courseId}/knowledge/resources?${q}`, {
    method: 'GET',
    token,
  })
}

export async function chunkPreview(
  token: string,
  resourceId: string,
  chunkSize: number,
  overlap: number,
  clearPersistedFirst: boolean,
) {
  return request<ChunkPreviewResp>(`/teacher/resources/${resourceId}/knowledge/chunk-preview`, {
    method: 'POST',
    token,
    body: { chunk_size: chunkSize, overlap, clear_persisted_first: clearPersistedFirst },
  })
}

export async function saveKnowledgeChunks(token: string, resourceId: string, chunks: ChunkSaveItem[]) {
  return request<unknown>(`/teacher/resources/${resourceId}/knowledge/chunks`, {
    method: 'PUT',
    token,
    body: { chunks },
  })
}

export async function confirmKnowledgeChunks(token: string, resourceId: string) {
  return request<{ confirmed_count: number }>(`/teacher/resources/${resourceId}/knowledge/chunks/confirm`, {
    method: 'POST',
    token,
  })
}

export type SavedChunkRow = {
  id: string
  chunk_index: number
  content: string
  char_start?: number | null
  char_end?: number | null
  confirmed_at?: string | null
  embedded_at?: string | null
  embedding_dims?: number | null
}

export async function listKnowledgeChunks(token: string, resourceId: string) {
  return request<{ items: SavedChunkRow[] }>(`/teacher/resources/${resourceId}/knowledge/chunks`, {
    method: 'GET',
    token,
  })
}

export async function clearKnowledgeChunks(token: string, resourceId: string) {
  return request<unknown>(`/teacher/resources/${resourceId}/knowledge/chunks`, {
    method: 'DELETE',
    token,
  })
}

export async function updateKnowledgeChunk(
  token: string,
  resourceId: string,
  chunkId: string,
  body: { content: string; char_start?: number; char_end?: number },
) {
  return request<unknown>(`/teacher/resources/${resourceId}/knowledge/chunks/${chunkId}`, {
    method: 'PATCH',
    token,
    body,
  })
}

export async function deleteKnowledgeChunk(token: string, resourceId: string, chunkId: string) {
  return request<unknown>(`/teacher/resources/${resourceId}/knowledge/chunks/${chunkId}`, {
    method: 'DELETE',
    token,
  })
}

export async function embedResource(token: string, resourceId: string, modelId: string) {
  return request<{ embedded_count: number }>(`/teacher/resources/${resourceId}/knowledge/embed`, {
    method: 'POST',
    token,
    body: { model_id: Number(modelId) },
  })
}
