import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listTeacherCourses, type TeacherCourseItem } from '../services/course'
import { createChatSession, listChatSessions, type ChatSessionItem } from '../services/knowledgeChat'
import { useAuth } from '../store/auth'

function toNum(v: number | string | undefined): number {
  if (v === undefined) return 0
  return typeof v === 'number' ? v : Number(v)
}

// TeacherKnowledgeChatsPage 展示对话管理入口：分页列表 + 新建会话。
export function TeacherKnowledgeChatsPage() {
  const { token } = useAuth()
  const navigate = useNavigate()
  const accessToken = token ?? ''
  const [courses, setCourses] = useState<TeacherCourseItem[]>([])
  const [courseId, setCourseId] = useState('')
  const [sessions, setSessions] = useState<ChatSessionItem[]>([])
  const [page, setPage] = useState(1)
  const [pageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [creating, setCreating] = useState(false)
  const [newTitle, setNewTitle] = useState('')
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  useEffect(() => {
    let cancelled = false
    async function loadCourses() {
      try {
        const data = await listTeacherCourses(accessToken, 1, 100, '', '', 'updated_at', 'desc')
        if (!cancelled) {
          setCourses(data.items)
          if (data.items.length > 0) setCourseId((v) => v || data.items[0].id)
        }
      } catch {
        /* noop */
      }
    }
    void loadCourses()
    return () => {
      cancelled = true
    }
  }, [accessToken])

  async function loadSessions(targetPage = page, targetCourseId = courseId) {
    setLoading(true)
    setError('')
    try {
      const data = await listChatSessions(accessToken, targetPage, pageSize, targetCourseId || undefined)
      setSessions(data.items)
      setTotal(toNum(data.total))
      setPage(data.page)
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载对话记录失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadSessions(1, courseId)
  }, [courseId, accessToken])

  const totalPages = useMemo(() => Math.max(1, Math.ceil(total / pageSize)), [total, pageSize])

  async function handleCreate() {
    if (!courseId) {
      setError('请先选择课程')
      return
    }
    setCreating(true)
    setError('')
    try {
      const data = await createChatSession(accessToken, courseId, newTitle.trim())
      setMessage('已创建新对话')
      setNewTitle('')
      navigate(`/teacher/knowledge/chats/${data.id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : '创建对话失败')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <h2>对话管理</h2>
        <p className="subtle">按分页查看知识库对话记录；点击进入详情查看历史并继续问答。</p>
      </header>
      {error && <div className="alert error">{error}</div>}
      {message && <div className="alert success">{message}</div>}

      <section className="card stack">
        <div className="action-row" style={{ gap: 12, alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <label style={{ minWidth: 240 }}>
            课程
            <select value={courseId} onChange={(e) => setCourseId(e.target.value)}>
              {courses.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.title}
                </option>
              ))}
            </select>
          </label>
          <label style={{ minWidth: 260 }}>
            新对话标题（可选）
            <input value={newTitle} onChange={(e) => setNewTitle(e.target.value)} placeholder="默认：新对话" />
          </label>
          <button type="button" className="primary" disabled={creating || !courseId} onClick={() => void handleCreate()}>
            新建对话
          </button>
          <button type="button" disabled={loading} onClick={() => void loadSessions(page, courseId)}>
            刷新
          </button>
        </div>
      </section>

      <section className="card stack">
        {loading && <p className="subtle">加载中...</p>}
        {!loading && sessions.length === 0 && <p className="subtle">暂无对话记录。</p>}
        <div className="course-list">
          {sessions.map((s) => (
            <article key={s.id} className="course-item">
              <div>
                <h4>{s.title}</h4>
                <p className="subtle">课程ID：{s.course_id}</p>
                <p className="mono subtle">
                  消息数：{toNum(s.message_count)} · 最近更新：{new Date(s.updated_at).toLocaleString()}
                </p>
              </div>
              <div className="action-row">
                <button className="primary" type="button" onClick={() => navigate(`/teacher/knowledge/chats/${s.id}`)}>
                  进入对话
                </button>
              </div>
            </article>
          ))}
        </div>
        <div className="pagination-bar">
          <button disabled={loading || page <= 1} onClick={() => void loadSessions(page - 1, courseId)}>
            上一页
          </button>
          <span className="mono">
            第 {page} / {totalPages} 页（共 {total} 条）
          </span>
          <button disabled={loading || page >= totalPages} onClick={() => void loadSessions(page + 1, courseId)}>
            下一页
          </button>
        </div>
      </section>
    </div>
  )
}
