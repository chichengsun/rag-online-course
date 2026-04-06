import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createCourse, deleteCourse, listTeacherCourses, type TeacherCourseItem } from '../services/course'
import { useAuth } from '../store/auth'

export function TeacherCoursesPage() {
  const { token } = useAuth()
  const navigate = useNavigate()
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [courses, setCourses] = useState<TeacherCourseItem[]>([])
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [sortBy, setSortBy] = useState('created_at')
  const [sortOrder, setSortOrder] = useState('desc')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [createForm, setCreateForm] = useState({ title: '', description: '' })
  const accessToken = token ?? ''
  const skipKeywordDebounceRef = useRef(true)

  function runSearch() {
    void loadCourses(1, pageSize, keyword, statusFilter, sortBy, sortOrder)
  }

  async function loadCourses(
    targetPage = page,
    targetSize = pageSize,
    targetKeyword = keyword,
    targetStatus = statusFilter,
    targetSortBy = sortBy,
    targetSortOrder = sortOrder,
  ) {
    setLoading(true)
    setError('')
    try {
      const data = await listTeacherCourses(
        accessToken,
        targetPage,
        targetSize,
        targetKeyword,
        targetStatus,
        targetSortBy,
        targetSortOrder,
      )
      setCourses(data.items)
      setTotal(data.total)
      setPage(data.page)
      setPageSize(data.page_size)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载课程列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadCourses(1, 10, '', '', 'created_at', 'desc')
  }, [])

  useEffect(() => {
    if (skipKeywordDebounceRef.current) {
      skipKeywordDebounceRef.current = false
      return
    }
    const t = window.setTimeout(() => {
      void loadCourses(1, pageSize, keyword, statusFilter, sortBy, sortOrder)
    }, 300)
    return () => window.clearTimeout(t)
  }, [keyword])

  async function handleCreateCourse() {
    setLoading(true)
    setError('')
    try {
      await createCourse(accessToken, createForm)
      setShowCreateModal(false)
      setCreateForm({ title: '', description: '' })
      await loadCourses(1, pageSize, keyword, statusFilter, sortBy, sortOrder)
      setMessage('课程创建成功')
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建课程失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <h2>课程列表管理</h2>
        <p className="subtle">支持搜索、分页、创建和进入课程目录管理。</p>
      </header>
      {message && <div className="alert success">{message}</div>}
      {error && <div className="alert error">{error}</div>}

      <section className="card stack">
        <div className="list-toolbar">
          <div className="list-toolbar-actions">
            <input
              placeholder="搜索课程标题或描述（支持回车）"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') runSearch()
              }}
            />
            <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              <option value="">全部状态</option>
              <option value="draft">draft</option>
              <option value="published">published</option>
              <option value="archived">archived</option>
            </select>
            <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
              <option value="created_at">创建时间</option>
              <option value="updated_at">更新时间</option>
              <option value="title">标题</option>
            </select>
            <select value={sortOrder} onChange={(e) => setSortOrder(e.target.value)}>
              <option value="desc">降序</option>
              <option value="asc">升序</option>
            </select>
            <button onClick={() => runSearch()} disabled={loading}>
              搜索
            </button>
          </div>
          <button className="primary" onClick={() => setShowCreateModal(true)}>
            新增课程
          </button>
        </div>

        <div className="course-list">
          {courses.length === 0 && !loading && (
            <p className="subtle" style={{ padding: '1rem 0' }}>
              暂无课程。可尝试调整关键词或状态筛选，或点击「新增课程」。
            </p>
          )}
          {courses.map((course) => (
            <article key={course.id} className="course-item">
              <div>
                <h4>{course.title}</h4>
                <p className="subtle">{course.description || '暂无描述'}</p>
                <p className="mono">状态：{course.status}</p>
              </div>
              <div className="action-row">
                <button className="primary" onClick={() => navigate(`/teacher/knowledge/${course.id}`)}>
                  知识库管理
                </button>
                <button onClick={() => navigate(`/teacher/course-content/${course.id}`)}>
                  进入章节与资源管理
                </button>
                <button
                  className="danger"
                  onClick={async () => {
                    if (!window.confirm(`确定删除课程「${course.title}」？此操作不可恢复。`)) return
                    try {
                      await deleteCourse(accessToken, course.id)
                      await loadCourses(page, pageSize, keyword, statusFilter, sortBy, sortOrder)
                      setMessage('课程已删除')
                    } catch (err) {
                      setError(err instanceof Error ? err.message : '删除课程失败')
                    }
                  }}
                >
                  删除
                </button>
              </div>
            </article>
          ))}
        </div>

        <div className="pagination-bar">
          <button
            disabled={page <= 1 || loading}
            onClick={() => void loadCourses(page - 1, pageSize, keyword, statusFilter, sortBy, sortOrder)}
          >
            上一页
          </button>
          <span className="mono">
            第 {page} / {Math.max(1, Math.ceil(total / pageSize))} 页（共 {total} 条）
          </span>
          <button
            disabled={page >= Math.max(1, Math.ceil(total / pageSize)) || loading}
            onClick={() => void loadCourses(page + 1, pageSize, keyword, statusFilter, sortBy, sortOrder)}
          >
            下一页
          </button>
        </div>
      </section>

      {showCreateModal && (
        <div className="modal-mask" onClick={() => setShowCreateModal(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h3>新增课程</h3>
            <label>
              课程标题
              <input
                value={createForm.title}
                onChange={(e) => setCreateForm((v) => ({ ...v, title: e.target.value }))}
              />
            </label>
            <label>
              课程描述
              <textarea
                rows={4}
                value={createForm.description}
                onChange={(e) => setCreateForm((v) => ({ ...v, description: e.target.value }))}
              />
            </label>
            <div className="action-row">
              <button onClick={() => setShowCreateModal(false)}>取消</button>
              <button className="primary" disabled={!createForm.title || loading} onClick={handleCreateCourse}>
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
