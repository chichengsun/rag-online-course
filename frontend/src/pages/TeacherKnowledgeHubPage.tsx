import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { listTeacherCourses, type TeacherCourseItem } from '../services/course'
import { useAuth } from '../store/auth'

// TeacherKnowledgeHubPage 教师侧「课程知识库」入口：列出课程并跳转到单课资源列表。
export function TeacherKnowledgeHubPage() {
  const { token } = useAuth()
  const navigate = useNavigate()
  const accessToken = token ?? ''
  const [courses, setCourses] = useState<TeacherCourseItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    async function run() {
      setLoading(true)
      setError('')
      try {
        const data = await listTeacherCourses(accessToken, 1, 50, '', '', 'updated_at', 'desc')
        if (!cancelled) setCourses(data.items)
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : '加载课程失败')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void run()
    return () => {
      cancelled = true
    }
  }, [accessToken])

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <h2>课程知识库管理</h2>
        <p className="subtle">选择课程后管理可解析文档的分块与向量嵌入；也可从「课程列表」直接进入某一门课的知识库。</p>
      </header>
      {error && <div className="alert error">{error}</div>}
      <section className="card stack">
        {loading && <p className="subtle">加载中…</p>}
        {!loading && courses.length === 0 && <p className="subtle">暂无课程，请先在课程列表中创建。</p>}
        <div className="course-list">
          {courses.map((c) => (
            <article key={c.id} className="course-item">
              <div>
                <h4>{c.title}</h4>
                <p className="subtle">{c.description || '暂无描述'}</p>
              </div>
              <div className="action-row">
                <button className="primary" onClick={() => navigate(`/teacher/knowledge/${c.id}`)}>
                  知识库管理
                </button>
              </div>
            </article>
          ))}
        </div>
      </section>
    </div>
  )
}
