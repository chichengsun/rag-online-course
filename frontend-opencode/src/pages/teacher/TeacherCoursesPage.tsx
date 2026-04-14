import { useEffect, useState, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getCourses, createCourse, updateCourse, deleteCourse } from '@/services/course'
import type { Course, CourseStatus } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

export function TeacherCoursesPage() {
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const accessToken = token ?? ''

  const [courses, setCourses] = useState<Course[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const [page, setPage] = useState(1)
  const [pageSize] = useState(10)
  const [total, setTotal] = useState(0)

  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<CourseStatus | ''>('')
  const [sortBy, setSortBy] = useState<'created_at' | 'updated_at' | 'title'>('created_at')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')

  const skipKeywordDebounceRef = useRef(true)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [editingCourse, setEditingCourse] = useState<Course | null>(null)
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    status: 'draft' as CourseStatus,
  })
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [courseToDelete, setCourseToDelete] = useState<Course | null>(null)

  const loadCourses = useCallback(async (
    targetPage = page,
    targetKeyword = keyword,
    targetStatus = statusFilter,
    targetSortBy = sortBy,
    targetSortOrder = sortOrder,
  ) => {
    setLoading(true)
    setError(null)
    try {
      const data = await getCourses(accessToken, {
        page: targetPage,
        page_size: pageSize,
        keyword: targetKeyword || undefined,
        status: targetStatus || undefined,
        sort_by: targetSortBy,
        sort_order: targetSortOrder,
      })
      setCourses(data.items)
      setTotal(data.total)
      setPage(data.page)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载课程列表失败')
    } finally {
      setLoading(false)
    }
  }, [accessToken, pageSize, page, keyword, statusFilter, sortBy, sortOrder])

  useEffect(() => {
    void loadCourses(1)
  }, [loadCourses])

  useEffect(() => {
    if (skipKeywordDebounceRef.current) {
      skipKeywordDebounceRef.current = false
      return
    }
    const timer = window.setTimeout(() => {
      void loadCourses(1, keyword, statusFilter, sortBy, sortOrder)
    }, 300)
    return () => window.clearTimeout(timer)
  }, [keyword, statusFilter, sortBy, sortOrder, loadCourses])

  const handleSearch = () => {
    void loadCourses(1, keyword, statusFilter, sortBy, sortOrder)
  }

  const handleOpenCreateDialog = () => {
    setDialogMode('create')
    setEditingCourse(null)
    setFormData({
      title: '',
      description: '',
      status: 'draft',
    })
    setFormErrors({})
    setDialogOpen(true)
  }

  const handleOpenEditDialog = (course: Course) => {
    setDialogMode('edit')
    setEditingCourse(course)
    setFormData({
      title: course.title,
      description: course.description,
      status: course.status,
    })
    setFormErrors({})
    setDialogOpen(true)
  }

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {}
    
    if (!formData.title.trim()) {
      errors.title = '课程标题不能为空'
    }
    
    setFormErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmitForm = async () => {
    if (!validateForm()) {
      return
    }

    setLoading(true)
    setError(null)
    try {
      if (dialogMode === 'create') {
        await createCourse(accessToken, {
          title: formData.title,
          description: formData.description || undefined,
        })
        setSuccessMessage('课程创建成功')
      } else if (editingCourse) {
        await updateCourse(accessToken, editingCourse.id, {
          title: formData.title,
          description: formData.description || undefined,
          status: formData.status,
        })
        setSuccessMessage('课程更新成功')
      }
      
      setDialogOpen(false)
      await loadCourses(page, keyword, statusFilter, sortBy, sortOrder)
      
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败')
    } finally {
      setLoading(false)
    }
  }

  const handleOpenDeleteDialog = (course: Course) => {
    setCourseToDelete(course)
    setDeleteDialogOpen(true)
  }

  const handleConfirmDelete = async () => {
    if (!courseToDelete) return

    setLoading(true)
    setError(null)
    try {
      await deleteCourse(accessToken, courseToDelete.id)
      setSuccessMessage('课程已删除')
      setDeleteDialogOpen(false)
      setCourseToDelete(null)
      await loadCourses(page, keyword, statusFilter, sortBy, sortOrder)
      
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除课程失败')
    } finally {
      setLoading(false)
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">课程管理</h1>
        <p className="text-muted-foreground mt-2">
          管理您的课程内容、章节和资源
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
            <div className="flex flex-1 flex-col gap-3 sm:flex-row sm:items-center">
              <div className="flex-1 max-w-md">
                <Input
                  placeholder="搜索课程标题..."
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

              <Select
                value={statusFilter}
                onValueChange={(value) => {
                  setStatusFilter(value as CourseStatus | '')
                  void loadCourses(1, keyword, value as CourseStatus | '', sortBy, sortOrder)
                }}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="全部状态" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">全部状态</SelectItem>
                  <SelectItem value="draft">草稿</SelectItem>
                  <SelectItem value="published">已发布</SelectItem>
                  <SelectItem value="archived">已归档</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={sortBy}
                onValueChange={(value) => {
                  const newSortBy = value as 'created_at' | 'updated_at' | 'title'
                  setSortBy(newSortBy)
                  void loadCourses(1, keyword, statusFilter, newSortBy, sortOrder)
                }}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="created_at">创建时间</SelectItem>
                  <SelectItem value="updated_at">更新时间</SelectItem>
                  <SelectItem value="title">标题</SelectItem>
                </SelectContent>
              </Select>

              <Select
                value={sortOrder}
                onValueChange={(value) => {
                  const newSortOrder = value as 'asc' | 'desc'
                  setSortOrder(newSortOrder)
                  void loadCourses(1, keyword, statusFilter, sortBy, newSortOrder)
                }}
              >
                <SelectTrigger className="w-[120px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="desc">降序</SelectItem>
                  <SelectItem value="asc">升序</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <Button onClick={handleOpenCreateDialog} disabled={loading}>
              新建课程
            </Button>
          </div>
        </CardContent>
      </Card>

      {courses.length === 0 && !loading ? (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                暂无课程数据
              </p>
              <Button onClick={handleOpenCreateDialog}>
                创建第一个课程
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
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">状态：</span>
                    <span className={`font-medium ${
                      course.status === 'published' 
                        ? 'text-green-600 dark:text-green-400' 
                        : course.status === 'draft'
                        ? 'text-yellow-600 dark:text-yellow-400'
                        : 'text-gray-600 dark:text-gray-400'
                    }`}>
                      {course.status === 'published' && '已发布'}
                      {course.status === 'draft' && '草稿'}
                      {course.status === 'archived' && '已归档'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">创建时间：</span>
                    <span className="text-foreground">
                      {new Date(course.created_at).toLocaleDateString('zh-CN')}
                    </span>
                  </div>
                </div>
              </CardContent>
              <div className="border-t p-4 flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate(`/teacher/knowledge/${course.id}`)}
                >
                  知识库管理
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate(`/teacher/course-content/${course.id}`)}
                >
                  章节管理
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleOpenEditDialog(course)}
                >
                  编辑
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => handleOpenDeleteDialog(course)}
                >
                  删除
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}

      {total > 0 && (
        <Card>
          <CardContent className="py-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                共 {total} 条记录，第 {page} / {totalPages} 页
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1 || loading}
                  onClick={() => void loadCourses(page - 1, keyword, statusFilter, sortBy, sortOrder)}
                >
                  上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= totalPages || loading}
                  onClick={() => void loadCourses(page + 1, keyword, statusFilter, sortBy, sortOrder)}
                >
                  下一页
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'create' ? '新建课程' : '编辑课程'}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label htmlFor="course-title" className="text-sm font-medium text-foreground">
                课程标题 <span className="text-destructive">*</span>
              </label>
              <Input
                id="course-title"
                placeholder="输入课程标题"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                disabled={loading}
                className={formErrors.title ? 'border-destructive' : ''}
              />
              {formErrors.title && (
                <p className="text-xs text-destructive">{formErrors.title}</p>
              )}
            </div>
            <div className="space-y-2">
              <label htmlFor="course-description" className="text-sm font-medium text-foreground">
                课程描述
              </label>
              <Textarea
                id="course-description"
                placeholder="输入课程描述（可选）"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                disabled={loading}
                rows={4}
              />
            </div>
            {dialogMode === 'edit' && (
              <div className="space-y-2">
                <label htmlFor="course-status" className="text-sm font-medium text-foreground">
                  课程状态
                </label>
                <Select
                  value={formData.status}
                  onValueChange={(value) => setFormData({ ...formData, status: value as CourseStatus })}
                  disabled={loading}
                >
                  <SelectTrigger id="course-status">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">草稿</SelectItem>
                    <SelectItem value="published">已发布</SelectItem>
                    <SelectItem value="archived">已归档</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              onClick={handleSubmitForm}
              disabled={loading || !formData.title.trim()}
            >
              {loading ? '提交中...' : dialogMode === 'create' ? '创建' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">
              确定要删除课程「{courseToDelete?.title}」吗？此操作不可恢复。
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={loading}
            >
              {loading ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}