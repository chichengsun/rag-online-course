import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { listKnowledgeResources, type KnowledgeResourceRow } from '../services/knowledge'
import { useAuth } from '../store/auth'

function num(v: number | string | undefined): number {
  if (v === undefined || v === null) return 0
  return typeof v === 'number' ? v : Number(v)
}

// TeacherKnowledgeResourcesPage 单课下可解析资源的分页列表，提供分块与嵌入入口。
export function TeacherKnowledgeResourcesPage() {
  const { token } = useAuth()
  const { courseId = '' } = useParams()
  const navigate = useNavigate()
  const accessToken = token ?? ''
  const [items, setItems] = useState<KnowledgeResourceRow[]>([])
  const [page, setPage] = useState(1)
  const [pageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function load(p = page) {
    if (!courseId) return
    setLoading(true)
    setError('')
    try {
      const data = await listKnowledgeResources(accessToken, courseId, p, pageSize)
      setItems(data.items)
      setTotal(num(data.total))
      setPage(data.page)
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load(1)
  }, [courseId, accessToken])

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <div className="action-row" style={{ justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
          <div>
            <h2>知识库资源</h2>
            <p className="subtle">课程 ID：{courseId} · 仅列出支持解析与分块的文档类型。</p>
          </div>
          <button onClick={() => navigate('/teacher/knowledge')}>返回课程列表</button>
        </div>
      </header>
      {error && <div className="alert error">{error}</div>}
      <section className="card stack">
        {loading && <p className="subtle">加载中…</p>}
        {!loading && items.length === 0 && <p className="subtle">暂无符合条件的资源。</p>}
        <div className="course-list">
          {items.map((r) => (
            <article key={r.id} className="course-item">
              <div>
                <h4>{r.title}</h4>
                <p className="subtle">
                  {r.chapter_title} · {r.resource_type}
                </p>
                <p className="mono subtle">
                  分块数 {num(r.chunk_count)} · 已嵌入 {num(r.embedded_count)} · 分块字符合计 {num(r.total_chunk_chars)}
                </p>
              </div>
              <div className="action-row">
                <button onClick={() => navigate(`/teacher/knowledge/${courseId}/chunk/${r.id}`)}>分块</button>
                <button
                  className="primary"
                  onClick={() => navigate(`/teacher/knowledge/${courseId}/chunk/${r.id}?tab=embed`)}
                >
                  嵌入
                </button>
              </div>
            </article>
          ))}
        </div>
        <div className="pagination-bar">
          <button disabled={page <= 1 || loading} onClick={() => void load(page - 1)}>
            上一页
          </button>
          <span className="mono">
            第 {page} / {totalPages} 页（共 {total} 条）
          </span>
          <button disabled={page >= totalPages || loading} onClick={() => void load(page + 1)}>
            下一页
          </button>
        </div>
      </section>
    </div>
  )
}
