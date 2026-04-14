/**
 * 课程服务层
 * 提供课程、章节、节的 CRUD 操作
 */
import { request, unwrapListItems } from './api'
import type {
  Course,
  Chapter,
  Section,
  CreateCourseReq,
  UpdateCourseReq,
  CreateChapterReq,
  CreateSectionReq,
  PaginatedResponse,
  CatalogResp,
} from '../types'

/**
 * 获取课程列表
 */
export async function getCourses(
  token: string,
  params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  } = {},
): Promise<PaginatedResponse<Course>> {
  const searchParams = new URLSearchParams()
  if (params.page !== undefined) searchParams.set('page', String(params.page))
  if (params.page_size !== undefined) searchParams.set('page_size', String(params.page_size))
  if (params.keyword) searchParams.set('keyword', params.keyword)
  if (params.status) searchParams.set('status', params.status)
  if (params.sort_by) searchParams.set('sort_by', params.sort_by)
  if (params.sort_order) searchParams.set('sort_order', params.sort_order)

  const query = searchParams.toString()
  return request<PaginatedResponse<Course>>(`/teacher/courses${query ? `?${query}` : ''}`, {
    method: 'GET',
    token,
  })
}

/**
 * 创建课程
 */
export async function createCourse(token: string, data: CreateCourseReq): Promise<Course> {
  return request<Course>('/teacher/courses', {
    method: 'POST',
    token,
    body: data,
  })
}

/**
 * 更新课程
 */
export async function updateCourse(
  token: string,
  courseId: number,
  data: UpdateCourseReq,
): Promise<Course> {
  return request<Course>(`/teacher/courses/${courseId}`, {
    method: 'PUT',
    token,
    body: data,
  })
}

/**
 * 删除课程
 */
export async function deleteCourse(token: string, courseId: number): Promise<void> {
  return request<void>(`/teacher/courses/${courseId}`, {
    method: 'DELETE',
    token,
  })
}

/**
 * 获取课程章节列表
 */
export async function getChapters(token: string, courseId: number): Promise<Chapter[]> {
  const data = await request<Chapter[] | { items: Chapter[] }>(`/teacher/courses/${courseId}/chapters`, {
    method: 'GET',
    token,
  })
  return unwrapListItems(data)
}

/**
 * 创建章节
 */
export async function createChapter(
  token: string,
  courseId: number,
  data: CreateChapterReq,
): Promise<Chapter> {
  return request<Chapter>(`/teacher/courses/${courseId}/chapters`, {
    method: 'POST',
    token,
    body: data,
  })
}

/**
 * 获取章节下的节列表
 */
export async function getSections(token: string, courseId: number, chapterId: number): Promise<Section[]> {
  const data = await request<Section[] | { items: Section[] }>(
    `/teacher/courses/${courseId}/chapters/${chapterId}/sections`,
    {
      method: 'GET',
      token,
    },
  )
  return unwrapListItems(data)
}

/**
 * 创建节
 */
export async function createSection(
  token: string,
  courseId: number,
  chapterId: number,
  data: CreateSectionReq,
): Promise<Section> {
  return request<Section>(`/teacher/courses/${courseId}/chapters/${chapterId}/sections`, {
    method: 'POST',
    token,
    body: data,
  })
}

export async function updateChapter(
  token: string,
  chapterId: number,
  data: CreateChapterReq,
): Promise<Chapter> {
  return request<Chapter>(`/teacher/chapters/${chapterId}`, {
    method: 'PUT',
    token,
    body: data,
  })
}

export async function deleteChapter(token: string, chapterId: number): Promise<void> {
  return request<void>(`/teacher/chapters/${chapterId}`, {
    method: 'DELETE',
    token,
  })
}

export async function updateSection(
  token: string,
  sectionId: number,
  data: CreateSectionReq,
): Promise<Section> {
  return request<Section>(`/teacher/sections/${sectionId}`, {
    method: 'PUT',
    token,
    body: data,
  })
}

export async function deleteSection(token: string, sectionId: number): Promise<void> {
  return request<void>(`/teacher/sections/${sectionId}`, {
    method: 'DELETE',
    token,
  })
}

/**
 * 获取学生课程目录
 * 学生端专用，返回章节→节→资源的层级结构
 */
export async function getCourseCatalog(
  token: string,
  courseId: number,
): Promise<CatalogResp> {
  return request<CatalogResp>(`/student/my/courses/${courseId}/catalog`, {
    method: 'GET',
    token,
  })
}
