import { request } from './api'

/** 课程设计：大纲中的小节草案。 */
export type OutlineSectionDraft = {
  title: string
}

/** 课程设计：大纲中的章节草案。 */
export type OutlineChapterDraft = {
  title: string
  sections: OutlineSectionDraft[]
}

export type GenerateOutlineDraftResp = {
  chapters: OutlineChapterDraft[]
}

export type ApplyOutlineDraftResp = {
  created_chapters: number
  created_sections: number
}

export type GenerateOutlineDraftBody = {
  qa_model_id?: number
  extra_hint?: string
}

/** 生成 AI 课程大纲草案。 */
export async function generateOutlineDraft(token: string, courseId: number, body: GenerateOutlineDraftBody = {}) {
  return request<GenerateOutlineDraftResp>(`/teacher/courses/${courseId}/design/outline-draft/generate`, {
    method: 'POST',
    token,
    body: Object.keys(body).length ? body : {},
  })
}

/** 将编辑后的草案章/节追加写入课程。 */
export async function applyOutlineDraft(token: string, courseId: number, chapters: OutlineChapterDraft[]) {
  return request<ApplyOutlineDraftResp>(`/teacher/courses/${courseId}/design/outline-draft/apply`, {
    method: 'POST',
    token,
    body: { chapters },
  })
}
