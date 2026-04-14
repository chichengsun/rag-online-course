import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getMyCourses, type MyCourseItem } from '@/services/student'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'

export function StudentMyCoursesPage() {
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const accessToken = token ?? ''

  const [courses, setCourses] = useState<MyCourseItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadCourses = async () => {
      setLoading(true)
      setError(null)
      try {
        const data = await getMyCourses(accessToken)
        setCourses(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : '加载课程列表失败')
      } finally {
        setLoading(false)
      }
    }

    if (accessToken) {
      void loadCourses()
    }
  }, [accessToken])

  const handleContinueLearning = (courseId: string) => {
    navigate(`/student/courses/${courseId}`)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">我的课程</h1>
        <p className="text-muted-foreground mt-2">
          查看您已选的课程，继续学习之旅
        </p>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <p className="text-muted-foreground">加载中...</p>
        </div>
      ) : courses.length === 0 ? (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                您还没有选择任何课程
              </p>
              <Button onClick={() => navigate('/student/courses')}>
                浏览课程
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {courses.map((course) => (
            <Card key={course.id} className="flex flex-col">
              <CardHeader>
                <CardTitle className="line-clamp-2">{course.title}</CardTitle>
                <CardDescription className="line-clamp-3">
                  {course.description || '暂无描述'}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1">
                <div className="space-y-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">学习进度</span>
                      <span className="text-foreground font-medium">0%</span>
                    </div>
                    <Progress value={0} />
                  </div>
                  <div className="text-sm text-muted-foreground">
                    选课时间：{new Date(course.enrolled_at).toLocaleDateString('zh-CN')}
                  </div>
                </div>
              </CardContent>
              <div className="border-t p-4">
                <Button
                  className="w-full"
                  onClick={() => handleContinueLearning(course.id)}
                >
                  继续学习
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}