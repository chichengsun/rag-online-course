import { request, unwrapListItems } from './api'
import type {
  Resource,
  InitUploadReq,
  InitUploadResp,
  ConfirmResourceReq,
  ConfirmResourceResp,
  ParseResourceResp,
  SummarizeResourceResp,
  PreviewResourceURLResp,
} from '../types'

export async function initUpload(
  token: string,
  sectionId: number,
  data: InitUploadReq,
): Promise<InitUploadResp> {
  return request<InitUploadResp>(`/teacher/sections/${sectionId}/resources/init-upload`, {
    method: 'POST',
    token,
    body: data,
  })
}

export async function confirmResource(
  token: string,
  sectionId: number,
  data: ConfirmResourceReq,
): Promise<ConfirmResourceResp> {
  return request<ConfirmResourceResp>(`/teacher/sections/${sectionId}/resources/confirm`, {
    method: 'POST',
    token,
    body: data,
  })
}

export async function getResource(token: string, resourceId: number): Promise<Resource> {
  return request<Resource>(`/teacher/resources/${resourceId}`, {
    method: 'GET',
    token,
  })
}

export async function getPreviewUrl(token: string, resourceId: number): Promise<string> {
  const resp = await request<PreviewResourceURLResp>(`/teacher/resources/${resourceId}/preview-url`, {
    method: 'GET',
    token,
  })
  return resp.preview_url
}

export async function parseResource(
  token: string,
  resourceId: number,
): Promise<ParseResourceResp> {
  return request<ParseResourceResp>(`/teacher/resources/${resourceId}/parse`, {
    method: 'POST',
    token,
    body: {},
  })
}

export async function summarizeResource(
  token: string,
  resourceId: number,
): Promise<SummarizeResourceResp> {
  return request<SummarizeResourceResp>(`/teacher/resources/${resourceId}/summarize`, {
    method: 'POST',
    token,
    body: {},
  })
}

export async function deleteResource(token: string, resourceId: number): Promise<void> {
  return request<void>(`/teacher/resources/${resourceId}`, {
    method: 'DELETE',
    token,
  })
}

export async function getSectionResources(
  token: string,
  sectionId: number,
): Promise<Resource[]> {
  const data = await request<Resource[] | { items: Resource[] }>(`/teacher/sections/${sectionId}/resources`, {
    method: 'GET',
    token,
  })
  return unwrapListItems(data)
}

export async function reorderResource(
  token: string,
  resourceId: number,
  sortOrder: number,
): Promise<Resource> {
  return request<Resource>(`/teacher/resources/${resourceId}/reorder`, {
    method: 'PUT',
    token,
    body: { sort_order: sortOrder },
  })
}

/** 更新资源标题；后端返回 204 无 body。 */
export async function updateResource(
  token: string,
  resourceId: number,
  title: string,
): Promise<void> {
  await request<unknown>(`/teacher/resources/${resourceId}`, {
    method: 'PUT',
    token,
    body: { title },
  })
}
