import { useEffect, useState, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import {
  getResource,
  getPreviewUrl,
  updateProgress,
  completeResource,
  getResourceProgress,
} from '@/services/student'
import type { Resource, ResourceProgressResp } from '@/types'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { ArrowLeft, CheckCircle, Clock, FileText, Play } from 'lucide-react'
import { toast } from 'sonner'

const resourceTypeLabels: Record<string, string> = {
  pdf: 'PDF 文档',
  doc: 'Word 文档',
  docx: 'Word 文档',
  ppt: 'PPT 演示',
  video: '视频',
  audio: '音频',
  txt: '文本文档',
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${minutes}:${secs.toString().padStart(2, '0')}`
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

export function StudentProgressPage() {
  const { resourceId } = useParams<{ resourceId: string }>()
  const navigate = useNavigate()
  const { token } = useAuthStore()
  const accessToken = token ?? ''

  const [resource, setResource] = useState<Resource | null>(null)
  const [progress, setProgress] = useState<ResourceProgressResp | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [completing, setCompleting] = useState(false)

  const videoRef = useRef<HTMLVideoElement>(null)
  const audioRef = useRef<HTMLAudioElement>(null)
  const progressUpdateTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastUpdateTimeRef = useRef<number>(0)

  const loadResourceData = useCallback(async () => {
    if (!resourceId) return
    const resourceIdNum = parseInt(resourceId, 10)
    if (isNaN(resourceIdNum)) return

    setLoading(true)
    try {
      const [resourceData, progressData, urlData] = await Promise.all([
        getResource(accessToken, resourceIdNum),
        getResourceProgress(accessToken, resourceIdNum),
        getPreviewUrl(accessToken, resourceIdNum),
      ])

      setResource(resourceData)
      setProgress(progressData)
      setPreviewUrl(urlData)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '加载资源失败')
      navigate('/student/courses')
    } finally {
      setLoading(false)
    }
  }, [accessToken, resourceId, navigate])

  useEffect(() => {
    void loadResourceData()
  }, [loadResourceData])

  const handleProgressUpdate = useCallback(
    async (watchedSeconds: number, progressPercent: number) => {
      if (!resourceId) return
      const resourceIdNum = parseInt(resourceId, 10)
      if (isNaN(resourceIdNum)) return

      try {
        const result = await updateProgress(accessToken, resourceIdNum, {
          resource_id: resourceIdNum,
          watched_seconds: watchedSeconds,
          progress_percent: progressPercent,
          is_completed: false,
        })
        setProgress(result)
      } catch (err) {
        console.error('更新进度失败:', err)
      }
    },
    [accessToken, resourceId],
  )

  const handleTimeUpdate = useCallback(() => {
    const mediaElement = videoRef.current || audioRef.current
    if (!mediaElement || !resource) return

    const currentTime = Math.floor(mediaElement.currentTime)
    const duration = resource.duration_seconds || Math.floor(mediaElement.duration) || 0

    if (duration === 0) return

    const progressPercent = Math.min(100, Math.round((currentTime / duration) * 100))

    const now = Date.now()
    if (now - lastUpdateTimeRef.current >= 5000) {
      lastUpdateTimeRef.current = now
      void handleProgressUpdate(currentTime, progressPercent)
    }
  }, [resource, handleProgressUpdate])

  const handleMediaEnd = useCallback(() => {
    const mediaElement = videoRef.current || audioRef.current
    if (!mediaElement || !resource) return

    const duration = resource.duration_seconds || Math.floor(mediaElement.duration) || 0
    void handleProgressUpdate(duration, 100)
  }, [resource, handleProgressUpdate])

  const handleMarkComplete = async () => {
    if (!resourceId) return
    const resourceIdNum = parseInt(resourceId, 10)
    if (isNaN(resourceIdNum)) return

    setCompleting(true)
    try {
      const result = await completeResource(accessToken, resourceIdNum)
      setProgress(result)
      toast.success('已标记为完成')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '标记完成失败')
    } finally {
      setCompleting(false)
    }
  }

  useEffect(() => {
    return () => {
      if (progressUpdateTimerRef.current) {
        clearTimeout(progressUpdateTimerRef.current)
      }
    }
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    )
  }

  if (!resource) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>资源不存在</CardTitle>
            <CardDescription>该资源可能已被删除或您无权访问。</CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate('/student/courses')}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              返回课程列表
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const isVideo = resource.resource_type === 'video'
  const isAudio = resource.resource_type === 'audio'
  const isPdf = resource.resource_type === 'pdf'
  const isMedia = isVideo || isAudio

  const currentProgress = progress?.progress_percent ?? 0
  const isCompleted = progress?.is_completed ?? false
  const watchedSeconds = progress?.updated_at ? resource.duration_seconds : 0

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate(-1)}
        >
          <ArrowLeft className="w-4 h-4 mr-1" />
          返回
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-foreground">{resource.title}</h1>
          <div className="flex items-center gap-4 mt-1 text-sm text-muted-foreground">
            <span className="flex items-center gap-1">
              <FileText className="w-4 h-4" />
              {resourceTypeLabels[resource.resource_type] || resource.resource_type}
            </span>
            <span>{formatFileSize(resource.size_bytes)}</span>
            {resource.duration_seconds > 0 && (
              <span className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                {formatDuration(resource.duration_seconds)}
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <Card>
            <CardContent className="p-0">
              {isMedia && previewUrl && (
                <div className="relative bg-black rounded-lg overflow-hidden">
                  {isVideo && (
                    <video
                      ref={videoRef}
                      src={previewUrl}
                      controls
                      className="w-full aspect-video"
                      onTimeUpdate={handleTimeUpdate}
                      onEnded={handleMediaEnd}
                    />
                  )}
                  {isAudio && (
                    <div className="p-8">
                      <div className="flex items-center justify-center mb-6">
                        <div className="w-32 h-32 bg-primary/10 rounded-full flex items-center justify-center">
                          <Play className="w-12 h-12 text-primary" />
                        </div>
                      </div>
                      <audio
                        ref={audioRef}
                        src={previewUrl}
                        controls
                        className="w-full"
                        onTimeUpdate={handleTimeUpdate}
                        onEnded={handleMediaEnd}
                      />
                    </div>
                  )}
                </div>
              )}

              {isPdf && previewUrl && (
                <iframe
                  src={previewUrl}
                  className="w-full h-[600px] border-0"
                  title={resource.title}
                />
              )}

              {!isMedia && !isPdf && previewUrl && (
                <div className="p-8 text-center">
                  <FileText className="w-16 h-16 mx-auto text-muted-foreground mb-4" />
                  <p className="text-muted-foreground mb-4">
                    此资源类型不支持在线预览
                  </p>
                  <a
                    href={previewUrl}
                    download={resource.title}
                    className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground px-2.5 py-1.5 text-sm font-medium hover:bg-primary/80"
                  >
                    下载资源
                  </a>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">学习进度</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>完成度</span>
                  <span className="font-medium">{currentProgress}%</span>
                </div>
                <Progress value={currentProgress} className="h-3" />
              </div>

              {resource.duration_seconds > 0 && (
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">已观看</span>
                  <span>{formatDuration(Math.floor(watchedSeconds))} / {formatDuration(resource.duration_seconds)}</span>
                </div>
              )}

              <div className="pt-4">
                {isCompleted ? (
                  <div className="flex items-center justify-center gap-2 py-3 bg-green-50 dark:bg-green-950/30 rounded-lg text-green-700 dark:text-green-400">
                    <CheckCircle className="w-5 h-5" />
                    <span className="font-medium">已完成学习</span>
                  </div>
                ) : (
                  <Button
                    className="w-full"
                    onClick={handleMarkComplete}
                    disabled={completing}
                  >
                    {completing ? '处理中...' : '标记为完成'}
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">资源信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">类型</span>
                <span>{resourceTypeLabels[resource.resource_type] || resource.resource_type}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">大小</span>
                <span>{formatFileSize(resource.size_bytes)}</span>
              </div>
              {resource.duration_seconds > 0 && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">时长</span>
                  <span>{formatDuration(resource.duration_seconds)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">上传时间</span>
                <span>{new Date(resource.created_at).toLocaleDateString('zh-CN')}</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}