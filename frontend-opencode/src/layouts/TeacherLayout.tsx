import { Navigate, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { BookOpen, Database, LayoutTemplate, LogOut, Menu, MessageSquare, Cpu, X, FileText } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

/** 判断当前路径是否属于「RAG 对话」域（会话与 AI 模型），与「课程管理」域互斥。 */
function isRagDialogDomain(pathname: string): boolean {
  if (pathname.startsWith('/teacher/ai-models')) return true
  if (pathname === '/teacher/knowledge/chats' || pathname.startsWith('/teacher/knowledge/chats/')) {
    return true
  }
  return false
}

type SubNavItem = {
  key: string
  title: string
  desc: string
  to: string
  icon: React.ComponentType<{ className?: string }>
  /** 当前路由是否视为该子项激活 */
  isActive: (pathname: string) => boolean
}

/** 课程管理域下子功能：课程列表与建设、知识库（分块/嵌入）。 */
const courseSubNav: SubNavItem[] = [
  {
    key: 'courses',
    title: '课程与内容',
    desc: '课程列表、章节与资源',
    to: '/teacher/courses',
    icon: BookOpen,
    isActive: (p) =>
      p === '/teacher/courses' || p.startsWith('/teacher/course-content') || p.startsWith('/teacher/resources/preview'),
  },
  {
    key: 'question-bank',
    title: '课程题库',
    desc: '手工维护与 AI 导题',
    to: '/teacher/question-bank',
    icon: FileText,
    isActive: (p) => p === '/teacher/question-bank' || /^\/teacher\/question-bank\/\d+/.test(p),
  },
  {
    key: 'knowledge',
    title: '知识库',
    desc: '分块、确认与向量嵌入',
    to: '/teacher/knowledge',
    icon: Database,
    isActive: (p) => {
      if (p === '/teacher/knowledge') return true
      if (p.startsWith('/teacher/knowledge/chats')) return false
      return /^\/teacher\/knowledge\/\d/.test(p)
    },
  },
  {
    key: 'course-design',
    title: '课程设计',
    desc: 'AI 生成教学大纲草案',
    to: '/teacher/course-design',
    icon: LayoutTemplate,
    isActive: (p) =>
      p === '/teacher/course-design' || /^\/teacher\/course-design\/\d+/.test(p),
  },
]

/** RAG 对话域下子功能：会话列表与模型配置。 */
const ragSubNav: SubNavItem[] = [
  {
    key: 'chats',
    title: '对话会话',
    desc: '按课程创建与继续问答',
    to: '/teacher/knowledge/chats',
    icon: MessageSquare,
    isActive: (p) => p === '/teacher/knowledge/chats' || p.startsWith('/teacher/knowledge/chats/'),
  },
  {
    key: 'ai-models',
    title: 'AI 模型',
    desc: '问答、嵌入、重排模型',
    to: '/teacher/ai-models',
    icon: Cpu,
    isActive: (p) => p === '/teacher/ai-models' || p.startsWith('/teacher/ai-models/'),
  },
]

/**
 * TeacherLayout 教师端壳层：顶栏居中切换「课程管理 / RAG 对话」两大模块，左侧仅展示当前模块下的子功能入口。
 */
export function TeacherLayout() {
  const { user, isAuthenticated } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const ragDomain = isRagDialogDomain(location.pathname)
  const subNav = ragDomain ? ragSubNav : courseSubNav

  /** 需要占满主内容区高度、由页面内部控制滚动的路由（避免整页出现第二条滚动条）。 */
  const outletFillHeight =
    /^\/teacher\/knowledge\/chats\/[^/]+$/.test(location.pathname) ||
    location.pathname.startsWith('/teacher/resources/preview')

  const headerSubtitle = useMemo(() => {
    const hit = subNav.find((item) => item.isActive(location.pathname))
    return hit?.title ?? (ragDomain ? 'RAG 对话' : '课程管理')
  }, [location.pathname, ragDomain, subNav])

  if (!isAuthenticated() || user?.role !== 'teacher') {
    return <Navigate to="/auth" replace />
  }

  const handleNavClick = (path: string) => {
    navigate(path)
    setSidebarOpen(false)
  }

  return (
    <div className="flex h-screen bg-background">
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
          aria-hidden
        />
      )}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 transform border-r border-border bg-card transition-transform duration-300 ease-in-out lg:static lg:translate-x-0',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <div className="flex h-16 items-center justify-between border-b border-border px-6">
          <div>
            <h1 className="text-lg font-semibold text-foreground">教师工作台</h1>
            <p className="text-xs text-muted-foreground">{ragDomain ? 'RAG 对话' : '课程管理'}</p>
          </div>
          <Button variant="ghost" size="icon" className="lg:hidden" onClick={() => setSidebarOpen(false)}>
            <X className="size-5" />
          </Button>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto p-4">
          {subNav.map((item) => {
            const Icon = item.icon
            const active = item.isActive(location.pathname)

            return (
              <button
                key={item.key}
                type="button"
                onClick={() => handleNavClick(item.to)}
                className={cn(
                  'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors',
                  active
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                )}
              >
                <Icon className="size-5 shrink-0" />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium">{item.title}</div>
                  <div
                    className={cn(
                      'truncate text-xs',
                      active ? 'text-primary-foreground/80' : 'text-muted-foreground',
                    )}
                  >
                    {item.desc}
                  </div>
                </div>
              </button>
            )
          })}
        </nav>

        <div className="border-t border-border p-4">
          <div className="mb-3 rounded-lg bg-muted/50 px-3 py-2">
            <div className="text-sm font-medium text-foreground">{user?.username}</div>
            <div className="text-xs text-muted-foreground">{user?.email}</div>
          </div>
          <Button
            variant="ghost"
            className="w-full justify-start gap-2 text-muted-foreground hover:text-destructive"
            onClick={() => {
              useAuthStore.getState().logout()
              navigate('/auth')
            }}
          >
            <LogOut className="size-4" />
            退出登录
          </Button>
        </div>
      </aside>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <header className="flex h-16 items-center gap-3 border-b border-border bg-card px-4 lg:px-6">
          <Button variant="ghost" size="icon" className="shrink-0 lg:hidden" onClick={() => setSidebarOpen(true)}>
            <Menu className="size-5" />
          </Button>

          <div className="flex min-w-0 flex-1 justify-center">
            <div
              className="inline-flex rounded-lg border border-border bg-muted/30 p-0.5"
              role="tablist"
              aria-label="教师端主模块"
            >
              <button
                type="button"
                role="tab"
                aria-selected={!ragDomain}
                onClick={() => handleNavClick('/teacher/courses')}
                className={cn(
                  'rounded-md px-4 py-1.5 text-sm font-medium transition-colors',
                  !ragDomain ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
                )}
              >
                课程管理
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={ragDomain}
                onClick={() => handleNavClick('/teacher/knowledge/chats')}
                className={cn(
                  'rounded-md px-4 py-1.5 text-sm font-medium transition-colors',
                  ragDomain ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
                )}
              >
                RAG 对话
              </button>
            </div>
          </div>

          <div className="hidden shrink-0 text-right sm:block lg:min-w-[140px]">
            <div className="text-sm font-medium text-foreground">{user?.username}</div>
            <div className="text-xs text-muted-foreground">{headerSubtitle}</div>
          </div>
        </header>

        {/* min-h-0 使子页可用 h-full/flex 占满剩余高度；非 fill 路由保留外层纵向滚动 */}
        <main className="flex min-h-0 flex-1 flex-col overflow-hidden bg-background p-4 lg:p-6">
          <div
            className={cn(
              'flex min-h-0 flex-1 flex-col',
              outletFillHeight ? 'overflow-hidden' : 'overflow-y-auto',
            )}
          >
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
