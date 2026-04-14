import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getCourses } from '@/services/course'
import { useAuthStore } from '@/stores/authStore'
import type { Course } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { FileText, ArrowRight } from 'lucide-react'

// 题库入口页：按课程维度进入题库管理。
export function TeacherQuestionBankHubPage() {
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const accessToken = token ?? ''
  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function loadCourses() {
      setLoading(true)
      setError(null)
      try {
        const data = await getCourses(accessToken, {
          page: 1,
          page_size: 100,
          sort_by: 'updated_at',
          sort_order: 'desc',
        })
        if (!cancelled) setCourses(data.items)
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : '加载课程失败')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void loadCourses()
    return () => {
      cancelled = true
    }
  }, [accessToken])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">课程题库</h1>
        <p className="text-muted-foreground mt-2">选择课程后进入题库管理，支持手工编辑和上传文本 AI 导题。</p>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {loading ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">加载中...</CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {courses.map((course) => (
            <Card key={course.id} className="group flex flex-col">
              <CardHeader>
                <div className="flex items-start gap-3">
                  <div className="rounded-lg bg-primary/10 p-2.5">
                    <FileText className="h-5 w-5 text-primary" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <CardTitle className="line-clamp-2 text-lg">{course.title}</CardTitle>
                    <CardDescription className="line-clamp-2 mt-1">{course.description || '暂无描述'}</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="mt-auto border-t pt-4">
                <Button className="w-full" onClick={() => navigate(`/teacher/question-bank/${course.id}/list`)}>
                  进入题库
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
