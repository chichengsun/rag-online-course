import { request } from './api'
import type {
  Resource,
  ResourceProgressResp,
  UpdateProgressReq,
} from '../types'

/**
 * 我的课程项
 */
export interface MyCourseItem {
  id: string
  title: string
  description: string
  enrolled_at: string
}

/**
 * 获取学生已选课程列表
 */
export async function getMyCourses(token: string): Promise<MyCourseItem[]> {
  return request<MyCourseItem[]>('/student/my/courses', {
    method: 'GET',
    token,
  })
}

/**
 * 获取已发布课程列表（供学生选课）
 */
export async function getPublishedCourses(token: string): Promise<MyCourseItem[]> {
  return request<MyCourseItem[]>('/student/courses', {
    method: 'GET',
    token,
  })
}

/**
 * 学生选课
 */
export async function enrollCourse(token: string, courseId: number): Promise<void> {
  return request<void>(`/student/courses/${courseId}/enroll`, {
    method: 'POST',
    token,
  })
}

/**
 * 课程目录章节项
 */
export interface CatalogChapter {
  id: number
  title: string
  sort_order: number
  sections: CatalogSection[]
}

/**
 * 课程目录节项
 */
export interface CatalogSection {
  id: number
  title: string
  sort_order: number
  resources: CatalogResource[]
}

/**
 * 课程目录资源项
 */
export interface CatalogResource {
  id: number
  title: string
  resource_type: string
  sort_order: number
  progress_percent: number
  is_completed: boolean
}

/**
 * 课程目录响应
 */
export interface CatalogResp {
  chapters: CatalogChapter[]
}

/**
 * 获取课程目录
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

/**
 * 获取资源详情（学生端）
 */
export async function getResource(
  token: string,
  resourceId: number,
): Promise<Resource> {
  return request<Resource>(`/student/resources/${resourceId}`, {
    method: 'GET',
    token,
  })
}

/**
 * 获取资源预览 URL（学生端）
 */
export async function getPreviewUrl(
  token: string,
  resourceId: number,
): Promise<string> {
  const resp = await request<{ preview_url: string }>(
    `/student/resources/${resourceId}/preview-url`,
    {
      method: 'GET',
      token,
    },
  )
  return resp.preview_url
}

/**
 * 更新学习进度
 */
export async function updateProgress(
  token: string,
  resourceId: number,
  data: UpdateProgressReq,
): Promise<ResourceProgressResp> {
  return request<ResourceProgressResp>(
    `/student/resources/${resourceId}/progress`,
    {
      method: 'PUT',
      token,
      body: data,
    },
  )
}

/**
 * 标记资源完成
 */
export async function completeResource(
  token: string,
  resourceId: number,
): Promise<ResourceProgressResp> {
  return request<ResourceProgressResp>(
    `/student/resources/${resourceId}/complete`,
    {
      method: 'POST',
      token,
      body: {},
    },
  )
}

/**
 * 获取资源进度
 */
export async function getResourceProgress(
  token: string,
  resourceId: number,
): Promise<ResourceProgressResp> {
  return request<ResourceProgressResp>(
    `/student/resources/${resourceId}/progress`,
    {
      method: 'GET',
      token,
    },
  )
}