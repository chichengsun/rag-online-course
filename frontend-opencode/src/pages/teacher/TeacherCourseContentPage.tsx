import { useEffect, useState, useCallback, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import {
  getChapters,
  createChapter,
  getSections,
  createSection,
  updateChapter,
  deleteChapter,
  updateSection,
  deleteSection,
} from '@/services/course'
import {
  getSectionResources,
  deleteResource,
  summarizeResource,
  initUpload,
  confirmResource,
} from '@/services/resource'
import { generateSectionLessonPlan } from '@/services/courseDesign'
import type { Chapter, Section, Resource, ResourceType, LessonPlanDraft } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button, buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from '@/components/ui/collapsible'
import {
  ChevronDown,
  ChevronUp,
  Plus,
  Trash2,
  Edit,
  ArrowUp,
  ArrowDown,
  Sparkles,
  Eye,
  Upload,
  ArrowLeft,
  BookOpen,
  Loader2,
} from 'lucide-react'
import { toast } from 'sonner'
import MarkdownRenderer from '@/components/MarkdownRenderer'
import { cn } from '@/lib/utils'

const resourceTypeLabels: Record<ResourceType, string> = {
  pdf: 'PDF',
  doc: 'DOC',
  docx: 'DOCX',
  ppt: 'PPT',
  txt: 'TXT',
  video: '视频',
  audio: '音频',
}

function inferResourceType(fileName: string): ResourceType {
  const ext = fileName.toLowerCase().split('.').pop() ?? ''
  if (ext === 'pdf') return 'pdf'
  if (ext === 'txt' || ext === 'md' || ext === 'markdown') return 'txt'
  if (ext === 'doc') return 'doc'
  if (ext === 'docx') return 'docx'
  if (ext === 'ppt' || ext === 'pptx') return 'ppt'
  if (['mp3', 'wav', 'ogg', 'oga', 'm4a', 'aac', 'flac', 'wma'].includes(ext)) return 'audio'
  if (['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) return 'video'
  return 'pdf'
}

function inferMimeType(file: File): string {
  if (file.type) return file.type
  const ext = file.name.toLowerCase().split('.').pop() ?? ''
  if (ext === 'pdf') return 'application/pdf'
  if (ext === 'txt') return 'text/plain; charset=utf-8'
  if (ext === 'md' || ext === 'markdown') return 'text/markdown; charset=utf-8'
  if (ext === 'ppt') return 'application/vnd.ms-powerpoint'
  if (ext === 'pptx') {
    return 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
  }
  if (ext === 'doc') return 'application/msword'
  if (ext === 'docx') {
    return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
  }
  if (ext === 'mp3') return 'audio/mpeg'
  if (ext === 'wav') return 'audio/wav'
  if (ext === 'ogg' || ext === 'oga') return 'audio/ogg'
  if (ext === 'm4a') return 'audio/mp4'
  if (ext === 'aac') return 'audio/aac'
  if (ext === 'flac') return 'audio/flac'
  if (ext === 'wma') return 'audio/x-ms-wma'
  if (['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) return 'video/mp4'
  return 'application/octet-stream'
}

/** 支持异步 AI 摘要的文档类 resource_type（与后端 isDocumentResourceType 一致）。 */
const SUMMARY_DOC_TYPES: ResourceType[] = ['pdf', 'txt', 'doc', 'docx', 'ppt']

function coerceNum(v: unknown): number {
  if (typeof v === 'number' && !Number.isNaN(v)) return v
  if (typeof v === 'string') {
    const n = parseInt(v, 10)
    return Number.isNaN(n) ? 0 : n
  }
  return 0
}

/** 将列表接口返回的资源行规范为 Resource（含摘要任务字段）。 */
function normalizeResourceRow(r: Record<string, unknown>): Resource {
  return {
    id: coerceNum(r.id),
    course_id: coerceNum(r.course_id),
    chapter_id: coerceNum(r.chapter_id),
    section_id: coerceNum(r.section_id),
    title: String(r.title ?? ''),
    resource_type: r.resource_type as ResourceType,
    sort_order: coerceNum(r.sort_order),
    object_key: String(r.object_key ?? ''),
    object_url: String(r.object_url ?? ''),
    mime_type: String(r.mime_type ?? ''),
    size_bytes: coerceNum(r.size_bytes),
    duration_seconds: coerceNum(r.duration_seconds),
    created_at: String(r.created_at ?? ''),
    updated_at: String(r.updated_at ?? ''),
    ai_summary_status: typeof r.ai_summary_status === 'string' ? r.ai_summary_status : 'idle',
    ai_summary: typeof r.ai_summary === 'string' ? r.ai_summary : undefined,
    ai_summary_error: typeof r.ai_summary_error === 'string' ? r.ai_summary_error : undefined,
    ai_summary_updated_at:
      typeof r.ai_summary_updated_at === 'string' ? r.ai_summary_updated_at : undefined,
  }
}

export function TeacherCourseContentPage() {
  const { token } = useAuthStore()
  const { courseId } = useParams<{ courseId: string }>()
  const navigate = useNavigate()
  const accessToken = token ?? ''

  const [chapters, setChapters] = useState<Chapter[]>([])
  const [sectionsByChapter, setSectionsByChapter] = useState<Record<number, Section[]>>({})
  const [resourcesBySection, setResourcesBySection] = useState<Record<number, Resource[]>>({})
  const [expandedChapters, setExpandedChapters] = useState<Set<number>>(new Set())
  const [expandedSections, setExpandedSections] = useState<Set<number>>(new Set())
  const [loading, setLoading] = useState(false)
  /** 展开显示 AI 摘要（Markdown）的资源 id */
  const [expandedSummaryResourceId, setExpandedSummaryResourceId] = useState<number | null>(null)

  const [chapterDialogOpen, setChapterDialogOpen] = useState(false)
  const [chapterDialogMode, setChapterDialogMode] = useState<'create' | 'edit'>('create')
  const [editingChapter, setEditingChapter] = useState<Chapter | null>(null)
  const [chapterForm, setChapterForm] = useState({ title: '', sort_order: 1 })

  const [sectionDialogOpen, setSectionDialogOpen] = useState(false)
  const [sectionDialogMode, setSectionDialogMode] = useState<'create' | 'edit'>('create')
  const [editingSection, setEditingSection] = useState<Section | null>(null)
  const [currentChapterId, setCurrentChapterId] = useState<number | null>(null)
  const [sectionForm, setSectionForm] = useState({ title: '', sort_order: 1 })

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<{
    type: 'chapter' | 'section' | 'resource'
    item: Chapter | Section | Resource
  } | null>(null)

  const [uploadDialogOpen, setUploadDialogOpen] = useState(false)
  const [uploadSectionId, setUploadSectionId] = useState<number | null>(null)
  const [uploadFile, setUploadFile] = useState<File | null>(null)
  const [uploadForm, setUploadForm] = useState({
    title: '',
    sort_order: 1,
  })
  const [lessonPlanDialogOpen, setLessonPlanDialogOpen] = useState(false)
  const [lessonPlanSection, setLessonPlanSection] = useState<Section | null>(null)
  const [lessonPlanGenerating, setLessonPlanGenerating] = useState(false)
  const [lessonPlanForm, setLessonPlanForm] = useState({
    objectivesText: '',
    teachingStyle: '',
    durationMinutes: 45,
    extraHint: '',
  })
  const [lessonPlanDraft, setLessonPlanDraft] = useState<LessonPlanDraft | null>(null)
  const loadCatalog = useCallback(async (opts?: { silent?: boolean }) => {
    if (!courseId) return
    const courseIdNum = parseInt(courseId, 10)
    if (isNaN(courseIdNum)) return

    const silent = opts?.silent === true
    if (!silent) setLoading(true)
    try {
      const chaptersData = await getChapters(accessToken, courseIdNum)
      setChapters(chaptersData)

      const sectionsMap: Record<number, Section[]> = {}
      const resourcesMap: Record<number, Resource[]> = {}

      await Promise.all(
        chaptersData.map(async (chapter) => {
          const sectionsData = await getSections(accessToken, courseIdNum, chapter.id)
          sectionsMap[chapter.id] = sectionsData

          await Promise.all(
            sectionsData.map(async (section) => {
              const resourcesData = await getSectionResources(accessToken, section.id)
              resourcesMap[section.id] = resourcesData.map((raw) =>
                normalizeResourceRow(raw as unknown as Record<string, unknown>),
              )
            })
          )
        })
      )

      setSectionsByChapter(sectionsMap)
      setResourcesBySection(resourcesMap)

      if (chaptersData.length > 0 && expandedChapters.size === 0) {
        setExpandedChapters(new Set([chaptersData[0].id]))
      }
    } catch (err) {
      if (!silent) {
        toast.error(err instanceof Error ? err.message : '加载课程目录失败')
      }
    } finally {
      if (!silent) setLoading(false)
    }
  }, [accessToken, courseId, expandedChapters.size])

  useEffect(() => {
    void loadCatalog()
  }, [loadCatalog])

  const hasRunningSummaryJob = useMemo(() => {
    for (const list of Object.values(resourcesBySection)) {
      for (const r of list) {
        if (r.ai_summary_status === 'running') return true
      }
    }
    return false
  }, [resourcesBySection])

  useEffect(() => {
    if (!hasRunningSummaryJob) return
    const timer = setInterval(() => {
      void loadCatalog({ silent: true })
    }, 2500)
    return () => clearInterval(timer)
  }, [hasRunningSummaryJob, loadCatalog])

  const getNextChapterSortOrder = () => {
    if (chapters.length === 0) return 1
    return Math.max(...chapters.map((c) => c.sort_order), 0) + 1
  }

  const getNextSectionSortOrder = (chapterId: number) => {
    const sections = sectionsByChapter[chapterId] || []
    if (sections.length === 0) return 1
    return Math.max(...sections.map((s) => s.sort_order), 0) + 1
  }

  const getNextResourceSortOrder = (sectionId: number) => {
    const resources = resourcesBySection[sectionId] || []
    if (resources.length === 0) return 1
    return Math.max(...resources.map((r) => r.sort_order), 0) + 1
  }

  const handleOpenCreateChapterDialog = () => {
    setChapterDialogMode('create')
    setEditingChapter(null)
    setChapterForm({ title: '', sort_order: getNextChapterSortOrder() })
    setChapterDialogOpen(true)
  }

  const handleOpenEditChapterDialog = (chapter: Chapter) => {
    setChapterDialogMode('edit')
    setEditingChapter(chapter)
    setChapterForm({ title: chapter.title, sort_order: chapter.sort_order })
    setChapterDialogOpen(true)
  }

  const handleSubmitChapter = async () => {
    if (!courseId || !chapterForm.title.trim()) return
    const courseIdNum = parseInt(courseId, 10)
    if (isNaN(courseIdNum)) return

    setLoading(true)
    try {
      if (chapterDialogMode === 'create') {
        await createChapter(accessToken, courseIdNum, {
          title: chapterForm.title,
          sort_order: chapterForm.sort_order,
        })
        toast.success('章节创建成功')
      } else if (editingChapter) {
        await updateChapter(accessToken, editingChapter.id, {
          title: chapterForm.title,
          sort_order: chapterForm.sort_order,
        })
        toast.success('章节更新成功')
      }
      setChapterDialogOpen(false)
      await loadCatalog()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '操作失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpenCreateSectionDialog = (chapterId: number) => {
    setSectionDialogMode('create')
    setEditingSection(null)
    setCurrentChapterId(chapterId)
    setSectionForm({ title: '', sort_order: getNextSectionSortOrder(chapterId) })
    setSectionDialogOpen(true)
  }

  const handleOpenEditSectionDialog = (section: Section) => {
    setSectionDialogMode('edit')
    setEditingSection(section)
    setCurrentChapterId(section.chapter_id)
    setSectionForm({ title: section.title, sort_order: section.sort_order })
    setSectionDialogOpen(true)
  }

  const handleSubmitSection = async () => {
    if (!courseId || !sectionForm.title.trim() || !currentChapterId) return
    const courseIdNum = parseInt(courseId, 10)
    if (isNaN(courseIdNum)) return

    setLoading(true)
    try {
      if (sectionDialogMode === 'create') {
        await createSection(accessToken, courseIdNum, currentChapterId, {
          title: sectionForm.title,
          sort_order: sectionForm.sort_order,
        })
        toast.success('节创建成功')
      } else if (editingSection) {
        await updateSection(accessToken, editingSection.id, {
          title: sectionForm.title,
          sort_order: sectionForm.sort_order,
        })
        toast.success('节更新成功')
      }
      setSectionDialogOpen(false)
      await loadCatalog()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '操作失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpenDeleteDialog = (
    type: 'chapter' | 'section' | 'resource',
    item: Chapter | Section | Resource
  ) => {
    setDeleteTarget({ type, item })
    setDeleteDialogOpen(true)
  }

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return

    setLoading(true)
    try {
      if (deleteTarget.type === 'chapter') {
        await deleteChapter(accessToken, deleteTarget.item.id)
        toast.success('章节已删除')
      } else if (deleteTarget.type === 'section') {
        await deleteSection(accessToken, deleteTarget.item.id)
        toast.success('节已删除')
      } else if (deleteTarget.type === 'resource') {
        await deleteResource(accessToken, deleteTarget.item.id)
        toast.success('资源已删除')
      }
      setDeleteDialogOpen(false)
      setDeleteTarget(null)
      await loadCatalog()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除失败')
    } finally {
      setLoading(false)
    }
  }

  const handleMoveChapter = async (chapter: Chapter, direction: 'up' | 'down') => {
    const currentIndex = chapters.findIndex((c) => c.id === chapter.id)
    if (currentIndex === -1) return

    const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1
    if (targetIndex < 0 || targetIndex >= chapters.length) return

    const targetChapter = chapters[targetIndex]
    const newSortOrder = targetChapter.sort_order

    setLoading(true)
    try {
      await updateChapter(accessToken, chapter.id, {
        title: chapter.title,
        sort_order: newSortOrder,
      })
      await loadCatalog()
      toast.success('排序已更新')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '排序失败')
    } finally {
      setLoading(false)
    }
  }

  const handleMoveSection = async (section: Section, direction: 'up' | 'down') => {
    const sections = sectionsByChapter[section.chapter_id] || []
    const currentIndex = sections.findIndex((s) => s.id === section.id)
    if (currentIndex === -1) return

    const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1
    if (targetIndex < 0 || targetIndex >= sections.length) return

    const targetSection = sections[targetIndex]
    const newSortOrder = targetSection.sort_order

    setLoading(true)
    try {
      await updateSection(accessToken, section.id, {
        title: section.title,
        sort_order: newSortOrder,
      })
      await loadCatalog()
      toast.success('排序已更新')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '排序失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpenUploadDialog = (sectionId: number) => {
    setUploadSectionId(sectionId)
    setUploadFile(null)
    setUploadForm({ title: '', sort_order: getNextResourceSortOrder(sectionId) })
    setUploadDialogOpen(true)
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] ?? null
    setUploadFile(file)
    if (file) {
      const title = file.name.replace(/\.[^.]+$/, '')
      setUploadForm((prev) => ({ ...prev, title }))
    }
  }

  const handleUploadResource = async () => {
    if (!courseId || !uploadSectionId || !uploadFile || !uploadForm.title.trim()) return
    const courseIdNum = parseInt(courseId, 10)
    if (isNaN(courseIdNum)) return

    setLoading(true)
    try {
      const initData = await initUpload(accessToken, uploadSectionId, {
        course_id: courseIdNum,
        file_name: uploadFile.name,
        resource_type: inferResourceType(uploadFile.name),
      })

      const putResp = await fetch(initData.upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': inferMimeType(uploadFile) },
        body: uploadFile,
      })

      if (!putResp.ok) {
        throw new Error(`文件上传失败（${putResp.status}）`)
      }

      await confirmResource(accessToken, uploadSectionId, {
        title: uploadForm.title,
        resource_type: inferResourceType(uploadFile.name),
        sort_order: uploadForm.sort_order,
        object_key: initData.object_key,
        mime_type: inferMimeType(uploadFile),
        size_bytes: uploadFile.size,
      })

      toast.success('资源上传成功')
      setUploadDialogOpen(false)
      setUploadFile(null)
      await loadCatalog()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '上传失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpenLessonPlanDialog = (section: Section) => {
    setLessonPlanSection(section)
    setLessonPlanDialogOpen(true)
    setLessonPlanDraft(null)
    setLessonPlanForm({
      objectivesText: '',
      teachingStyle: '',
      durationMinutes: 45,
      extraHint: '',
    })
  }

  const handleGenerateLessonPlan = async () => {
    if (!lessonPlanSection) return
    const objectives = lessonPlanForm.objectivesText
      .split('\n')
      .map((x) => x.trim())
      .filter(Boolean)
    if (objectives.length === 0) {
      toast.error('请至少填写 1 条教学目标')
      return
    }
    setLessonPlanGenerating(true)
    try {
      const out = await generateSectionLessonPlan(accessToken, lessonPlanSection.id, {
        objectives,
        teaching_style: lessonPlanForm.teachingStyle.trim() || undefined,
        duration_minutes: lessonPlanForm.durationMinutes > 0 ? lessonPlanForm.durationMinutes : undefined,
        extra_hint: lessonPlanForm.extraHint.trim() || undefined,
      })
      setLessonPlanDraft(out.plan)
      toast.success('教案草案已生成，可继续调整后复制使用')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '教案生成失败')
    } finally {
      setLessonPlanGenerating(false)
    }
  }

  /** 在内置预览页查看资源（PDF/音视频/Office 转 PDF/文本等），避免直接打开预签名链接触发下载。 */
  const handlePreviewResource = (resource: Resource) => {
    navigate(`/teacher/resources/preview?resourceId=${resource.id}`)
  }

  /** 启动异步摘要任务；成功后刷新目录。 */
  const startSummaryJob = async (resource: Resource) => {
    try {
      const result = await summarizeResource(accessToken, resource.id)
      if (result.status === 'running') {
        toast.success('AI 摘要任务已启动，请稍候…')
      } else if (result.status === 'succeeded' && result.summary) {
        toast.success('摘要已就绪')
        setExpandedSummaryResourceId(resource.id)
      } else {
        toast.info(result.status === 'failed' ? '上次摘要失败' : '摘要状态已更新')
      }
      await loadCatalog()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '摘要任务启动失败')
    }
  }

  /**
   * 摘要按钮：idle/failed 时启动任务；running 时无操作；succeeded 时展开/收起 Markdown 摘要。
   */
  const handleSummaryButton = (resource: Resource) => {
    if (!SUMMARY_DOC_TYPES.includes(resource.resource_type)) {
      toast.error('仅 PDF、Word、PPT、TXT 等文档支持 AI 摘要')
      return
    }
    const st = resource.ai_summary_status ?? 'idle'
    if (st === 'running') return
    if (st === 'succeeded' && (resource.ai_summary ?? '').trim() !== '') {
      setExpandedSummaryResourceId((prev) => (prev === resource.id ? null : resource.id))
      return
    }
    void startSummaryJob(resource)
  }

  const handleRegenerateSummary = async (resource: Resource) => {
    setExpandedSummaryResourceId(resource.id)
    await startSummaryJob(resource)
  }

  const toggleChapter = (chapterId: number) => {
    setExpandedChapters((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(chapterId)) {
        newSet.delete(chapterId)
      } else {
        newSet.add(chapterId)
      }
      return newSet
    })
  }

  const toggleSection = (sectionId: number) => {
    setExpandedSections((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(sectionId)) {
        newSet.delete(sectionId)
      } else {
        newSet.add(sectionId)
      }
      return newSet
    })
  }

  if (!courseId) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>未选择课程</CardTitle>
            <CardDescription>请先从课程列表进入章节与资源管理。</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate('/teacher/courses')}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              返回课程列表
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/teacher/courses')}
            >
              <ArrowLeft className="w-4 h-4 mr-1" />
              返回
            </Button>
          </div>
          <h1 className="text-3xl font-bold text-foreground">章节与资源管理</h1>
          <p className="text-muted-foreground mt-2">
            结构：章节 → 节 → 资源。管理课程的完整内容结构。
          </p>
        </div>
        <Button onClick={handleOpenCreateChapterDialog} disabled={loading}>
          <Plus className="w-4 h-4 mr-2" />
          新增章节
        </Button>
      </div>

      <div className="space-y-4">
        {chapters.length === 0 && !loading && (
          <Card>
            <CardContent className="py-12">
              <div className="text-center">
                <BookOpen className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
                <p className="text-muted-foreground mb-4">暂无章节</p>
                <Button onClick={handleOpenCreateChapterDialog}>
                  <Plus className="w-4 h-4 mr-2" />
                  创建第一个章节
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {chapters.map((chapter, chapterIndex) => {
          const sections = sectionsByChapter[chapter.id] || []
          const isExpanded = expandedChapters.has(chapter.id)

          return (
            <Collapsible
              key={chapter.id}
              open={isExpanded}
              onOpenChange={() => toggleChapter(chapter.id)}
            >
              <Card>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <CollapsibleTrigger
                      className={cn(
                        buttonVariants({ variant: 'ghost', size: 'default' }),
                        'h-auto flex-1 justify-start gap-2 p-0 text-left font-normal',
                      )}
                    >
                      {isExpanded ? (
                        <ChevronDown className="w-5 h-5 shrink-0 text-muted-foreground" />
                      ) : (
                        <ChevronUp className="w-5 h-5 shrink-0 text-muted-foreground" />
                      )}
                      <span className="text-lg font-semibold">
                        第 {chapter.sort_order} 章：{chapter.title}
                      </span>
                      <span className="text-sm text-muted-foreground">({sections.length} 个节)</span>
                    </CollapsibleTrigger>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        disabled={chapterIndex === 0 || loading}
                        onClick={() => handleMoveChapter(chapter, 'up')}
                      >
                        <ArrowUp className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        disabled={chapterIndex === chapters.length - 1 || loading}
                        onClick={() => handleMoveChapter(chapter, 'down')}
                      >
                        <ArrowDown className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenCreateSectionDialog(chapter.id)}
                        disabled={loading}
                      >
                        <Plus className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenEditChapterDialog(chapter)}
                        disabled={loading}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenDeleteDialog('chapter', chapter)}
                        disabled={loading}
                      >
                        <Trash2 className="w-4 h-4 text-destructive" />
                      </Button>
                    </div>
                  </div>
                </CardHeader>

                <CollapsibleContent>
                  <CardContent className="pt-0">
                    <div className="pl-8 space-y-3">
                      {sections.length === 0 && (
                        <p className="text-muted-foreground text-sm py-4">
                          该章节下暂无节
                        </p>
                      )}

                      {sections.map((section, sectionIndex) => {
                        const resources = resourcesBySection[section.id] || []
                        const isSectionExpanded = expandedSections.has(section.id)

                        return (
                          <Collapsible
                            key={section.id}
                            open={isSectionExpanded}
                            onOpenChange={() => toggleSection(section.id)}
                          >
                            <div className="border rounded-lg p-4">
                              <div className="flex items-center justify-between">
                                <CollapsibleTrigger
                                  className={cn(
                                    buttonVariants({ variant: 'ghost', size: 'default' }),
                                    'h-auto flex-1 justify-start gap-2 p-0 text-left font-normal',
                                  )}
                                >
                                  {isSectionExpanded ? (
                                    <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground" />
                                  ) : (
                                    <ChevronUp className="h-4 w-4 shrink-0 text-muted-foreground" />
                                  )}
                                  <span className="font-medium">
                                    {section.sort_order}. {section.title}
                                  </span>
                                  <span className="text-sm text-muted-foreground">
                                    ({resources.length} 个资源)
                                  </span>
                                </CollapsibleTrigger>
                                <div className="flex items-center gap-1">
                                  <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    disabled={sectionIndex === 0 || loading}
                                    onClick={() => handleMoveSection(section, 'up')}
                                  >
                                    <ArrowUp className="w-3 h-3" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    disabled={sectionIndex === sections.length - 1 || loading}
                                    onClick={() => handleMoveSection(section, 'down')}
                                  >
                                    <ArrowDown className="w-3 h-3" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    onClick={() => handleOpenUploadDialog(section.id)}
                                    disabled={loading}
                                    title="上传资源"
                                  >
                                    <Upload className="w-3 h-3" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    onClick={() => handleOpenLessonPlanDialog(section)}
                                    disabled={loading || lessonPlanGenerating}
                                    title="生成教案"
                                  >
                                    <Sparkles className="w-3 h-3" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    onClick={() => handleOpenEditSectionDialog(section)}
                                    disabled={loading}
                                  >
                                    <Edit className="w-3 h-3" />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    onClick={() => handleOpenDeleteDialog('section', section)}
                                    disabled={loading}
                                  >
                                    <Trash2 className="w-3 h-3 text-destructive" />
                                  </Button>
                                </div>
                              </div>

                              <CollapsibleContent>
                                <div className="mt-3 pl-6 space-y-2">
                                  {resources.length === 0 && (
                                    <p className="text-muted-foreground text-sm py-2">
                                      该节下暂无资源
                                    </p>
                                  )}

                                  {resources.map((resource) => {
                                    const sumStatus = resource.ai_summary_status ?? 'idle'
                                    const summaryBusy = sumStatus === 'running'
                                    const canSummary = SUMMARY_DOC_TYPES.includes(resource.resource_type)
                                    return (
                                      <div key={resource.id} className="space-y-2">
                                        <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-md">
                                          <div className="flex items-center gap-3 min-w-0">
                                            <span className="text-xs font-medium px-2 py-0.5 bg-primary/10 text-primary rounded shrink-0">
                                              {resourceTypeLabels[resource.resource_type]}
                                            </span>
                                            <span className="text-sm truncate">{resource.title}</span>
                                            <span className="text-xs text-muted-foreground shrink-0">
                                              ({(resource.size_bytes / 1024 / 1024).toFixed(2)} MB)
                                            </span>
                                            {canSummary && summaryBusy && (
                                              <span className="text-xs text-muted-foreground shrink-0">
                                                摘要生成中…
                                              </span>
                                            )}
                                            {canSummary && sumStatus === 'failed' && (
                                              <span className="text-xs text-destructive shrink-0">摘要失败</span>
                                            )}
                                          </div>
                                          <div className="flex items-center gap-1 shrink-0">
                                            <Button
                                              variant="ghost"
                                              size="icon-sm"
                                              type="button"
                                              onClick={() => handlePreviewResource(resource)}
                                              disabled={loading}
                                              title="查看"
                                            >
                                              <Eye className="w-3 h-3" />
                                            </Button>
                                            {canSummary && (
                                              <Button
                                                variant="ghost"
                                                size="icon-sm"
                                                type="button"
                                                onClick={() => handleSummaryButton(resource)}
                                                disabled={loading || summaryBusy}
                                                title={
                                                  sumStatus === 'succeeded' && resource.ai_summary
                                                    ? expandedSummaryResourceId === resource.id
                                                      ? '收起摘要'
                                                      : '查看 AI 摘要'
                                                    : '生成 AI 摘要'
                                                }
                                              >
                                                {summaryBusy ? (
                                                  <Loader2 className="w-3 h-3 animate-spin" />
                                                ) : (
                                                  <Sparkles className="w-3 h-3" />
                                                )}
                                              </Button>
                                            )}
                                            <Button
                                              variant="ghost"
                                              size="icon-sm"
                                              type="button"
                                              onClick={() => handleOpenDeleteDialog('resource', resource)}
                                              disabled={loading}
                                              title="删除"
                                            >
                                              <Trash2 className="w-3 h-3 text-destructive" />
                                            </Button>
                                          </div>
                                        </div>
                                        {canSummary && sumStatus === 'failed' && resource.ai_summary_error && (
                                          <p className="text-xs text-destructive px-3">
                                            {resource.ai_summary_error}
                                          </p>
                                        )}
                                        {canSummary &&
                                          expandedSummaryResourceId === resource.id &&
                                          resource.ai_summary &&
                                          sumStatus === 'succeeded' && (
                                            <div className="rounded-md border border-border bg-card px-3 py-3 text-sm">
                                              <div className="mb-2 flex items-center justify-between gap-2">
                                                <span className="font-medium text-foreground">AI 摘要</span>
                                                <Button
                                                  variant="ghost"
                                                  size="sm"
                                                  className="h-7 text-xs"
                                                  type="button"
                                                  disabled={loading || summaryBusy}
                                                  onClick={() => void handleRegenerateSummary(resource)}
                                                >
                                                  重新生成
                                                </Button>
                                              </div>
                                              <MarkdownRenderer content={resource.ai_summary} className="max-w-none" />
                                            </div>
                                          )}
                                      </div>
                                    )
                                  })}
                                </div>
                              </CollapsibleContent>
                            </div>
                          </Collapsible>
                        )
                      })}
                    </div>
                  </CardContent>
                </CollapsibleContent>
              </Card>
            </Collapsible>
          )
        })}
      </div>

      <Dialog open={chapterDialogOpen} onOpenChange={setChapterDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {chapterDialogMode === 'create' ? '新增章节' : '编辑章节'}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">章节标题</label>
              <Input
                placeholder="输入章节标题"
                value={chapterForm.title}
                onChange={(e) =>
                  setChapterForm({ ...chapterForm, title: e.target.value })
                }
                disabled={loading}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">排序</label>
              <Input
                type="number"
                min={1}
                value={chapterForm.sort_order}
                onChange={(e) =>
                  setChapterForm({
                    ...chapterForm,
                    sort_order: parseInt(e.target.value, 10) || 1,
                  })
                }
                disabled={loading}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setChapterDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              onClick={handleSubmitChapter}
              disabled={loading || !chapterForm.title.trim()}
            >
              {loading ? '提交中...' : chapterDialogMode === 'create' ? '创建' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={sectionDialogOpen} onOpenChange={setSectionDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {sectionDialogMode === 'create' ? '新增节' : '编辑节'}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">节标题</label>
              <Input
                placeholder="输入节标题"
                value={sectionForm.title}
                onChange={(e) =>
                  setSectionForm({ ...sectionForm, title: e.target.value })
                }
                disabled={loading}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">排序</label>
              <Input
                type="number"
                min={1}
                value={sectionForm.sort_order}
                onChange={(e) =>
                  setSectionForm({
                    ...sectionForm,
                    sort_order: parseInt(e.target.value, 10) || 1,
                  })
                }
                disabled={loading}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setSectionDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              onClick={handleSubmitSection}
              disabled={loading || !sectionForm.title.trim()}
            >
              {loading ? '提交中...' : sectionDialogMode === 'create' ? '创建' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">
              确定要删除
              {!deleteTarget
                ? '该项'
                : deleteTarget.type === 'chapter'
                  ? `章节「${(deleteTarget.item as Chapter).title}」及其下全部节与资源`
                  : deleteTarget.type === 'section'
                    ? `节「${(deleteTarget.item as Section).title}」及其下全部资源`
                    : deleteTarget.type === 'resource'
                      ? `资源「${(deleteTarget.item as Resource).title}」`
                      : '该项'}
              吗？此操作不可恢复。
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={loading}
            >
              {loading ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={uploadDialogOpen} onOpenChange={setUploadDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>上传资源</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">选择文件</label>
              <Input
                type="file"
                accept=".txt,.pdf,.doc,.docx,.ppt,.pptx,.md,.markdown,.mp4,.mov,.mkv,.webm,.avi,.mp3,.wav,.ogg,.oga,.m4a,.aac,.flac,.wma,video/*,audio/*"
                onChange={handleFileChange}
                disabled={loading}
              />
              <p className="text-xs text-muted-foreground">
                支持 PDF、DOC、DOCX、PPT、TXT、视频、音频等格式
              </p>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">资源标题</label>
              <Input
                placeholder="输入资源标题"
                value={uploadForm.title}
                onChange={(e) =>
                  setUploadForm({ ...uploadForm, title: e.target.value })
                }
                disabled={loading}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">排序</label>
              <Input
                type="number"
                min={1}
                value={uploadForm.sort_order}
                onChange={(e) =>
                  setUploadForm({
                    ...uploadForm,
                    sort_order: parseInt(e.target.value, 10) || 1,
                  })
                }
                disabled={loading}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setUploadDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              onClick={handleUploadResource}
              disabled={loading || !uploadFile || !uploadForm.title.trim()}
            >
              {loading ? '上传中...' : '上传'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={lessonPlanDialogOpen} onOpenChange={setLessonPlanDialogOpen}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>
              生成小节教案草案
              {lessonPlanSection ? `：${lessonPlanSection.title}` : ''}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2 sm:col-span-2">
                <label className="text-sm font-medium">教学目标（每行 1 条）</label>
                <textarea
                  className="min-h-[110px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  placeholder={'例如：\n理解检索增强生成（RAG）的核心流程\n能够根据案例拆解查询改写步骤'}
                  value={lessonPlanForm.objectivesText}
                  onChange={(e) => setLessonPlanForm((p) => ({ ...p, objectivesText: e.target.value }))}
                  disabled={lessonPlanGenerating}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">授课风格（可选）</label>
                <Input
                  placeholder="如：案例驱动、实操优先"
                  value={lessonPlanForm.teachingStyle}
                  onChange={(e) => setLessonPlanForm((p) => ({ ...p, teachingStyle: e.target.value }))}
                  disabled={lessonPlanGenerating}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">建议时长（分钟）</label>
                <Input
                  type="number"
                  min={1}
                  value={lessonPlanForm.durationMinutes}
                  onChange={(e) =>
                    setLessonPlanForm((p) => ({ ...p, durationMinutes: parseInt(e.target.value, 10) || 45 }))
                  }
                  disabled={lessonPlanGenerating}
                />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <label className="text-sm font-medium">补充约束（可选）</label>
                <Input
                  placeholder="如：面向零基础，课堂需包含 1 次小组讨论"
                  value={lessonPlanForm.extraHint}
                  onChange={(e) => setLessonPlanForm((p) => ({ ...p, extraHint: e.target.value }))}
                  disabled={lessonPlanGenerating}
                />
              </div>
            </div>

            {lessonPlanDraft && (
              <div className="max-h-[48vh] space-y-3 overflow-y-auto rounded-md border border-border bg-muted/20 p-4 text-sm">
                <div>
                  <h4 className="font-semibold text-foreground">{lessonPlanDraft.title}</h4>
                </div>
                <div>
                  <p className="mb-1 font-medium text-foreground">教学目标</p>
                  <ul className="list-disc space-y-1 pl-5 text-muted-foreground">
                    {lessonPlanDraft.objectives.map((x, idx) => (
                      <li key={`obj-${idx}`}>{x}</li>
                    ))}
                  </ul>
                </div>
                <div>
                  <p className="mb-1 font-medium text-foreground">教学步骤</p>
                  <div className="space-y-2">
                    {lessonPlanDraft.steps.map((step, idx) => (
                      <div key={`step-${idx}`} className="rounded border border-border bg-card px-3 py-2">
                        <p className="font-medium text-foreground">
                          {idx + 1}. {step.phase}（{step.duration_minutes} min）
                        </p>
                        <ul className="mt-1 list-disc space-y-1 pl-5 text-muted-foreground">
                          {step.activities.map((a, ai) => (
                            <li key={`act-${idx}-${ai}`}>{a}</li>
                          ))}
                        </ul>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setLessonPlanDialogOpen(false)} disabled={lessonPlanGenerating}>
              关闭
            </Button>
            <Button onClick={handleGenerateLessonPlan} disabled={lessonPlanGenerating}>
              {lessonPlanGenerating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  生成中...
                </>
              ) : (
                <>
                  <Sparkles className="mr-2 h-4 w-4" />
                  生成教案
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
