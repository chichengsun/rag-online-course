import { Navigate, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'

type TeacherNavItem = {
  key: string
  title: string
  desc: string
  to?: string
}

// TeacherLayout 提供教师端全局框架：左侧大类导航，右侧业务工作区。
export function TeacherLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const navItems: TeacherNavItem[] = [
    { key: 'course-list', title: '课程列表管理', desc: '搜索、分页、创建课程', to: '/teacher/courses' },
    { key: 'course-content', title: '章节与资源管理', desc: '章节目录与资源上传维护', to: '/teacher/course-content' },
    { key: 'knowledge', title: '课程知识库管理', desc: '分块、嵌入与可解析资源', to: '/teacher/knowledge' },
    { key: 'chat-management', title: '对话管理', desc: '分页查看会话并进入详情问答', to: '/teacher/knowledge/chats' },
    { key: 'ai-models', title: '模型管理', desc: '问答、嵌入、重排模型配置', to: '/teacher/ai-models' },
    { key: 'dashboard', title: '教学数据', desc: '建设中', to: undefined },
  ]

  if (!user || user.role !== 'teacher') {
    return <Navigate to="/auth" replace />
  }

  return (
    <main className="teacher-app-shell">
      <aside className="teacher-app-sidebar">
        <div className="brand-block">
          <h1>Teacher Console</h1>
          <p>面向教师的课程运营工作台</p>
        </div>

        <nav className="category-nav">
          {navItems.map((item) => {
            const active = item.to
              ? location.pathname === item.to || location.pathname.startsWith(`${item.to}/`)
              : false
            return (
              <button
                key={item.key}
                className={active ? 'category-item active' : 'category-item'}
                disabled={!item.to}
                onClick={() => item.to && navigate(item.to)}
              >
                <span>{item.title}</span>
                <small>{item.desc}</small>
              </button>
            )
          })}
        </nav>

        <div className="sidebar-footer">
          <p className="subtle">
            当前用户：{user.username}（{user.role}）
          </p>
          <button className="danger" onClick={logout}>
            退出登录
          </button>
        </div>
      </aside>

      <section className="teacher-app-main">
        <Outlet />
      </section>
    </main>
  )
}
