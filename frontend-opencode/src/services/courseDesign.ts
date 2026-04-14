/**
 * 课程设计 API：AI 教学大纲草案生成与应用（章/节写入课程末尾）。
 */
import { request } from './api'
import type {
  ApplyOutlineDraftResp,
  GenerateOutlineDraftResp,
  GenerateSectionLessonPlanResp,
  OutlineChapterDraft,
} from '@/types'

export type GenerateOutlineDraftBody = {
  qa_model_id?: number
  extra_hint?: string
}

/** POST /teacher/courses/:courseId/design/outline-draft/generate */
export async function generateOutlineDraft(
  token: string,
  courseId: number,
  body: GenerateOutlineDraftBody = {},
): Promise<GenerateOutlineDraftResp> {
  return request<GenerateOutlineDraftResp>(
    `/teacher/courses/${courseId}/design/outline-draft/generate`,
    {
      method: 'POST',
      token,
      body: Object.keys(body).length ? body : {},
    },
  )
}

/** POST /teacher/courses/:courseId/design/outline-draft/apply */
export async function applyOutlineDraft(
  token: string,
  courseId: number,
  chapters: OutlineChapterDraft[],
): Promise<ApplyOutlineDraftResp> {
  return request<ApplyOutlineDraftResp>(
    `/teacher/courses/${courseId}/design/outline-draft/apply`,
    {
      method: 'POST',
      token,
      body: { chapters },
    },
  )
}


export type GenerateSectionLessonPlanBody = {
  qa_model_id?: number
  objectives: string[]
  teaching_style?: string
  duration_minutes?: number
  extra_hint?: string
}

/** POST /teacher/sections/:sectionId/lesson-plan/generate */
export async function generateSectionLessonPlan(
  token: string,
  sectionId: number,
  body: GenerateSectionLessonPlanBody,
): Promise<GenerateSectionLessonPlanResp> {
  return request<GenerateSectionLessonPlanResp>(
    `/teacher/sections/${sectionId}/lesson-plan/generate`,
    {
      method: 'POST',
      token,
      body,
    },
  )
}
