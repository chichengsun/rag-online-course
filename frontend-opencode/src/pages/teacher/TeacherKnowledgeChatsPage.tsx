import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getSessions, createSession, deleteSession } from '@/services/chat'
import { getCourses } from '@/services/course'
import type { ChatSession, Course } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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

export function TeacherKnowledgeChatsPage() {
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const accessToken = token ?? ''

  const [sessions, setSessions] = useState<ChatSession[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const [courses, setCourses] = useState<Course[]>([])
  const [coursesLoading, setCoursesLoading] = useState(false)

  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [selectedCourseId, setSelectedCourseId] = useState<string>('')
  const [newSessionTitle, setNewSessionTitle] = useState('')
  const [creating, setCreating] = useState(false)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [sessionToDelete, setSessionToDelete] = useState<ChatSession | null>(null)
  const [deleting, setDeleting] = useState(false)

  const loadCourses = useCallback(async () => {
    setCoursesLoading(true)
    try {
      const data = await getCourses(accessToken, { page: 1, page_size: 100 })
      setCourses(data.items)
      if (data.items.length > 0 && !selectedCourseId) {
        setSelectedCourseId(String(data.items[0].id))
      }
    } catch (err) {
      console.error('加载课程列表失败:', err)
    } finally {
      setCoursesLoading(false)
    }
  }, [accessToken, selectedCourseId])

  const loadSessions = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await getSessions()
      setSessions(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载会话列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadCourses()
    void loadSessions()
  }, [loadCourses, loadSessions])

  const handleOpenCreateDialog = () => {
    setNewSessionTitle('')
    setCreateDialogOpen(true)
  }

  const handleCreateSession = async () => {
    if (!selectedCourseId) {
      setError('请选择课程')
      return
    }

    setCreating(true)
    setError(null)
    try {
      const session = await createSession(selectedCourseId, {
        title: newSessionTitle.trim() || undefined,
      })
      setSuccessMessage('会话创建成功')
      setCreateDialogOpen(false)
      navigate(`/teacher/knowledge/chats/${session.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建会话失败')
    } finally {
      setCreating(false)
    }
  }

  const handleOpenDeleteDialog = (session: ChatSession) => {
    setSessionToDelete(session)
    setDeleteDialogOpen(true)
  }

  const handleConfirmDelete = async () => {
    if (!sessionToDelete) return

    setDeleting(true)
    setError(null)
    try {
      await deleteSession(String(sessionToDelete.id))
      setSuccessMessage('会话已删除')
      setDeleteDialogOpen(false)
      setSessionToDelete(null)
      await loadSessions()
      
      setTimeout(() => setSuccessMessage(null), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除会话失败')
    } finally {
      setDeleting(false)
    }
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">对话管理</h1>
        <p className="text-muted-foreground mt-2">
          管理知识库对话会话，点击进入详情查看历史并继续问答
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
              <p className="text-sm text-muted-foreground">
                共 {sessions.length} 个会话
              </p>
            </div>
            <Button onClick={handleOpenCreateDialog} disabled={loading || coursesLoading}>
              新建对话
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 会话列表 */}
      {sessions.length === 0 && !loading ? (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                暂无对话会话
              </p>
              <Button onClick={handleOpenCreateDialog} disabled={coursesLoading}>
                创建第一个会话
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {sessions.map((session) => (
            <Card key={session.id} className="flex flex-col">
              <CardHeader>
                <CardTitle className="line-clamp-2">
                  {session.title || '未命名会话'}
                </CardTitle>
                <CardDescription className="line-clamp-1">
                  课程ID: {session.course_id || '未知'}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex-1">
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">创建时间：</span>
                    <span className="text-foreground">
                      {session.created_at ? formatDate(session.created_at) : '-'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">更新时间：</span>
                    <span className="text-foreground">
                      {session.updated_at ? formatDate(session.updated_at) : '-'}
                    </span>
                  </div>
                </div>
              </CardContent>
              <div className="border-t p-4 flex flex-wrap gap-2">
                <Button
                  variant="default"
                  size="sm"
                  onClick={() => navigate(`/teacher/knowledge/chats/${session.id}`)}
                >
                  进入对话
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => handleOpenDeleteDialog(session)}
                >
                  删除
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>新建对话</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label htmlFor="course-select" className="text-sm font-medium text-foreground">
                选择课程 <span className="text-destructive">*</span>
              </label>
              <Select
                value={selectedCourseId}
                onValueChange={(value) => value && setSelectedCourseId(value)}
                disabled={creating || coursesLoading}
              >
                <SelectTrigger id="course-select">
                  <SelectValue placeholder="选择课程" />
                </SelectTrigger>
                <SelectContent>
                  {courses.map((course) => (
                    <SelectItem key={course.id} value={String(course.id)}>
                      {course.title}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {courses.length === 0 && !coursesLoading && (
                <p className="text-xs text-muted-foreground">
                  暂无课程，请先创建课程
                </p>
              )}
            </div>
            <div className="space-y-2">
              <label htmlFor="session-title" className="text-sm font-medium text-foreground">
                会话标题
              </label>
              <Input
                id="session-title"
                placeholder="输入会话标题（可选，默认为「新对话」）"
                value={newSessionTitle}
                onChange={(e) => setNewSessionTitle(e.target.value)}
                disabled={creating}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateDialogOpen(false)}
              disabled={creating}
            >
              取消
            </Button>
            <Button
              onClick={handleCreateSession}
              disabled={creating || !selectedCourseId || courses.length === 0}
            >
              {creating ? '创建中...' : '创建'}
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
              确定要删除会话「{sessionToDelete?.title || '未命名会话'}」吗？此操作不可恢复。
            </p>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={deleting}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleting}
            >
              {deleting ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}