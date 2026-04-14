import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import {
  createQuestion,
  parseImportFile,
  confirmImportBatch,
  QUESTION_TYPE_OPTIONS,
  type QuestionBankPayload,
  type QuestionType,
} from '@/services/questionBank'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

type DraftRow = QuestionBankPayload & { key: string }

function newDraftKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

function payloadsFromDrafts(rows: DraftRow[]): QuestionBankPayload[] {
  return rows.map(({ question_type, stem, reference_answer }) => ({
    question_type,
    stem,
    reference_answer,
  }))
}

/** 手工新增与 AI 解析导题（解析结果为草稿，确认后入库）。 */
export function TeacherQuestionBankMaintainPage() {
  const navigate = useNavigate()
  const { courseId } = useParams<{ courseId: string }>()
  const { token } = useAuthStore()
  const accessToken = token ?? ''
  const courseIdNum = Number(courseId || 0)

  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [file, setFile] = useState<File | null>(null)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [filePreviewText, setFilePreviewText] = useState('')
  const [parsing, setParsing] = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [drafts, setDrafts] = useState<DraftRow[]>([])
  const [creating, setCreating] = useState(false)

  const [createForm, setCreateForm] = useState<QuestionBankPayload>({
    question_type: 'single_choice',
    stem: '',
    reference_answer: '',
  })

  const showSuccess = (msg: string) => {
    setSuccess(msg)
    window.setTimeout(() => setSuccess(null), 4000)
  }

  const handlePreparePreview = async () => {
    if (!file) return
    setError(null)
    try {
      const text = await file.text()
      setFilePreviewText(text)
      setPreviewOpen(true)
    } catch {
      setError('读取文件失败，请确认文件编码为 UTF-8')
    }
  }

  const handleParse = async () => {
    if (!courseIdNum || !file) return
    setParsing(true)
    setError(null)
    try {
      const questions = await parseImportFile(accessToken, courseIdNum, file)
      setDrafts(
        questions.map((q) => ({
          ...q,
          key: newDraftKey(),
        })),
      )
      if (questions.length === 0) {
        setError('模型未返回有效题目，请检查文件内容')
      } else {
        showSuccess(`已解析 ${questions.length} 道题，可在下方编辑后确认入库`)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '解析失败')
    } finally {
      setParsing(false)
    }
  }

  const handleConfirmDrafts = async () => {
    if (!courseIdNum || drafts.length === 0) return
    for (const d of drafts) {
      if (!d.stem.trim() || !d.reference_answer.trim() || !d.question_type) {
        setError('请补全每道题的类型、题干与参考答案后再入库')
        return
      }
    }
    setConfirming(true)
    setError(null)
    try {
      const { created_count } = await confirmImportBatch(accessToken, courseIdNum, {
        source_file_name: file?.name ?? '',
        questions: payloadsFromDrafts(drafts),
      })
      setDrafts([])
      setFile(null)
      showSuccess(`已成功入库 ${created_count} 道题`)
      navigate(`/teacher/question-bank/${courseIdNum}/list`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '入库失败')
    } finally {
      setConfirming(false)
    }
  }

  const handleCreateSubmit = async () => {
    if (!courseIdNum) return
    setCreating(true)
    setError(null)
    try {
      await createQuestion(accessToken, courseIdNum, createForm)
      setCreateForm({ question_type: 'single_choice', stem: '', reference_answer: '' })
      showSuccess('新增成功')
      navigate(`/teacher/question-bank/${courseIdNum}/list`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '提交失败')
    } finally {
      setCreating(false)
    }
  }

  const updateDraft = (index: number, patch: Partial<QuestionBankPayload>) => {
    setDrafts((prev) => {
      const next = [...prev]
      const row = next[index]
      if (!row) return prev
      next[index] = { ...row, ...patch }
      return next
    })
  }

  const removeDraft = (index: number) => {
    setDrafts((prev) => prev.filter((_, i) => i !== index))
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}
      {success && (
        <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
          <p className="text-sm text-emerald-800 dark:text-emerald-200">{success}</p>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>上传文本 · AI 解析</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">
            先预览完整文件内容，再调用模型解析；解析结果为草稿，编辑无误后「确认入库」才会写入题库。
          </p>
          <Input type="file" accept=".txt,.md,.csv" onChange={(e) => setFile(e.target.files?.[0] ?? null)} />
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" disabled={!file} onClick={() => void handlePreparePreview()}>
              预览完整文件内容
            </Button>
            <Button disabled={!file || parsing} onClick={() => void handleParse()}>
              {parsing ? '解析中...' : 'AI 解析为草稿'}
            </Button>
          </div>
          {previewOpen && (
            <div className="rounded-lg border border-border bg-muted/30 p-3">
              <p className="text-sm mb-2 font-medium">
                文件预览（全文）· {file?.name}
                <span className="ml-2 text-muted-foreground font-normal">
                  {filePreviewText.length > 0 ? `共 ${filePreviewText.length} 字符` : ''}
                </span>
              </p>
              <pre className="text-xs whitespace-pre-wrap max-h-[70vh] overflow-auto rounded-md bg-background p-3 border">
                {filePreviewText || '（空内容）'}
              </pre>
              <Button variant="outline" size="sm" className="mt-3" onClick={() => setPreviewOpen(false)}>
                关闭预览
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {drafts.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>解析草稿（{drafts.length}）— 编辑后确认入库</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {drafts.map((row, index) => (
              <div key={row.key} className="rounded-lg border border-border p-3 space-y-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <span className="text-xs text-muted-foreground">第 {index + 1} 题</span>
                  <Button variant="ghost" size="sm" className="text-destructive" onClick={() => removeDraft(index)}>
                    移除本题
                  </Button>
                </div>
                <select
                  className="h-10 w-full max-w-md rounded-md border border-input bg-background px-3 text-sm"
                  value={row.question_type}
                  onChange={(e) => updateDraft(index, { question_type: e.target.value as QuestionType })}
                >
                  {QUESTION_TYPE_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}（{opt.value}）
                    </option>
                  ))}
                </select>
                <Textarea
                  placeholder="题干（选择题请包含选项）"
                  rows={4}
                  value={row.stem}
                  onChange={(e) => updateDraft(index, { stem: e.target.value })}
                />
                <Textarea
                  placeholder="参考答案"
                  rows={2}
                  value={row.reference_answer}
                  onChange={(e) => updateDraft(index, { reference_answer: e.target.value })}
                />
              </div>
            ))}
            <div className="flex flex-wrap gap-2 pt-2">
              <Button disabled={confirming} onClick={() => void handleConfirmDrafts()}>
                {confirming ? '入库中...' : '确认入库'}
              </Button>
              <Button variant="outline" onClick={() => setDrafts([])} disabled={confirming}>
                清空草稿
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>手工新增单题</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <select
            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
            value={createForm.question_type}
            onChange={(e) => setCreateForm((p) => ({ ...p, question_type: e.target.value as QuestionType }))}
          >
            {QUESTION_TYPE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}（{opt.value}）
              </option>
            ))}
          </select>
          <Textarea
            placeholder="题目题干（选择题请把选项一并写在题干中）"
            rows={4}
            value={createForm.stem}
            onChange={(e) => setCreateForm((p) => ({ ...p, stem: e.target.value }))}
          />
          <Textarea
            placeholder="题目参考答案"
            rows={3}
            value={createForm.reference_answer}
            onChange={(e) => setCreateForm((p) => ({ ...p, reference_answer: e.target.value }))}
          />
          <Button
            disabled={
              creating || !createForm.question_type || !createForm.stem || !createForm.reference_answer
            }
            onClick={() => void handleCreateSubmit()}
          >
            {creating ? '提交中...' : '新增并前往列表'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
