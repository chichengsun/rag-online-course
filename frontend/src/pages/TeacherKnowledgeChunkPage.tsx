import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  chunkPreview,
  confirmKnowledgeChunks,
  deleteKnowledgeChunk,
  embedResource,
  listKnowledgeChunks,
  saveKnowledgeChunks,
  updateKnowledgeChunk,
  type ChunkPreviewSegment,
  type ChunkSaveItem,
  type SavedChunkRow,
} from '../services/knowledge'
import { listAIModels, type AIModelItem } from '../services/aiModels'
import { useAuth } from '../store/auth'

type EditablePreviewChunk = ChunkPreviewSegment & { key: string }

type TabKey = 'chunk' | 'embed'

// TeacherKnowledgeChunkPage：分块管理（预览与已持久化分块分离）与向量嵌入分标签展示。
export function TeacherKnowledgeChunkPage() {
  const { token } = useAuth()
  const { courseId = '', resourceId = '' } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()
  const accessToken = token ?? ''

  const tab: TabKey = searchParams.get('tab') === 'embed' ? 'embed' : 'chunk'
  function setTab(next: TabKey) {
    setSearchParams(next === 'chunk' ? {} : { tab: 'embed' })
  }

  const [chunkSize, setChunkSize] = useState(800)
  const [overlap, setOverlap] = useState(120)
  /** 再次预览时是否先清空库内已保存分块（推荐开启，避免与历史数据混用） */
  const [clearOnPreview, setClearOnPreview] = useState(true)

  const [previewSegments, setPreviewSegments] = useState<EditablePreviewChunk[]>([])
  const [savedChunks, setSavedChunks] = useState<SavedChunkRow[]>([])
  const [editingSavedId, setEditingSavedId] = useState<string | null>(null)
  const [editDraft, setEditDraft] = useState('')

  const [embeddingModels, setEmbeddingModels] = useState<AIModelItem[]>([])
  const [selectedModelId, setSelectedModelId] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [viewer, setViewer] = useState<{ title: string; content: string } | null>(null)

  const loadSavedChunks = useCallback(async () => {
    if (!resourceId) return
    const { items } = await listKnowledgeChunks(accessToken, resourceId)
    setSavedChunks(items)
  }, [accessToken, resourceId])

  useEffect(() => {
    void loadSavedChunks()
  }, [loadSavedChunks])

  useEffect(() => {
    let cancelled = false
    async function loadModels() {
      try {
        const { items } = await listAIModels(accessToken)
        if (!cancelled) {
          const emb = items.filter((m) => m.model_type === 'embedding')
          setEmbeddingModels(emb)
          if (emb.length && !selectedModelId) setSelectedModelId(emb[0].id)
        }
      } catch {
        /* ignore */
      }
    }
    void loadModels()
    return () => {
      cancelled = true
    }
  }, [accessToken])

  const savedSummary = useMemo(() => {
    const n = savedChunks.length
    if (!n) return '尚未保存分块'
    const emb = savedChunks.filter((x) => x.embedded_at).length
    const conf = savedChunks.filter((x) => x.confirmed_at).length
    return `已保存 ${n} 条 · 已确认 ${conf} 条 · 已嵌入 ${emb} 条`
  }, [savedChunks])

  function assignKeys(segs: ChunkPreviewSegment[]): EditablePreviewChunk[] {
    return segs.map((s) => ({
      ...s,
      key: typeof crypto !== 'undefined' && crypto.randomUUID ? crypto.randomUUID() : `${s.index}-${Math.random()}`,
    }))
  }

  async function handlePreview() {
    setError('')
    setMessage('')
    setLoading(true)
    try {
      const { segments: segs } = await chunkPreview(
        accessToken,
        resourceId,
        chunkSize,
        overlap,
        clearOnPreview,
      )
      setPreviewSegments(assignKeys(segs))
      await loadSavedChunks()
      setMessage(
        clearOnPreview
          ? `已按参数生成 ${segs.length} 条预览分块（已清空库内旧分块）。仅保存在下方「保存预览到库」后才会持久化。`
          : `已生成 ${segs.length} 条预览分块（未改动库内已保存数据）。`,
      )
    } catch (e) {
      setError(e instanceof Error ? e.message : '预览失败')
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
    [previewSegments],
  )

  async function handleSavePreviewToDB() {
    setError('')
    setMessage('')
    if (previewSegments.length === 0) {
      setError('请先在「预览」区域生成分块')
      return
    }
    setLoading(true)
    try {
      await saveKnowledgeChunks(accessToken, resourceId, savePayloadFromPreview)
      setMessage('已将预览结果保存为正式分块（覆盖写入该资源下原有分块）。')
      setPreviewSegments([])
      await loadSavedChunks()
    } catch (e) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setLoading(false)
    }
  }

  async function handleConfirm() {
    setError('')
    setMessage('')
    setLoading(true)
    try {
      const { confirmed_count } = await confirmKnowledgeChunks(accessToken, resourceId)
      setMessage(`已确认 ${confirmed_count} 条分块。可在「向量嵌入」页执行嵌入。`)
      await loadSavedChunks()
    } catch (e) {
      setError(e instanceof Error ? e.message : '确认失败')
    } finally {
      setLoading(false)
    }
  }

  function startEditSaved(row: SavedChunkRow) {
    setEditingSavedId(row.id)
    setEditDraft(row.content)
  }

  async function saveEditedChunk() {
    if (!editingSavedId) return
    setError('')
    setMessage('')
    setLoading(true)
    try {
      await updateKnowledgeChunk(accessToken, resourceId, editingSavedId, { content: editDraft })
      setMessage('已更新该分块；该条向量与确认状态已清空，需重新确认后再嵌入。')
      setEditingSavedId(null)
      await loadSavedChunks()
    } catch (e) {
      setError(e instanceof Error ? e.message : '更新失败')
    } finally {
      setLoading(false)
    }
  }

  async function removeSavedChunk(id: string) {
    if (!window.confirm('确定删除该条已保存分块？删除后序号会重新连续。')) return
    setError('')
    setMessage('')
    setLoading(true)
    try {
      await deleteKnowledgeChunk(accessToken, resourceId, id)
      setMessage('已删除该分块')
      if (editingSavedId === id) setEditingSavedId(null)
      await loadSavedChunks()
    } catch (e) {
      setError(e instanceof Error ? e.message : '删除失败')
    } finally {
      setLoading(false)
    }
  }

  async function handleEmbed() {
    setError('')
    setMessage('')
    if (!selectedModelId) {
      setError('请选择嵌入模型')
      return
    }
    setLoading(true)
    try {
      const { embedded_count } = await embedResource(accessToken, resourceId, selectedModelId)
      setMessage(`嵌入完成，共写入 ${embedded_count} 条向量。`)
      await loadSavedChunks()
    } catch (e) {
      setError(e instanceof Error ? e.message : '嵌入失败')
    } finally {
      setLoading(false)
    }
  }

  function openViewer(title: string, content: string) {
    setViewer({ title, content })
  }

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <div className="action-row" style={{ justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
          <div>
            <h2>资源分块与向量</h2>
            <p className="subtle">{savedSummary}</p>
          </div>
          <button type="button" onClick={() => navigate(`/teacher/knowledge/${courseId}`)}>
            返回资源列表
          </button>
        </div>
      </header>
      {error && <div className="alert error">{error}</div>}
      {message && <div className="alert success">{message}</div>}

      <div className="card" style={{ display: 'flex', gap: 8, flexWrap: 'wrap', padding: 12 }}>
        <button type="button" className={tab === 'chunk' ? 'primary' : ''} onClick={() => setTab('chunk')}>
          分块管理
        </button>
        <button type="button" className={tab === 'embed' ? 'primary' : ''} onClick={() => setTab('embed')}>
          向量嵌入
        </button>
      </div>

      {tab === 'chunk' && (
        <div className="stack" style={{ gap: 16 }}>
          <section className="card stack">
            <h3 style={{ marginTop: 0 }}>一、预览分块（未保存到库）</h3>
            <p className="subtle" style={{ fontSize: 13 }}>
              此处仅内存预览。保存到库请使用下方「保存预览到库」。与下方「已保存分块」相互独立展示。
            </p>
            <label className="subtle" style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
              <input
                type="checkbox"
                checked={clearOnPreview}
                onChange={(e) => setClearOnPreview(e.target.checked)}
              />
              预览前清空库内已保存分块与向量（推荐，避免与旧数据混用）
            </label>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'minmax(260px, 320px) 1fr',
                gap: 16,
                alignItems: 'start',
              }}
            >
              <div className="stack">
                <label>
                  分块大小（字符）
                  <input
                    type="number"
                    min={50}
                    max={32000}
                    value={chunkSize}
                    onChange={(e) => setChunkSize(Number(e.target.value))}
                  />
                </label>
                <label>
                  重叠（字符）
                  <input type="number" min={0} value={overlap} onChange={(e) => setOverlap(Number(e.target.value))} />
                </label>
                <button type="button" className="primary" disabled={loading} onClick={() => void handlePreview()}>
                  生成预览
                </button>
                <button
                  type="button"
                  disabled={loading || previewSegments.length === 0}
                  onClick={() => void handleSavePreviewToDB()}
                >
                  保存预览到库
                </button>
                <button type="button" disabled={loading} onClick={() => void handleConfirm()}>
                  确认分块（供嵌入）
                </button>
              </div>
              <div className="stack">
                <h4 className="subtle" style={{ margin: 0 }}>
                  预览结果
                </h4>
                {previewSegments.length === 0 && <p className="subtle">点击「生成预览」后在此编辑或删除草稿块。</p>}
                <div className="stack" style={{ gap: 12, maxHeight: '52vh', overflow: 'auto' }}>
                  {previewSegments.map((s, i) => (
                    <div
                      key={s.key}
                      className="card"
                      style={{ padding: 12, boxShadow: 'none', borderStyle: 'dashed' }}
                    >
                      <div className="action-row" style={{ marginBottom: 8 }}>
                        <span className="mono subtle">预览 #{i + 1}</span>
                        <div className="action-row">
                          <button type="button" onClick={() => openViewer(`预览分块 #${i + 1}`, s.content)}>
                            查看全文
                          </button>
                          <button type="button" className="danger" onClick={() => removePreviewSegment(s.key)}>
                            删除
                          </button>
                        </div>
                      </div>
                      <textarea
                        rows={5}
                        value={s.content}
                        onChange={(e) => updatePreviewContent(s.key, e.target.value)}
                        style={{ width: '100%', fontFamily: 'inherit' }}
                      />
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </section>

          <section className="card stack">
            <h3 style={{ marginTop: 0 }}>二、已保存分块（正式数据）</h3>
            <p className="subtle" style={{ fontSize: 13 }}>
              来自数据库；可单独编辑、删除。编辑会清空该条向量与确认状态。删除后序号自动重排。
            </p>
            {savedChunks.length === 0 && <p className="subtle">暂无已保存分块。</p>}
            <div className="stack" style={{ gap: 12 }}>
              {savedChunks.map((row) => (
                <div key={row.id} className="card" style={{ padding: 12, boxShadow: 'none', border: '1px solid #e2e8f0' }}>
                  <div className="action-row" style={{ marginBottom: 8, flexWrap: 'wrap' }}>
                    <span className="mono subtle">
                      #{row.chunk_index + 1} · id {row.id}
                      {row.confirmed_at ? ' · 已确认' : ' · 未确认'}
                      {row.embedded_at ? ` · 已嵌入${row.embedding_dims != null ? `(${row.embedding_dims}维)` : ''}` : ' · 未嵌入'}
                    </span>
                    <div className="action-row">
                      {editingSavedId !== row.id && (
                        <button type="button" onClick={() => startEditSaved(row)}>
                          编辑
                        </button>
                      )}
                      <button type="button" onClick={() => openViewer(`已保存分块 #${row.chunk_index + 1}`, row.content)}>
                        查看全文
                      </button>
                      <button type="button" className="danger" onClick={() => void removeSavedChunk(row.id)}>
                        删除
                      </button>
                    </div>
                  </div>
                  {editingSavedId === row.id ? (
                    <>
                      <textarea
                        rows={6}
                        value={editDraft}
                        onChange={(e) => setEditDraft(e.target.value)}
                        style={{ width: '100%', fontFamily: 'inherit' }}
                      />
                      <div className="action-row">
                        <button type="button" onClick={() => setEditingSavedId(null)}>
                          取消
                        </button>
                        <button type="button" className="primary" disabled={loading} onClick={() => void saveEditedChunk()}>
                          保存修改
                        </button>
                      </div>
                    </>
                  ) : (
                    <pre
                      style={{
                        margin: 0,
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word',
                        fontSize: 13,
                        fontFamily: 'inherit',
                      }}
                    >
                      {row.content}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          </section>
        </div>
      )}

      {tab === 'embed' && (
        <section className="card stack">
          <h3 style={{ marginTop: 0 }}>向量嵌入</h3>
          <p className="subtle">
            仅对已确认且尚未嵌入的分块调用嵌入接口。请先完成「分块管理」中的保存与确认；此处不再重复展示分块列表。
          </p>
          <div className="action-row" style={{ flexWrap: 'wrap', alignItems: 'flex-end', gap: 12 }}>
            <label style={{ minWidth: 220 }}>
              嵌入模型
              <select value={selectedModelId} onChange={(e) => setSelectedModelId(e.target.value)}>
                <option value="">请选择</option>
                {embeddingModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.model_id})
                  </option>
                ))}
              </select>
            </label>
            <button type="button" className="primary" disabled={loading} onClick={() => void handleEmbed()}>
              开始嵌入
            </button>
          </div>
        </section>
      )}

      {viewer && (
        <div className="modal-mask" role="presentation">
          <div className="modal-card" role="dialog" aria-modal="true" style={{ maxWidth: 980, width: 'min(980px, 95vw)' }}>
            <div className="modal-card-header">
              <h3>{viewer.title}</h3>
              <button type="button" className="modal-close-btn" aria-label="关闭" onClick={() => setViewer(null)}>
                ×
              </button>
            </div>
            <pre
              style={{
                margin: 0,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                fontSize: 14,
                lineHeight: 1.6,
                fontFamily: 'inherit',
                background: '#f8fafc',
                border: '1px solid #e2e8f0',
                borderRadius: 10,
                padding: 14,
                maxHeight: '70vh',
                overflow: 'auto',
              }}
            >
              {viewer.content}
            </pre>
            <div className="action-row" style={{ justifyContent: 'flex-end' }}>
              <button type="button" onClick={() => setViewer(null)}>
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
