import { useEffect, useState, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getPublishedCourses, enrollCourse, type MyCourseItem } from '@/services/student'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

export function StudentCoursesPage() {
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const accessToken = token ?? ''

  const [courses, setCourses] = useState<MyCourseItem[]>([])
  const [filteredCourses, setFilteredCourses] = useState<MyCourseItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const [keyword, setKeyword] = useState('')
  const skipKeywordDebounceRef = useRef(true)

  const [enrollDialogOpen, setEnrollDialogOpen] = useState(false)
  const [courseToEnroll, setCourseToEnroll] = useState<MyCourseItem | null>(null)
  const [enrolling, setEnrolling] = useState(false)

  const loadCourses = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await getPublishedCourses(accessToken)
      setCourses(data)
      setFilteredCourses(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载课程列表失败')
    } finally {
      setLoading(false)
    }
  }, [accessToken])

  useEffect(() => {
    if (accessToken) {
      void loadCourses()
    }
  }, [accessToken, loadCourses])

  useEffect(() => {
    if (skipKeywordDebounceRef.current) {
      skipKeywordDebounceRef.current = false
      return
    }
    const timer = window.setTimeout(() => {
      const filtered = courses.filter(
        (course) =>
          course.title.toLowerCase().includes(keyword.toLowerCase()) ||
          (course.description && course.description.toLowerCase().includes(keyword.toLowerCase())),
      )
      setFilteredCourses(filtered)
    }, 300)
    return () => window.clearTimeout(timer)
  }, [keyword, courses])

  const handleSearch = () => {
    const filtered = courses.filter(
      (course) =>
        course.title.toLowerCase().includes(keyword.toLowerCase()) ||
        (course.description && course.description.toLowerCase().includes(keyword.toLowerCase())),
    )
    setFilteredCourses(filtered)
  }

  const handleOpenEnrollDialog = (course: MyCourseItem) => {
    setCourseToEnroll(course)
    setEnrollDialogOpen(true)
  }

  const handleConfirmEnroll = async () => {
    if (!courseToEnroll) return

    setEnrolling(true)
    setError(null)
    try {
      await enrollCourse(accessToken, parseInt(courseToEnroll.id, 10))
      setSuccessMessage(`已成功选课「${courseToEnroll.title}」`)
      setEnrollDialogOpen(false)
      setCourseToEnroll(null)
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '选课失败，请稍后重试')
    } finally {
      setEnrolling(false)
    }
  }

  const handleViewCatalog = (courseId: string) => {
    navigate(`/student/courses/${courseId}`)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">课程列表</h1>
        <p className="text-muted-foreground mt-2">
          浏览已发布的课程，选择您感兴趣的课程开始学习
        </p>
      </div>

      {successMessage && (
        <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
          <p className="text-sm text-green-600 dark:text-green-400">{successMessage}</p>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-lg bg-destructive/10 border border-destructive/20">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex-1 max-w-md">
              <Input
                placeholder="搜索课程标题或描述..."
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    handleSearch()
                  }
                }}
                disabled={loading}
              />
            </div>
            <Button onClick={handleSearch} disabled={loading}>
              搜索
            </Button>
          </div>
        </CardContent>
      </Card>

      {filteredCourses.length === 0 && !loading ? (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                {keyword ? '没有找到匹配的课程' : '暂无可选课程'}
              </p>
              {keyword && (
                <Button variant="outline" onClick={() => setKeyword('')}>
                  清除搜索
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filteredCourses.map((course) => (
            <Card key={course.id} className="flex flex-col">
              <CardHeader>
                <div className="flex items-start justify-between gap-2">
                  <CardTitle className="line-clamp-2">{course.title}</CardTitle>
                  <Badge variant="success">已发布</Badge>
                </div>
                <CardDescription className="line-clamp-3">
                  {course.description || '暂无描述'}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1">
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">课程ID：</span>
                    <span className="text-foreground font-mono">{course.id}</span>
                  </div>
                </div>
              </CardContent>
              <div className="border-t p-4 flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleViewCatalog(course.id)}
                >
                  查看目录
                </Button>
                <Button
                  size="sm"
                  onClick={() => handleOpenEnrollDialog(course)}
                >
                  选课
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={enrollDialogOpen} onOpenChange={setEnrollDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>确认选课</DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">
              确定要选择课程「{courseToEnroll?.title}」吗？
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setEnrollDialogOpen(false)}
              disabled={enrolling}
            >
              取消
            </Button>
            <Button
              onClick={handleConfirmEnroll}
              disabled={enrolling}
            >
              {enrolling ? '选课中...' : '确认选课'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}