import { useCallback, useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/authStore'
import { getCourses } from '@/services/course'
import { getModels } from '@/services/aiModels'
import { applyOutlineDraft, generateOutlineDraft } from '@/services/courseDesign'
import type { AIModelListItem, OutlineChapterDraft } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ArrowLeft, Loader2, Sparkles, Plus, CheckCircle2 } from 'lucide-react'

type LocationState = { title?: string }

/**
 * 单门课程的 AI 大纲草案页：生成、可编辑预览、追加写入章/节（排在现有章节之后）。
 */
export function TeacherCourseDesignPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { courseId: courseIdParam } = useParams<{ courseId: string }>()
  const { token } = useAuthStore()
  const accessToken = token ?? ''

  const courseId = Number(courseIdParam)
  const state = (location.state ?? {}) as LocationState

  const [courseTitle, setCourseTitle] = useState(state.title ?? '')
  const [extraHint, setExtraHint] = useState('')
  const [qaModelId, setQaModelId] = useState('')
  const [qaModels, setQaModels] = useState<AIModelListItem[]>([])
  const [chapters, setChapters] = useState<OutlineChapterDraft[]>([])
  const [generating, setGenerating] = useState(false)
  const [applying, setApplying] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const selectedQaModel = useMemo(
    () => qaModels.find((m) => String(m.id) === qaModelId),
    [qaModels, qaModelId],
  )
  const qaModelTriggerLabel = selectedQaModel
    ? `${selectedQaModel.name} (${selectedQaModel.model_id})`
    : undefined

  const loadQAModels = useCallback(async () => {
    try {
      const data = await getModels()
      const list = data.filter((m) => m.model_type === 'qa')
      setQaModels(list)
      if (list.length > 0 && !qaModelId) {
        setQaModelId(String(list[0].id))
      }
    } catch (e) {
      console.warn('加载问答模型列表失败', e)
    }
  }, [qaModelId])

  useEffect(() => {
    void loadQAModels()
  }, [loadQAModels])

  useEffect(() => {
    if (courseTitle || !Number.isFinite(courseId) || courseId <= 0 || !accessToken) return
    let cancelled = false
    void getCourses(accessToken, { page: 1, page_size: 100 })
      .then((data) => {
        if (cancelled) return
        const c = data.items.find((x) => x.id === courseId)
        if (c) setCourseTitle(c.title)
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [accessToken, courseId, courseTitle])

  const updateChapterTitle = (idx: number, title: string) => {
    setChapters((prev) => {
      const next = [...prev]
      if (!next[idx]) return prev
      next[idx] = { ...next[idx], title }
      return next
    })
  }

  const updateSectionTitle = (ci: number, si: number, title: string) => {
    setChapters((prev) => {
      const next = [...prev]
      const ch = next[ci]
      if (!ch?.sections[si]) return prev
      const secs = [...ch.sections]
      secs[si] = { ...secs[si], title }
      next[ci] = { ...ch, sections: secs }
      return next
    })
  }

  const addSection = (ci: number) => {
    setChapters((prev) => {
      const next = [...prev]
      const ch = next[ci]
      if (!ch) return prev
      next[ci] = {
        ...ch,
        sections: [...ch.sections, { title: '新小节' }],
      }
      return next
    })
  }

  const handleGenerate = async () => {
    if (!Number.isFinite(courseId) || courseId <= 0) {
      toast.error('课程 ID 无效')
      return
    }
    setGenerating(true)
    setError(null)
    try {
      const data = await generateOutlineDraft(accessToken, courseId, {
        qa_model_id: qaModelId ? Number(qaModelId) : undefined,
        extra_hint: extraHint.trim() || undefined,
      })
      setChapters(data.chapters ?? [])
      toast.success('已生成大纲草案，可修改后再应用到课程')
    } catch (err) {
      const msg = err instanceof Error ? err.message : '生成失败'
      setError(msg)
      toast.error(msg)
    } finally {
      setGenerating(false)
    }
  }

  const handleApply = async () => {
    if (!Number.isFinite(courseId) || courseId <= 0) {
      toast.error('课程 ID 无效')
      return
    }
    if (chapters.length === 0) {
      toast.error('请先生成或编辑大纲')
      return
    }
    setApplying(true)
    setError(null)
    try {
      const out = await applyOutlineDraft(accessToken, courseId, chapters)
      toast.success(`已追加 ${out.created_chapters} 章、${out.created_sections} 节，可在课程与内容中继续编辑`)
      navigate(`/teacher/course-content/${courseId}`)
    } catch (err) {
      const msg = err instanceof Error ? err.message : '应用失败'
      setError(msg)
      toast.error(msg)
    } finally {
      setApplying(false)
    }
  }

  const displayTitle = courseTitle || `课程 #${courseIdParam}`

  if (!Number.isFinite(courseId) || courseId <= 0) {
    return (
      <div className="space-y-4">
        <p className="text-destructive">无效的课程 ID</p>
        <Button variant="outline" onClick={() => navigate('/teacher/course-design')}>
          返回
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-4">
        <Button variant="outline" size="sm" onClick={() => navigate('/teacher/course-design')}>
          <ArrowLeft className="mr-2 size-4" />
          返回列表
        </Button>
        <div className="min-w-0 flex-1">
          <h1 className="text-2xl font-semibold text-foreground">课程设计</h1>
          <p className="truncate text-sm text-muted-foreground">{displayTitle}</p>
        </div>
        <Button variant="outline" size="sm" onClick={() => navigate(`/teacher/course-content/${courseId}`)}>
          打开课程与内容
        </Button>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">生成设置</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground">补充说明（可选）</label>
            <Textarea
              placeholder="例如：面向零基础、侧重实验、共 16 周…"
              value={extraHint}
              onChange={(e) => setExtraHint(e.target.value)}
              disabled={generating}
              rows={3}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground">将一并交给模型，用于贴合先修、周次或侧重点</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-foreground">问答模型</label>
            <Select
              value={qaModelId}
              onValueChange={(v) => v && setQaModelId(v)}
              disabled={qaModels.length === 0 || generating}
            >
              <SelectTrigger className="w-full min-w-0 max-w-md">
                <SelectValue placeholder="选择问答模型">{qaModelTriggerLabel}</SelectValue>
              </SelectTrigger>
              <SelectContent>
                {qaModels.map((model) => (
                  <SelectItem key={String(model.id)} value={String(model.id)}>
                    {model.name} ({model.model_id})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {qaModels.length === 0 && (
              <p className="text-xs text-destructive">未配置 QA 模型，请先在「RAG 对话 → AI 模型」中添加</p>
            )}
          </div>
          <Button onClick={() => void handleGenerate()} disabled={generating || qaModels.length === 0}>
            {generating ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                生成中…
              </>
            ) : (
              <>
                <Sparkles className="mr-2 size-4" />
                生成大纲草案
              </>
            )}
          </Button>
        </CardContent>
      </Card>

      {chapters.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-2 space-y-0">
            <CardTitle className="text-base">大纲预览（可编辑）</CardTitle>
            <p className="text-xs text-muted-foreground">应用后将追加到现有章节之后，不覆盖已有章/节</p>
          </CardHeader>
          <CardContent className="space-y-6">
            {chapters.map((ch, ci) => (
              <div key={ci} className="rounded-lg border border-border bg-muted/20 p-4">
                <div className="mb-3 flex flex-wrap items-center gap-2">
                  <span className="text-xs font-medium text-muted-foreground">第 {ci + 1} 章</span>
                  <Input
                    value={ch.title}
                    onChange={(e) => updateChapterTitle(ci, e.target.value)}
                    className="max-w-xl flex-1 font-medium"
                  />
                </div>
                <ul className="space-y-2 pl-2">
                  {ch.sections.map((sec, si) => (
                    <li key={si} className="flex items-center gap-2">
                      <span className="w-8 shrink-0 text-xs text-muted-foreground">{si + 1}.</span>
                      <Input
                        value={sec.title}
                        onChange={(e) => updateSectionTitle(ci, si, e.target.value)}
                        className="flex-1"
                      />
                    </li>
                  ))}
                </ul>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="mt-2"
                  onClick={() => addSection(ci)}
                >
                  <Plus className="mr-1 size-3" />
                  添加小节
                </Button>
              </div>
            ))}
            <Button
              onClick={() => void handleApply()}
              disabled={applying}
              className="w-full sm:w-auto"
            >
              {applying ? (
                <>
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  写入中…
                </>
              ) : (
                <>
                  <CheckCircle2 className="mr-2 size-4" />
                  将大纲追加到课程
                </>
              )}
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
