import { NavLink, Outlet, useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

/** 单课程题库：子路由分为「已入库列表」与「新增 / AI 导题」两页。 */
export function TeacherQuestionBankCourseLayout() {
  const { courseId } = useParams<{ courseId: string }>()
  const navigate = useNavigate()
  const base = `/teacher/question-bank/${courseId ?? ''}`

  const tabCls = ({ isActive }: { isActive: boolean }) =>
    cn(
      'inline-flex items-center rounded-md px-3 py-2 text-sm font-medium transition-colors',
      isActive
        ? 'bg-primary text-primary-foreground'
        : 'text-muted-foreground hover:bg-muted hover:text-foreground',
    )

  return (
    <div className="space-y-6">
      <div>
        <div className="mb-2">
          <Button variant="outline" size="sm" onClick={() => navigate('/teacher/question-bank')}>
            <ArrowLeft className="mr-1 h-4 w-4" />
            返回课程选择
          </Button>
        </div>
        <h1 className="text-3xl font-bold text-foreground">课程题库</h1>
        <p className="text-muted-foreground mt-1 text-sm">课程 ID：{courseId}</p>
      </div>

      <nav className="flex flex-wrap gap-2 border-b border-border pb-3">
        <NavLink to={`${base}/list`} className={tabCls} end>
          已入库题目
        </NavLink>
        <NavLink to={`${base}/maintain`} className={tabCls}>
          新增与 AI 导题
        </NavLink>
      </nav>

      <Outlet />
    </div>
  )
}
