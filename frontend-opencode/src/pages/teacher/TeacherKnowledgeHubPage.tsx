import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getCourses } from '@/services/course'
import type { Course } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Database, BookOpen, ArrowRight } from 'lucide-react'

/**
 * 教师知识库入口页面
 * 显示课程列表，点击跳转到知识库资源列表
 */
export function TeacherKnowledgeHubPage() {
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
        if (!cancelled) {
          setCourses(data.items)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : '加载课程列表失败')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadCourses()

    return () => {
      cancelled = true
    }
  }, [accessToken])

  const handleNavigateToKnowledge = (courseId: number) => {
    navigate(`/teacher/knowledge/${courseId}`)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">知识库管理</h1>
        <p className="text-muted-foreground mt-2">
          选择课程后管理可解析文档的分块与向量嵌入
        </p>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {loading && (
        <Card>
          <CardContent className="py-12">
            <div className="flex flex-col items-center justify-center gap-3">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
              <p className="text-muted-foreground">加载中...</p>
            </div>
          </CardContent>
        </Card>
      )}

      {!loading && courses.length === 0 && !error && (
        <Card>
          <CardContent className="py-12">
            <div className="flex flex-col items-center justify-center gap-4">
              <div className="rounded-full bg-muted p-4">
                <Database className="h-8 w-8 text-muted-foreground" />
              </div>
              <div className="text-center">
                <p className="text-muted-foreground mb-2">暂无课程数据</p>
                <p className="text-sm text-muted-foreground">
                  请先在「课程管理」中创建课程
                </p>
              </div>
              <Button
                variant="outline"
                onClick={() => navigate('/teacher/courses')}
              >
                前往课程管理
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {!loading && courses.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {courses.map((course) => (
            <Card
              key={course.id}
              className="group flex flex-col transition-all hover:border-primary/50 hover:shadow-md"
            >
              <CardHeader>
                <div className="flex items-start gap-3">
                  <div className="rounded-lg bg-primary/10 p-2.5">
                    <BookOpen className="h-5 w-5 text-primary" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <CardTitle className="line-clamp-2 text-lg">
                      {course.title}
                    </CardTitle>
                    <CardDescription className="line-clamp-2 mt-1">
                      {course.description || '暂无描述'}
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>

              <CardContent className="flex-1">
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">状态</span>
                    <span
                      className={`font-medium ${
                        course.status === 'published'
                          ? 'text-green-600 dark:text-green-400'
                          : course.status === 'draft'
                            ? 'text-yellow-600 dark:text-yellow-400'
                            : 'text-gray-600 dark:text-gray-400'
                      }`}
                    >
                      {course.status === 'published' && '已发布'}
                      {course.status === 'draft' && '草稿'}
                      {course.status === 'archived' && '已归档'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">更新时间</span>
                    <span className="text-foreground">
                      {new Date(course.updated_at).toLocaleDateString('zh-CN')}
                    </span>
                  </div>
                </div>
              </CardContent>

              <div className="border-t p-4">
                <Button
                  className="w-full group-hover:bg-primary"
                  onClick={() => handleNavigateToKnowledge(course.id)}
                >
                  <span>进入知识库</span>
                  <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}