import { request } from './api'

export type CreateCoursePayload = {
  title: string
  description: string
}

export type TeacherCourseItem = {
  id: string
  title: string
  description: string
  status: 'draft' | 'published' | 'archived'
  cover_image_url?: string
  created_at?: string
  updated_at?: string
}

export type ListTeacherCoursesResp = {
  page: number
  page_size: number
  total: number
  items: TeacherCourseItem[]
}

export type UpdateCoursePayload = {
  title: string
  description: string
  status: 'draft' | 'published' | 'archived'
}

export type CreateChapterPayload = {
  title: string
  sort_order: number
}

export type CreateSectionPayload = {
  title: string
  sort_order: number
}

export type InitUploadPayload = {
  course_id: number
  file_name: string
  resource_type: 'ppt' | 'pdf' | 'txt' | 'video' | 'doc' | 'docx' | 'audio'
}

export type ConfirmResourcePayload = {
  title: string
  resource_type: 'ppt' | 'pdf' | 'txt' | 'video' | 'doc' | 'docx' | 'audio'
  sort_order: number
  object_key: string
  mime_type: string
  size_bytes: number
}

export type ChapterItem = {
  id: string
  course_id: string
  title: string
  sort_order: number
}

export type SectionItem = {
  id: string
  chapter_id: string
  course_id: string
  title: string
  sort_order: number
}

export type ResourceItem = {
  id: string
  chapter_id: string
  section_id: string
  course_id: string
  title: string
  resource_type: 'ppt' | 'pdf' | 'txt' | 'video' | 'doc' | 'docx' | 'audio'
  sort_order: number
  preview_url?: string
  object_url?: string
  mime_type?: string
  size_bytes?: number
  ai_summary?: string | null
  ai_summary_updated_at?: string | null
  ai_summary_status?: 'idle' | 'running' | 'succeeded' | 'failed'
  ai_summary_error?: string | null
}

export type TeacherResourceDetail = {
  object_key: string
  object_url: string
  resource_type: string
  mime_type: string
  ai_summary?: string
  ai_summary_updated_at?: string
}

export type SummarizeResourceResp = {
  summary: string
  updated_at: string
}

export async function createCourse(token: string, payload: CreateCoursePayload) {
  return request<{ id: string }>('/teacher/courses', {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function listTeacherCourses(
  token: string,
  page: number,
  pageSize: number,
  keyword = '',
  status = '',
  sortBy = 'created_at',
  sortOrder = 'desc',
) {
  const query = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
    keyword,
    status,
    sort_by: sortBy,
    sort_order: sortOrder,
  })
  return request<ListTeacherCoursesResp>(`/teacher/courses?${query.toString()}`, {
    method: 'GET',
    token,
  })
}

export async function updateCourse(token: string, courseId: string, payload: UpdateCoursePayload) {
  return request<unknown>(`/teacher/courses/${courseId}`, {
    method: 'PUT',
    token,
    body: payload,
  })
}

export async function createChapter(token: string, courseId: string, payload: CreateChapterPayload) {
  return request<{ id: string }>(`/teacher/courses/${courseId}/chapters`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function createSection(
  token: string,
  courseId: string,
  chapterId: string,
  payload: CreateSectionPayload,
) {
  return request<{ id: string }>(`/teacher/courses/${courseId}/chapters/${chapterId}/sections`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function listCourseSections(token: string, courseId: string, chapterId: string) {
  return request<{ items: SectionItem[] }>(`/teacher/courses/${courseId}/chapters/${chapterId}/sections`, {
    method: 'GET',
    token,
  })
}

export async function deleteSection(token: string, sectionId: string) {
  return request<unknown>(`/teacher/sections/${sectionId}`, {
    method: 'DELETE',
    token,
  })
}

export async function initUpload(token: string, sectionId: string, payload: InitUploadPayload) {
  return request<{ object_key: string; upload_url: string; expire_seconds: number }>(
    `/teacher/sections/${sectionId}/resources/init-upload`,
    { method: 'POST', token, body: payload },
  )
}

export async function confirmResource(token: string, sectionId: string, payload: ConfirmResourcePayload) {
  return request<{ id: string }>(`/teacher/sections/${sectionId}/resources/confirm`, {
    method: 'POST',
    token,
    body: payload,
  })
}

export async function listCourseChapters(token: string, courseId: string) {
  return request<{ items: ChapterItem[] }>(`/teacher/courses/${courseId}/chapters`, {
    method: 'GET',
    token,
  })
}

export async function listSectionResources(token: string, sectionId: string) {
  return request<{ items: ResourceItem[] }>(`/teacher/sections/${sectionId}/resources`, {
    method: 'GET',
    token,
  })
}

export async function getTeacherResourceDetail(token: string, resourceId: string) {
  return request<TeacherResourceDetail>(`/teacher/resources/${resourceId}`, {
    method: 'GET',
    token,
  })
}

export async function summarizeResource(token: string, resourceId: string) {
  return request<SummarizeResourceResp>(`/teacher/resources/${resourceId}/summarize`, {
    method: 'POST',
    token,
    body: {},
  })
}

export async function deleteCourse(token: string, courseId: string) {
  return request<unknown>(`/teacher/courses/${courseId}`, {
    method: 'DELETE',
    token,
  })
}

export async function deleteChapter(token: string, chapterId: string) {
  return request<unknown>(`/teacher/chapters/${chapterId}`, {
    method: 'DELETE',
    token,
  })
}

export async function deleteResource(token: string, resourceId: string) {
  return request<unknown>(`/teacher/resources/${resourceId}`, {
    method: 'DELETE',
    token,
  })
}
