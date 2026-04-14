import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  chunkPreview,
  saveChunks,
  confirmChunks,
  getChunks,
  embedResource,
  deleteChunk,
  updateChunk,
} from '@/services/knowledge'
import { getModels } from '@/services/aiModels'
import type {
  KnowledgeChunk,
  ChunkPreviewSegment,
  ChunkSaveItem,
  AIModelListItem,
} from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { toast } from 'sonner'
import {
  ArrowLeft,
  Play,
  Save,
  CheckCircle,
  Database,
  Trash2,
  Edit3,
  Eye,
  Cpu,
} from 'lucide-react'

type EditablePreviewChunk = ChunkPreviewSegment & { key: string }
type TabKey = 'chunk' | 'embed'

export function TeacherKnowledgeChunkPage() {
  const { courseId = '', resourceId = '' } = useParams<{ courseId: string; resourceId: string }>()
  const navigate = useNavigate()

  const [activeTab, setActiveTab] = useState<TabKey>('chunk')
  const [chunkSize, setChunkSize] = useState(1000)
  const [overlap, setOverlap] = useState(200)
  const [clearOnPreview, setClearOnPreview] = useState(true)
  const [previewSegments, setPreviewSegments] = useState<EditablePreviewChunk[]>([])
  const [savedChunks, setSavedChunks] = useState<KnowledgeChunk[]>([])
  const [editingSavedId, setEditingSavedId] = useState<string | null>(null)
  const [editDraft, setEditDraft] = useState('')
  const [embeddingModels, setEmbeddingModels] = useState<AIModelListItem[]>([])
  const [selectedModelId, setSelectedModelId] = useState('')
  const [loading, setLoading] = useState(false)
  const [viewer, setViewer] = useState<{ title: string; content: string } | null>(null)
  const [embedding, setEmbedding] = useState(false)

  const numericResourceId = Number(resourceId)

  const loadSavedChunks = useCallback(async () => {
    if (!numericResourceId) return
    try {
      const chunks = await getChunks(numericResourceId)
      setSavedChunks(chunks)
    } catch (err) {
      const msg = err instanceof Error ? err.message : '加载分块失败'
      toast.error(msg)
    }
  }, [numericResourceId])

  useEffect(() => {
    void loadSavedChunks()
  }, [loadSavedChunks])

  useEffect(() => {
    let cancelled = false
    async function loadModels() {
      try {
        const models = await getModels()
        if (!cancelled) {
          const emb = models.filter((m) => m.model_type === 'embedding')
          setEmbeddingModels(emb)
          if (emb.length && !selectedModelId) {
            setSelectedModelId(emb[0].id)
          }
        }
      } catch {
        // Ignore model loading errors
      }
    }
    void loadModels()
    return () => {
      cancelled = true
    }
  }, [selectedModelId])

  const savedSummary = useMemo(() => {
    const n = savedChunks.length
    if (!n) return '尚未保存分块'
    const emb = savedChunks.filter((x) => x.is_embedded).length
    const conf = savedChunks.filter((x) => x.is_confirmed).length
    return `已保存 ${n} 条 · 已确认 ${conf} 条 · 已嵌入 ${emb} 条`
  }, [savedChunks])

  function assignKeys(segs: ChunkPreviewSegment[]): EditablePreviewChunk[] {
    return segs.map((s) => ({
      ...s,
      key:
        typeof crypto !== 'undefined' && crypto.randomUUID
          ? crypto.randomUUID()
          : `${s.index}-${Math.random()}`,
    }))
  }

  async function handlePreview() {
    if (!numericResourceId) {
      toast.error('资源ID无效')
      return
    }
    setLoading(true)
    try {
      const resp = await chunkPreview(numericResourceId, {
        chunk_size: chunkSize,
        overlap,
        clear_persisted_first: clearOnPreview,
      })
      setPreviewSegments(assignKeys(resp.segments))
      await loadSavedChunks()
      toast.success(
        clearOnPreview
          ? `已生成 ${resp.segments.length} 条预览分块（已清空库内旧分块）`
          : `已生成 ${resp.segments.length} 条预览分块`
      )
    } catch (err) {
      const msg = err instanceof Error ? err.message : '预览失败'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  function updatePreviewContent(key: string, content: string) {
    setPreviewSegments((prev) => prev.map((s) => (s.key === key ? { ...s, content } : s)))
  }

  function removePreviewSegment(key: string) {
    setPreviewSegments((prev) => prev.filter((s) => s.key !== key))
  }

  const savePayloadFromPreview: ChunkSaveItem[] = useMemo(
    () =>
      previewSegments.map((s) => ({
        content: s.content,
        char_start: s.char_start,
        char_end: s.char_end,
      })),
    [previewSegments]
  )

  async function handleSavePreviewToDB() {
    if (!numericResourceId) {
      toast.error('资源ID无效')
      return
    }
    if (previewSegments.length === 0) {
      toast.error('请先生成预览分块')
      return
    }
    setLoading(true)
    try {
      await saveChunks(numericResourceId, { chunks: savePayloadFromPreview })
      toast.success('已将预览结果保存为正式分块')
      setPreviewSegments([])
      await loadSavedChunks()
    } catch (err) {
      const msg = err instanceof Error ? err.message : '保存失败'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  async function handleConfirm() {
    if (!numericResourceId) {
      toast.error('资源ID无效')
      return
    }
    setLoading(true)
    try {
      await confirmChunks(numericResourceId)
      toast.success('分块已确认，可执行向量嵌入')
      await loadSavedChunks()
    } catch (err) {
      const msg = err instanceof Error ? err.message : '确认失败'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  function startEditSaved(row: KnowledgeChunk) {
    setEditingSavedId(row.id)
    setEditDraft(row.content)
  }

  async function saveEditedChunk() {
    if (!numericResourceId || !editingSavedId) return
    setLoading(true)
    try {
      await updateChunk(numericResourceId, editingSavedId, { content: editDraft })
      toast.success('分块已更新')
      setEditingSavedId(null)
      await loadSavedChunks()
    } catch (err) {
      const msg = err instanceof Error ? err.message : '更新失败'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  async function removeSavedChunk(id: string) {
    if (!numericResourceId) return
    if (!window.confirm('确定删除该分块？删除后无法恢复。')) return
    setLoading(true)
    try {
      await deleteChunk(numericResourceId, id)
      toast.success('分块已删除')
      if (editingSavedId === id) setEditingSavedId(null)
      await loadSavedChunks()
    } catch (err) {
      const msg = err instanceof Error ? err.message : '删除失败'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  async function handleEmbed() {
    if (!numericResourceId) {
      toast.error('资源ID无效')
      return
    }
    if (!selectedModelId) {
      toast.error('请选择嵌入模型')
      return
    }
    setEmbedding(true)
    try {
      await embedResource(numericResourceId, { model_id: Number(selectedModelId) })
      toast.success('向量嵌入完成')
      await loadSavedChunks()
    } catch (err) {
      const msg = err instanceof Error ? err.message : '嵌入失败'
      toast.error(msg)
    } finally {
      setEmbedding(false)
    }
  }

  function handleBack() {
    navigate(`/teacher/knowledge/${courseId}`)
  }

  return (
    <div className="container mx-auto py-6 px-4 max-w-7xl">
      <div className="mb-6">
        <div className="flex items-center justify-between flex-wrap gap-4">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">资源分块与向量</h1>
            <p className="text-sm text-gray-500 mt-1">{savedSummary}</p>
          </div>
          <Button variant="outline" onClick={handleBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            返回资源列表
          </Button>
        </div>
      </div>

      <div className="flex gap-2 mb-6">
        <Button
          variant={activeTab === 'chunk' ? 'default' : 'outline'}
          onClick={() => setActiveTab('chunk')}
        >
          <Database className="w-4 h-4 mr-2" />
          分块管理
        </Button>
        <Button
          variant={activeTab === 'embed' ? 'default' : 'outline'}
          onClick={() => setActiveTab('embed')}
        >
          <Cpu className="w-4 h-4 mr-2" />
          向量嵌入
        </Button>
      </div>

      {activeTab === 'chunk' && (
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>一、预览分块（未保存到库）</CardTitle>
              <CardDescription>
                此处仅内存预览。保存到库请使用下方「保存预览到库」。
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
                <input
                  type="checkbox"
                  checked={clearOnPreview}
                  onChange={(e) => setClearOnPreview(e.target.checked)}
                  className="rounded border-gray-300"
                />
                预览前清空库内已保存分块（推荐，避免与旧数据混用）
              </label>

              <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      分块大小（字符）
                    </label>
                    <Input
                      type="number"
                      min={50}
                      max={32000}
                      value={chunkSize}
                      onChange={(e) => setChunkSize(Number(e.target.value))}
                      placeholder="默认1000"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      重叠（字符）
                    </label>
                    <Input
                      type="number"
                      min={0}
                      value={overlap}
                      onChange={(e) => setOverlap(Number(e.target.value))}
                      placeholder="默认200"
                    />
                  </div>
                  <Button
                    onClick={() => void handlePreview()}
                    disabled={loading}
                    className="w-full"
                  >
                    <Play className="w-4 h-4 mr-2" />
                    生成预览
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => void handleSavePreviewToDB()}
                    disabled={loading || previewSegments.length === 0}
                    className="w-full"
                  >
                    <Save className="w-4 h-4 mr-2" />
                    保存预览到库
                  </Button>
                  <Button
                    variant="secondary"
                    onClick={() => void handleConfirm()}
                    disabled={loading}
                    className="w-full"
                  >
                    <CheckCircle className="w-4 h-4 mr-2" />
                    确认分块
                  </Button>
                </div>

                <div className="lg:col-span-2 space-y-4">
                  <h4 className="text-sm font-medium text-gray-700">预览结果</h4>
                  {previewSegments.length === 0 && (
                    <p className="text-sm text-gray-500">点击「生成预览」后在此编辑或删除草稿块。</p>
                  )}
                  <div className="space-y-3 max-h-[60vh] overflow-y-auto pr-2">
                    {previewSegments.map((s, i) => (
                      <Card key={s.key} className="border-dashed">
                        <CardContent className="p-4 space-y-3">
                          <div className="flex items-center justify-between">
                            <span className="text-xs font-mono text-gray-500">
                              预览 #{i + 1}
                            </span>
                            <div className="flex gap-2">
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() =>
                                  setViewer({
                                    title: `预览分块 #${i + 1}`,
                                    content: s.content,
                                  })
                                }
                              >
                                <Eye className="w-4 h-4 mr-1" />
                                查看全文
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => removePreviewSegment(s.key)}
                                className="text-red-600 hover:text-red-700"
                              >
                                <Trash2 className="w-4 h-4 mr-1" />
                                删除
                              </Button>
                            </div>
                          </div>
                          <Textarea
                            rows={4}
                            value={s.content}
                            onChange={(e) => updatePreviewContent(s.key, e.target.value)}
                            className="text-sm"
                          />
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>二、已保存分块（正式数据）</CardTitle>
              <CardDescription>
                来自数据库；可单独编辑、删除。编辑会清空该条向量与确认状态。
              </CardDescription>
            </CardHeader>
            <CardContent>
              {savedChunks.length === 0 && (
                <p className="text-sm text-gray-500">暂无已保存分块。</p>
              )}
              <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-2">
                {savedChunks.map((row) => (
                  <Card key={row.id} className="border-gray-200">
                    <CardContent className="p-4 space-y-3">
                      <div className="flex items-center justify-between flex-wrap gap-2">
                        <span className="text-xs font-mono text-gray-500">
                          #{row.chunk_index + 1} · ID {row.id}
                          {row.is_confirmed ? ' · 已确认' : ' · 未确认'}
                          {row.is_embedded ? ' · 已嵌入' : ' · 未嵌入'}
                        </span>
                        <div className="flex gap-2">
                          {editingSavedId !== row.id && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => startEditSaved(row)}
                            >
                              <Edit3 className="w-4 h-4 mr-1" />
                              编辑
                            </Button>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() =>
                              setViewer({
                                title: `已保存分块 #${row.chunk_index + 1}`,
                                content: row.content,
                              })
                            }
                          >
                            <Eye className="w-4 h-4 mr-1" />
                            查看全文
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => void removeSavedChunk(row.id)}
                            className="text-red-600 hover:text-red-700"
                          >
                            <Trash2 className="w-4 h-4 mr-1" />
                            删除
                          </Button>
                        </div>
                      </div>
                      {editingSavedId === row.id ? (
                        <div className="space-y-3">
                          <Textarea
                            rows={6}
                            value={editDraft}
                            onChange={(e) => setEditDraft(e.target.value)}
                            className="text-sm"
                          />
                          <div className="flex gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => setEditingSavedId(null)}
                            >
                              取消
                            </Button>
                            <Button
                              size="sm"
                              onClick={() => void saveEditedChunk()}
                              disabled={loading}
                            >
                              保存修改
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <pre className="text-sm text-gray-700 whitespace-pre-wrap break-words font-sans bg-gray-50 p-3 rounded">
                          {row.content}
                        </pre>
                      )}
                    </CardContent>
                  </Card>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {activeTab === 'embed' && (
        <Card>
          <CardHeader>
            <CardTitle>向量嵌入</CardTitle>
            <CardDescription>
              仅对已确认且尚未嵌入的分块调用嵌入接口。请先完成「分块管理」中的保存与确认。
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex flex-wrap items-end gap-4">
              <div className="w-full max-w-xs">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  嵌入模型
                </label>
                <Select value={selectedModelId} onValueChange={(value) => setSelectedModelId(value ?? '')}>
                  <SelectTrigger>
                    <SelectValue placeholder="请选择模型" />
                  </SelectTrigger>
                  <SelectContent>
                    {embeddingModels.map((m) => (
                      <SelectItem key={m.id} value={m.id}>
                        {m.name} ({m.model_id})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <Button
                onClick={() => void handleEmbed()}
                disabled={embedding || !selectedModelId}
              >
                <Cpu className="w-4 h-4 mr-2" />
                {embedding ? '嵌入中...' : '开始嵌入'}
              </Button>
            </div>

            {savedChunks.length > 0 && (
              <div className="bg-gray-50 p-4 rounded-lg">
                <h4 className="text-sm font-medium text-gray-700 mb-2">当前状态</h4>
                <div className="text-sm text-gray-600 space-y-1">
                  <p>总 chunks: {savedChunks.length}</p>
                  <p>已确认: {savedChunks.filter((x) => x.is_confirmed).length}</p>
                  <p>已嵌入: {savedChunks.filter((x) => x.is_embedded).length}</p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      <Dialog open={!!viewer} onOpenChange={() => setViewer(null)}>
        <DialogContent className="max-w-4xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>{viewer?.title}</DialogTitle>
          </DialogHeader>
          <div className="overflow-y-auto max-h-[60vh]">
            <pre className="text-sm text-gray-700 whitespace-pre-wrap break-words font-sans bg-gray-50 p-4 rounded">
              {viewer?.content}
            </pre>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setViewer(null)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
