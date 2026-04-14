/**
 * API 响应类型定义
 * 与后端 internal/api/response/response.go 中的 Envelope 结构对应
 */

/**
 * API 统一响应包装器
 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

/**
 * 分页响应包装器
 */
export interface PaginatedResponse<T> {
  page: number
  page_size: number
  total: number
  items: T[]
}

/**
 * 错误类型（与 services/api.ts 中的 ApiError 对应）
 * 后端返回非 2xx 时抛出此错误
 */
export class ApiError extends Error {
  code: number
  status: number

  constructor(message: string, code = 500, status = 500) {
    super(message)
    this.code = code
    this.status = status
  }
}

// ============================================================
// 枚举类型
// 与后端 internal/domain/enums.go 对应
// ============================================================

/** 用户角色 */
export type UserRole = 'student' | 'teacher'

/** 课程状态 */
export type CourseStatus = 'draft' | 'published' | 'archived'

/** 资源类型 */
export type ResourceType = 'ppt' | 'pdf' | 'txt' | 'video' | 'doc' | 'docx' | 'audio'

/** 选课状态 */
export type EnrollmentStatus = 'active' | 'dropped' | 'completed'

/** AI 模型类型 */
export type AIModelType = 'qa' | 'embedding' | 'rerank'

// ============================================================
// 实体类型
// 与后端 internal/domain/*.go 对应
// ============================================================

/**
 * 用户实体
 * 对应后端 domain.User
 */
export interface User {
  id: number
  email: string
  username: string
  name: string
  role: UserRole
  is_active: boolean
  created_at: string
  updated_at: string
}

/**
 * 课程实体
 * 对应后端 domain.Course
 */
export interface Course {
  id: number
  teacher_id: number
  title: string
  description: string
  status: CourseStatus
  cover_image_url: string
  created_at: string
  updated_at: string
}

/** 课程设计：大纲中的「节」，与 dto/course OutlineSectionDraft 一致 */
export interface OutlineSectionDraft {
  title: string
}

/** 课程设计：大纲中的「章」，与 dto/course OutlineChapterDraft 一致 */
export interface OutlineChapterDraft {
  title: string
  sections: OutlineSectionDraft[]
}

/** AI 生成大纲草案响应 */
export interface GenerateOutlineDraftResp {
  chapters: OutlineChapterDraft[]
}

/** 将大纲写入课程后的统计 */
export interface ApplyOutlineDraftResp {
  created_chapters: number
  created_sections: number
}

/** 小节教案：步骤 */
export interface LessonPlanStep {
  phase: string
  duration_minutes: number
  activities: string[]
  resource_refs: string[]
}

/** 小节教案草案（结构化） */
export interface LessonPlanDraft {
  title: string
  objectives: string[]
  preparation: string[]
  steps: LessonPlanStep[]
  assessment: string[]
  homework: string[]
}

/** 生成小节教案草案响应 */
export interface GenerateSectionLessonPlanResp {
  course_title: string
  chapter_title: string
  section_title: string
  resource_count: number
  plan: LessonPlanDraft
}

/**
 * 章节实体
 * 对应后端 domain.Chapter
 */
export interface Chapter {
  id: number
  course_id: number
  title: string
  sort_order: number
  created_at: string
  updated_at: string
}

/**
 * 节实体（章节下的子节，用于组织学习资源）
 * 对应后端 domain.Section
 */
export interface Section {
  id: number
  chapter_id: number
  course_id: number
  title: string
  sort_order: number
  created_at: string
  updated_at: string
}

/**
 * 学习资源实体
 * 对应后端 domain.Resource
 */
export interface Resource {
  id: number
  course_id: number
  chapter_id: number
  section_id: number
  title: string
  resource_type: ResourceType
  sort_order: number
  object_key: string
  object_url: string
  mime_type: string
  size_bytes: number
  duration_seconds: number
  created_at: string
  updated_at: string
  /** 异步 AI 摘要状态，与 chapter_resources.ai_summary_status 一致 */
  ai_summary_status?: string
  ai_summary?: string
  ai_summary_error?: string
  ai_summary_updated_at?: string
}

/**
 * 选课记录实体
 * 对应后端 domain.Enrollment
 */
export interface Enrollment {
  id: number
  course_id: number
  student_id: number
  status: EnrollmentStatus
  enrolled_at: string
  updated_at: string
}

/**
 * 知识分块实体
 * 对应后端 domain.KnowledgeChunk（通过 embedding_chunks 表）
 */
export interface KnowledgeChunk {
  /** 后端 embedding_chunks.id，接口常以十进制字符串返回 */
  id: string
  resource_id: number
  course_id: number
  chunk_index: number
  content: string
  char_start: number | null
  char_end: number | null
  metadata: Record<string, unknown>
  is_embedded: boolean
  is_confirmed: boolean
  created_at: string
  updated_at: string
}

/**
 * 聊天会话实体
 * 对应后端 domain.ChatSession（通过 knowledge_chat_sessions 表）
 */
export interface ChatSession {
  id: string
  course_id: string
  teacher_id?: string
  title: string
  created_at: string
  updated_at: string
  message_count?: number
  last_message_at?: string | null
}

/**
 * 聊天消息实体
 * 对应后端 domain.ChatMessage（通过 knowledge_chat_messages 表）
 */
export interface ChatMessage {
  id: string
  session_id: string
  role: 'user' | 'assistant'
  content: string
  references: ReferenceItem[]
  created_at: string
}

/**
 * AI 模型配置实体
 * 对应后端 domain.AIModel（通过 teacher_ai_models 表）
 */
export interface AIModel {
  id: number
  teacher_id: number
  name: string
  model_type: AIModelType
  api_base_url: string
  model_id: string
  has_api_key: boolean
  created_at: string
  updated_at: string
}

/**
 * 资源学习记录
 * 对应后端 domain.ResourceLearningRecord
 */
export interface ResourceLearningRecord {
  id: number
  resource_id: number
  student_id: number
  started_at: string
  completed_at: string
  watched_seconds: number
  progress_percent: number
  is_completed: boolean
  updated_at: string
}

// ============================================================
// 引用项类型（对话回答中引用的知识分块）
// 对应后端 dto/knowledge/chat.go 中的 ReferenceItem
// ============================================================

/**
 * 回答中引用的来源分块
 */
export interface ReferenceItem {
  citation_no: number
  chunk_id: string
  resource_id: string
  resource_title: string
  chunk_index: number
  score: number
  snippet: string
  full_content: string
}

// ============================================================
// DTO 类型（请求/响应数据结构）
// 与后端 internal/dto/**/*.go 对应
// ============================================================

// -------------------- 认证相关 --------------------

/**
 * 登录请求
 * 对应后端 dto/user/LoginReq
 */
export interface LoginReq {
  account: string
  password: string
}

/**
 * 登录响应
 * 对应后端 dto/user/LoginResp
 */
export interface LoginResp {
  access_token: string
  refresh_token: string
  user: User
}

/**
 * 注册请求
 * 对应后端 dto/user/RegisterReq
 */
export interface RegisterReq {
  email: string
  username: string
  name: string
  password: string
  role: UserRole
}

/**
 * 注册响应
 * 对应后端 dto/user/RegisterResp
 */
export interface RegisterResp {
  id: string
}

/**
 * 刷新 Token 请求
 * 对应后端 dto/user/RefreshReq
 */
export interface RefreshReq {
  refresh_token: string
}

/**
 * 刷新 Token 响应
 * 对应后端 dto/user/RefreshResp
 */
export interface RefreshResp {
  access_token: string
}

/**
 * 当前用户资料响应
 * 对应后端 dto/user/MeResp
 */
export type MeResp = User

// -------------------- 课程相关 --------------------

/**
 * 创建课程请求
 * 对应后端 dto/course/CreateCourseReq
 */
export interface CreateCourseReq {
  title: string
  description?: string
}

/**
 * 创建课程响应
 * 对应后端 dto/course/CreateCourseResp
 */
export interface CreateCourseResp {
  id: string
}

/**
 * 更新课程请求
 * 对应后端 dto/course/UpdateCourseReq
 */
export interface UpdateCourseReq {
  title: string
  description?: string
  status: CourseStatus
}

/**
 * 课程列表响应
 * 对应后端 dto/course/ListCoursesResp
 */
export type ListCoursesResp = PaginatedResponse<Course>

// -------------------- 章节相关 --------------------

/**
 * 创建章节请求
 * 对应后端 dto/course/CreateChapterReq
 */
export interface CreateChapterReq {
  title: string
  sort_order: number
}

/**
 * 创建章节响应
 * 对应后端 dto/course/CreateChapterResp
 */
export interface CreateChapterResp {
  id: string
}

/**
 * 章节列表响应
 * 对应后端 dto/course/ListChaptersResp
 */
export type ListChaptersResp = Chapter[]

/**
 * 调整章节顺序请求
 * 对应后端 dto/course/ReorderChapterReq
 */
export interface ReorderChapterReq {
  sort_order: number
}

// -------------------- 节（Section）相关 --------------------

/**
 * 创建节请求
 * 对应后端 dto/course/CreateSectionReq
 */
export interface CreateSectionReq {
  title: string
  sort_order: number
}

/**
 * 创建节响应
 * 对应后端 dto/course/CreateSectionResp
 */
export interface CreateSectionResp {
  id: string
}

/**
 * 节列表响应
 * 对应后端 dto/course/ListSectionsResp
 */
export type ListSectionsResp = Section[]

/**
 * 调整节顺序请求
 * 对应后端 dto/course/ReorderSectionReq
 */
export interface ReorderSectionReq {
  sort_order: number
}

// -------------------- 资源相关 --------------------

/**
 * 初始化资源上传请求
 * 对应后端 dto/course/InitUploadReq
 */
export interface InitUploadReq {
  course_id: number
  file_name: string
  resource_type: ResourceType
}

/**
 * 初始化资源上传响应
 * 对应后端 dto/course/InitUploadResp
 */
export interface InitUploadResp {
  object_key: string
  upload_url: string
  expire_seconds: number
}

/**
 * 确认资源入库请求
 * 对应后端 dto/course/ConfirmResourceReq
 */
export interface ConfirmResourceReq {
  title: string
  resource_type: ResourceType
  sort_order: number
  object_key: string
  mime_type: string
  size_bytes: number
}

/**
 * 确认资源响应
 * 对应后端 dto/course/ConfirmResourceResp
 */
export interface ConfirmResourceResp {
  id: string
}

/**
 * 资源列表响应
 * 对应后端 dto/course/ListResourcesResp
 */
export type ListResourcesResp = Resource[]

/**
 * 预览 URL 响应
 * 对应后端 dto/course/PreviewResourceURLResp
 */
export interface PreviewResourceURLResp {
  preview_url: string
}

/**
 * 解析资源响应
 * 对应后端 dto/course/ParseResourceResp
 */
export interface ParseResourceResp {
  status: string
  markdown?: string
  metadata?: Record<string, unknown>
  error?: string
  image_count: number
}

/**
 * AI 摘要任务响应（异步 accepted 时 status=running）
 * 对应后端 dto/course/SummarizeResourceResp
 */
export interface SummarizeResourceResp {
  status: string
  summary?: string
  updated_at?: string
  error?: string
}

/**
 * 调整资源顺序请求
 * 对应后端 dto/course/ReorderResourceReq
 */
export interface ReorderResourceReq {
  sort_order: number
}

// -------------------- 知识库相关 --------------------

/**
 * 课程知识资源列表响应
 * 对应后端 dto/knowledge/ListKnowledgeResourcesResp
 */
export type ListKnowledgeResourcesResp = PaginatedResponse<Resource>

/**
 * 分块预览请求
 * 对应后端 dto/knowledge/ChunkPreviewReq
 */
export interface ChunkPreviewReq {
  chunk_size: number
  overlap?: number
  clear_persisted_first?: boolean
}

/**
 * 分块预览片段
 * 对应后端 dto/knowledge/ChunkPreviewSegment
 */
export interface ChunkPreviewSegment {
  index: number
  content: string
  char_start: number
  char_end: number
}

/**
 * 分块预览响应
 * 对应后端 dto/knowledge/ChunkPreviewResp
 */
export interface ChunkPreviewResp {
  segments: ChunkPreviewSegment[]
}

/**
 * 待保存的分块
 * 对应后端 dto/knowledge/ChunkSaveItem
 */
export interface ChunkSaveItem {
  content: string
  char_start?: number | null
  char_end?: number | null
  metadata?: Record<string, unknown>
}

/**
 * 保存分块请求
 * 对应后端 dto/knowledge/SaveKnowledgeChunksReq
 */
export interface SaveKnowledgeChunksReq {
  chunks: ChunkSaveItem[]
}

/**
 * 嵌入资源请求
 * 对应后端 dto/knowledge/EmbedResourceReq
 */
export interface EmbedResourceReq {
  model_id: number
}

/**
 * 嵌入资源响应
 * 对应后端 dto/knowledge/EmbedResourceResp
 */
export interface EmbedResourceResp {
  embedded_count: number
}

/**
 * 确认分块响应
 * 对应后端 dto/knowledge/ConfirmKnowledgeChunksResp
 */
export interface ConfirmKnowledgeChunksResp {
  confirmed_count: number
}

/**
 * 分块列表响应
 * 对应后端 dto/knowledge/ListKnowledgeChunksResp（items 为行 map）
 */
export interface ListKnowledgeChunksResp {
  items: Record<string, unknown>[]
}

/**
 * 更新知识分块请求
 * 对应后端 dto/knowledge/UpdateKnowledgeChunkReq
 */
export interface UpdateKnowledgeChunkReq {
  content: string
  char_start?: number | null
  char_end?: number | null
}

// -------------------- 对话相关 --------------------

/**
 * 创建聊天会话请求
 * 对应后端 dto/knowledge/CreateChatSessionReq
 */
export interface CreateChatSessionReq {
  title?: string
}

/**
 * 创建聊天会话响应
 * 对应后端 dto/knowledge/CreateChatSessionResp
 */
export interface CreateChatSessionResp {
  id: string
}

/**
 * 更新聊天会话请求
 * 对应后端 dto/knowledge/UpdateChatSessionReq
 */
export interface UpdateChatSessionReq {
  title: string
}

/**
 * 聊天会话列表响应
 * 对应后端 dto/knowledge/ListChatSessionsResp
 */
export type ListChatSessionsResp = PaginatedResponse<ChatSession>

/**
 * 聊天消息列表响应
 * 对应后端 dto/knowledge/ListChatMessagesResp
 */
export type ListChatMessagesResp = PaginatedResponse<ChatMessage>

/**
 * 会话内提问请求
 * 对应后端 dto/knowledge/AskInSessionReq
 */
export interface AskInSessionReq {
  question: string
  top_k?: number
  use_rerank?: boolean
  qa_model_id?: number
  semantic_min_score?: number
  keyword_min_score?: number
}

/**
 * 会话内提问响应（非流式）
 * 对应后端 dto/knowledge/AskInSessionResp
 */
export interface AskInSessionResp {
  session_id: string
  user_message_id: string
  assistant_message_id: string
  answer: string
  references: ReferenceItem[]
}

// -------------------- AI 模型相关 --------------------

/**
 * AI 模型列表项（不含 api_key）
 * 对应后端 dto/knowledge/AIModelListItem
 */
export interface AIModelListItem {
  id: string
  name: string
  model_type: AIModelType
  api_base_url: string
  model_id: string
  has_api_key: boolean
}

/**
 * AI 模型列表响应
 */
export type ListAIModelsResp = AIModelListItem[]

/**
 * 创建 AI 模型请求
 * 对应后端 dto/knowledge/CreateAIModelReq
 */
export interface CreateAIModelReq {
  name: string
  model_type: AIModelType
  api_base_url: string
  model_id: string
  api_key?: string
}

/**
 * 创建 AI 模型响应
 * 对应后端 dto/knowledge/CreateAIModelResp
 */
export interface CreateAIModelResp {
  id: string
}

/**
 * 更新 AI 模型请求
 * 对应后端 dto/knowledge/UpdateAIModelReq
 */
export interface UpdateAIModelReq {
  name: string
  api_base_url: string
  model_id: string
  api_key?: string
}

/**
 * 测试 AI 模型连接请求
 * 对应后端 dto/knowledge/TestAIModelConnectionReq
 */
export interface TestAIModelConnectionReq {
  model_type: AIModelType
  api_base_url: string
  model_id: string
  api_key?: string
  existing_model_id?: number
}

/**
 * 测试 AI 模型连接响应
 * 对应后端 dto/knowledge/TestAIModelConnectionResp
 */
export interface TestAIModelConnectionResp {
  ok: boolean
  message: string
  http_status?: number
}

// -------------------- 选课相关 --------------------

/**
 * 学生选课请求
 */
export interface EnrollReq {
  course_id: number
}

/**
 * 学生选课响应
 */
export interface EnrollResp {
  id: string
}

// -------------------- 学习进度相关 --------------------

/**
 * 更新资源进度请求
 */
export interface UpdateProgressReq {
  resource_id: number
  watched_seconds: number
  progress_percent: number
  is_completed: boolean
}

/**
 * 资源进度响应
 */
export interface ResourceProgressResp {
  resource_id: number
  progress_percent: number
  is_completed: boolean
  updated_at: string
}

// -------------------- 学生端课程目录相关 --------------------

/**
 * 课程目录中的资源
 * 对应后端 catalog_repository.go 中的资源行
 */
export interface CatalogResource {
  id: string
  course_id: string
  chapter_id: string
  section_id: string
  title: string
  resource_type: string
  sort_order: number
  object_url: string
  mime_type: string
  size_bytes: number
  ai_summary: string | null
  ai_summary_updated_at: string | null
}

/**
 * 课程目录中的节
 * 对应后端 catalog_repository.go 中的节行
 */
export interface CatalogSection {
  id: string
  title: string
  sort_order: number
  resources: CatalogResource[]
}

/**
 * 课程目录中的章节
 * 对应后端 catalog_repository.go 中的章节行
 */
export interface CatalogChapter {
  id: string
  title: string
  sort_order: number
  sections: CatalogSection[]
}

/**
 * 课程目录响应
 * 对应后端 dto/learning/learning.go 中的 CatalogResp
 */
export interface CatalogResp {
  course_id: string
  chapters: CatalogChapter[]
}

// ============================================================
// 工具类型
// ============================================================

/**
 * 从 API 响应中提取 data 字段的类型
 */
export type UnwrapApiResponse<T extends ApiResponse<unknown>> = T extends ApiResponse<infer U> ? U : never

/**
 * 从分页响应中提取 items 的类型
 */
export type UnwrapPaginatedResponse<T> = T extends PaginatedResponse<infer U> ? U : never
