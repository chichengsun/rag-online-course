import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import {
  listQuestionBank,
  updateQuestion,
  deleteQuestion,
  QUESTION_TYPE_LABELS,
  QUESTION_TYPE_OPTIONS,
  type QuestionType,
  type QuestionBankItem,
  type QuestionBankPayload,
} from '@/services/questionBank'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

const emptyPayload = (): QuestionBankPayload => ({
  question_type: 'single_choice',
  stem: '',
  reference_answer: '',
})

/** 已入库题目分页列表；编辑在弹窗中进行，与维护页分离。 */
export function TeacherQuestionBankListPage() {
  const { courseId } = useParams<{ courseId: string }>()
  const { token } = useAuthStore()
  const accessToken = token ?? ''
  const courseIdNum = Number(courseId || 0)

  const [items, setItems] = useState<QuestionBankItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)
  const [keywordInput, setKeywordInput] = useState('')
  const [appliedKeyword, setAppliedKeyword] = useState('')
  const [typeFilter, setTypeFilter] = useState<QuestionType | ''>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [deletingBatch, setDeletingBatch] = useState(false)

  const [editOpen, setEditOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editForm, setEditForm] = useState<QuestionBankPayload>(emptyPayload())
  const [savingEdit, setSavingEdit] = useState(false)

  const loadItems = useCallback(async () => {
    if (!courseIdNum) return
    setLoading(true)
    setError(null)
    try {
      const out = await listQuestionBank(accessToken, courseIdNum, {
        page,
        page_size: pageSize,
        keyword: appliedKeyword || undefined,
        question_type: typeFilter || undefined,
      })
      setItems(out.items)
      setTotal(out.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载题库失败')
    } finally {
      setLoading(false)
    }
  }, [accessToken, courseIdNum, page, pageSize, appliedKeyword, typeFilter])

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  const openEdit = (item: QuestionBankItem) => {
    setEditingId(Number(item.id))
    setEditForm({
      question_type: item.question_type as QuestionType,
      stem: item.stem,
      reference_answer: item.reference_answer,
    })
    setEditOpen(true)
  }

  const closeEdit = () => {
    setEditOpen(false)
    setEditingId(null)
    setEditForm(emptyPayload())
  }

  const handleEditSave = async () => {
    if (!editingId) return
    setSavingEdit(true)
    setError(null)
    try {
      await updateQuestion(accessToken, editingId, editForm)
      closeEdit()
      await loadItems()
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSavingEdit(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!window.confirm('确定删除该题目吗？')) return
    setError(null)
    try {
      await deleteQuestion(accessToken, id)
      await loadItems()
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除失败')
    }
  }

  const toggleSelected = (id: number, checked: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (checked) next.add(id)
      else next.delete(id)
      return next
    })
  }

  const handleDeleteSelected = async () => {
    if (selectedIds.size === 0) return
    if (!window.confirm(`确定删除已选中的 ${selectedIds.size} 道题吗？`)) return
    setDeletingBatch(true)
    setError(null)
    try {
      for (const id of selectedIds) {
        await deleteQuestion(accessToken, id)
      }
      setSelectedIds(new Set())
      await loadItems()
    } catch (err) {
      setError(err instanceof Error ? err.message : '批量删除失败')
    } finally {
      setDeletingBatch(false)
    }
  }

  const allPageSelected =
    items.length > 0 && items.every((item) => selectedIds.has(Number(item.id)))

  const toggleSelectAllPage = (checked: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      for (const item of items) {
        const id = Number(item.id)
        if (checked) next.add(id)
        else next.delete(id)
      }
      return next
    })
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const applyFilters = () => {
    setAppliedKeyword(keywordInput.trim())
    setPage(1)
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>已入库题目</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-2 items-end">
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">关键词（题干 / 答案 / 类型）</span>
              <Input
                className="w-64"
                placeholder="输入后点「搜索」"
                value={keywordInput}
                onChange={(e) => setKeywordInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && applyFilters()}
              />
            </div>
            <select
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter((e.target.value || '') as QuestionType | '')
                setPage(1)
              }}
            >
              <option value="">全部题型</option>
              {QUESTION_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <Button variant="secondary" onClick={applyFilters}>
              搜索
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleDeleteSelected}
              disabled={selectedIds.size === 0 || deletingBatch}
            >
              {deletingBatch ? '删除中...' : `批量删除（${selectedIds.size}）`}
            </Button>
          </div>

          {!loading && items.length > 0 && (
            <label className="flex items-center gap-2 text-sm text-muted-foreground">
              <input
                type="checkbox"
                checked={allPageSelected}
                onChange={(e) => toggleSelectAllPage(e.target.checked)}
              />
              全选本页（{items.length}）
            </label>
          )}

          {loading ? (
            <p className="text-sm text-muted-foreground">加载中...</p>
          ) : items.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无题目，请前往「新增与 AI 导题」添加。</p>
          ) : (
            items.map((item) => (
              <div key={item.id} className="rounded-lg border border-border p-3">
                <label className="mb-2 flex items-center gap-2 text-xs text-muted-foreground">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(Number(item.id))}
                    onChange={(e) => toggleSelected(Number(item.id), e.target.checked)}
                  />
                  选择
                </label>
                <p className="text-sm font-medium mb-1 whitespace-pre-wrap">{item.stem}</p>
                <p className="text-xs text-muted-foreground mb-1">
                  类型：{QUESTION_TYPE_LABELS[item.question_type] || item.question_type}
                </p>
                <p className="text-sm">参考答案：{item.reference_answer}</p>
                {item.source_file_name ? (
                  <p className="text-xs text-muted-foreground mt-1">来源：{item.source_file_name}</p>
                ) : null}
                <div className="mt-2 flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => openEdit(item)}>
                    编辑
                  </Button>
                  <Button variant="destructive" size="sm" onClick={() => void handleDelete(Number(item.id))}>
                    删除
                  </Button>
                </div>
              </div>
            ))
          )}

          {!loading && total > 0 && (
            <div className="flex flex-wrap items-center justify-between gap-3 pt-4 border-t">
              <p className="text-sm text-muted-foreground">
                共 {total} 条 · 第 {page} / {totalPages} 页
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  下一页
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={editOpen} onOpenChange={(open) => !open && closeEdit()}>
        <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto" showCloseButton>
          <DialogHeader>
            <DialogTitle>编辑题目{editingId != null ? `（ID: ${editingId}）` : ''}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <select
              className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
              value={editForm.question_type}
              onChange={(e) => setEditForm((p) => ({ ...p, question_type: e.target.value as QuestionType }))}
            >
              {QUESTION_TYPE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}（{opt.value}）
                </option>
              ))}
            </select>
            <Textarea
              placeholder="题干"
              rows={5}
              value={editForm.stem}
              onChange={(e) => setEditForm((p) => ({ ...p, stem: e.target.value }))}
            />
            <Textarea
              placeholder="参考答案"
              rows={3}
              value={editForm.reference_answer}
              onChange={(e) => setEditForm((p) => ({ ...p, reference_answer: e.target.value }))}
            />
          </div>
          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={closeEdit} disabled={savingEdit}>
              取消
            </Button>
            <Button
              disabled={
                savingEdit ||
                !editForm.question_type ||
                !editForm.stem.trim() ||
                !editForm.reference_answer.trim()
              }
              onClick={() => void handleEditSave()}
            >
              {savingEdit ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
