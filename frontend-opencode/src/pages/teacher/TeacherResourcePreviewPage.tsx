import { useEffect, useState, useRef, useCallback } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getResource, getPreviewUrl } from '@/services/resource'
import type { Resource } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

/**
 * 获取资源类型显示名称
 */
function getResourceTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    pdf: 'PDF',
    video: '视频',
    audio: '音频',
    doc: 'Word',
    docx: 'Word',
    ppt: 'PPT',
    txt: '文本',
    md: 'Markdown',
  }
  return labels[type] || type.toUpperCase()
}

/**
 * 获取资源类型图标
 */
function getResourceTypeIcon(type: string): string {
  const icons: Record<string, string> = {
    pdf: '📄',
    video: '🎬',
    audio: '🎵',
    doc: '📝',
    docx: '📝',
    ppt: '📊',
    txt: '📃',
    md: '📃',
  }
  return icons[type] || '📁'
}

/**
 * TeacherResourcePreviewPage：教师端资源内嵌预览；主内容区纵向占满，预览区 flex 吃满剩余高度，减少底部留白。
 */
export function TeacherResourcePreviewPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { token } = useAuthStore()

  const resourceId = searchParams.get('resourceId')

  const [resource, setResource] = useState<Resource | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)

  const containerRef = useRef<HTMLDivElement>(null)
  const videoRef = useRef<HTMLVideoElement>(null)

  /**
   * 加载资源详情和预览 URL
   */
  const loadResource = useCallback(async () => {
    if (!resourceId || !token) {
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)

    try {
      const resourceData = await getResource(token, Number(resourceId))
      setResource(resourceData)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载资源失败')
    } finally {
      setLoading(false)
    }
  }, [resourceId, token])

  useEffect(() => {
    void loadResource()
  }, [loadResource])

  const loadPreviewUrl = useCallback(async () => {
    if (!resourceId || !token || !resource) return
    setPreviewLoading(true)
    setPreviewError(null)
    try {
      const url = await getPreviewUrl(token, Number(resourceId))
      setPreviewUrl(url)
    } catch (err) {
      const msg = err instanceof Error ? err.message : '加载预览地址失败'
      setPreviewError(msg)
      // Office 在转换失败时给出可打开原文件的兜底路径，避免整页不可用。
      if (['doc', 'docx', 'ppt'].includes(resource.resource_type)) {
        setPreviewUrl(resource.object_url || null)
      } else {
        setPreviewUrl(null)
      }
    } finally {
      setPreviewLoading(false)
    }
  }, [resourceId, resource, token])

  useEffect(() => {
    void loadPreviewUrl()
  }, [loadPreviewUrl])

  /**
   * 切换全屏模式
   */
  const toggleFullscreen = useCallback(() => {
    if (!containerRef.current) return

    if (!isFullscreen) {
      containerRef.current.requestFullscreen?.()
      setIsFullscreen(true)
    } else {
      document.exitFullscreen?.()
      setIsFullscreen(false)
    }
  }, [isFullscreen])

  /**
   * 监听全屏变化
   */
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(document.fullscreenElement !== null)
    }

    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange)
    }
  }, [])

  /**
   * 返回上一页
   */
  function handleGoBack() {
    navigate(-1)
  }

  /**
   * 按类型渲染预览：在 flex 容器内使用 min-h-0 + flex-1，使 iframe/视频撑满可用高度。
   */
  function renderPreview() {
    if (!resource) return null
    if (previewLoading) {
      return (
        <div className="flex min-h-0 flex-1 items-center justify-center text-muted-foreground">
          正在加载预览...
        </div>
      )
    }
    if (!previewUrl) {
      return (
        <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-3 text-muted-foreground">
          <p>当前资源暂时无法内嵌预览</p>
          {previewError && <p className="text-xs text-destructive">{previewError}</p>}
        </div>
      )
    }

    const { resource_type } = resource
    const mediaFill = 'h-full min-h-0 w-full flex-1 rounded-lg border-0 bg-background'

    switch (resource_type) {
      case 'pdf':
        return <iframe src={previewUrl} className={mediaFill} title={resource.title} />

      case 'video':
        return (
          <video
            ref={videoRef}
            src={previewUrl}
            controls
            className="h-full min-h-0 w-full flex-1 rounded-lg bg-black object-contain"
            preload="metadata"
          >
            您的浏览器不支持视频播放
          </video>
        )

      case 'audio':
        return (
          <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-6 px-4">
            <div className="text-6xl">🎵</div>
            <audio src={previewUrl} controls className="w-full max-w-md" preload="metadata">
              您的浏览器不支持音频播放
            </audio>
          </div>
        )

      case 'doc':
      case 'docx':
      case 'ppt':
        // 后端已将 Office 转为 PDF 预签名 URL，与 PDF 相同以内嵌查看，避免外链无法访问私有存储。
        return <iframe src={previewUrl} className={mediaFill} title={resource.title} />

      case 'txt':
        return <iframe src={previewUrl} className={mediaFill} title={resource.title} />

      default:
        return (
          <div className="flex min-h-0 flex-1 flex-col items-center justify-center text-muted-foreground">
            <div className="mb-4 text-4xl">📁</div>
            <p>暂不支持此类型资源的预览</p>
            <p className="mt-2 text-sm">资源类型：{resource_type}</p>
          </div>
        )
    }
  }

  const shellCentered = 'flex h-full min-h-0 flex-col overflow-y-auto'

  if (!resourceId) {
    return (
      <div className={shellCentered}>
        <div className="flex flex-1 flex-col items-center justify-center p-4">
          <Card className="w-full max-w-lg">
            <CardContent className="py-12">
              <div className="text-center">
                <p className="text-muted-foreground">缺少资源 ID 参数</p>
                <Button variant="outline" className="mt-4" onClick={handleGoBack}>
                  返回
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className={shellCentered}>
        <div className="flex flex-1 flex-col items-center justify-center p-4">
          <Card className="w-full max-w-lg">
            <CardContent className="py-12">
              <div className="text-center">
                <p className="text-muted-foreground">加载中...</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={shellCentered}>
        <div className="flex flex-1 flex-col items-center justify-center p-4">
          <Card className="w-full max-w-lg">
            <CardContent className="py-12">
              <div className="text-center">
                <p className="text-destructive">{error}</p>
                <Button variant="outline" className="mt-4" onClick={handleGoBack}>
                  返回
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
      <div className="shrink-0">
        <div className="mb-2 flex items-center gap-4">
          <Button variant="outline" size="sm" onClick={handleGoBack}>
            返回
          </Button>
        </div>
        <h1 className="text-3xl font-bold text-foreground">资源预览</h1>
      </div>

      {resource && (
        <Card className="shrink-0">
          <CardHeader>
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0 flex-1">
                <CardTitle className="line-clamp-2">
                  {getResourceTypeIcon(resource.resource_type)} {resource.title}
                </CardTitle>
              </div>
              <Badge variant="secondary">{getResourceTypeLabel(resource.resource_type)}</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1">
                <span>文件名：</span>
                <span className="font-medium text-foreground">
                  {resource.object_key?.split('/').pop() || '-'}
                </span>
              </div>
              <div className="flex items-center gap-1">
                <span>大小：</span>
                <span className="font-medium text-foreground">{formatFileSize(resource.size_bytes)}</span>
              </div>
              <div className="flex items-center gap-1">
                <span>类型：</span>
                <span className="font-medium text-foreground">{resource.mime_type || '-'}</span>
              </div>
              {resource.duration_seconds > 0 && (
                <div className="flex items-center gap-1">
                  <span>时长：</span>
                  <span className="font-medium text-foreground">
                    {Math.floor(resource.duration_seconds / 60)}:
                    {String(resource.duration_seconds % 60).padStart(2, '0')}
                  </span>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      <Card className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <CardHeader className="shrink-0">
          <div className="flex items-center justify-between">
            <CardTitle>预览</CardTitle>
            <div className="flex items-center gap-2">
              {resource?.object_url && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => window.open(resource.object_url, '_blank', 'noopener,noreferrer')}
                >
                  新窗口打开
                </Button>
              )}
              {resource && ['pdf', 'video'].includes(resource.resource_type) && (
                <Button variant="outline" size="sm" onClick={toggleFullscreen}>
                  {isFullscreen ? '退出全屏' : '全屏'}
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent className="flex min-h-0 flex-1 flex-col overflow-hidden pt-0">
          {previewError && ['doc', 'docx', 'ppt'].includes(resource?.resource_type ?? '') && (
            <div className="mb-2 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2 text-xs text-destructive">
              Office 预览转换失败：{previewError}
            </div>
          )}
          <div
            ref={containerRef}
            className={cn(
              'flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg',
              isFullscreen && 'fixed inset-0 z-50 bg-background',
            )}
          >
            {isFullscreen && (
              <div className="absolute right-4 top-4 z-10">
                <Button variant="secondary" size="sm" onClick={toggleFullscreen}>
                  退出全屏
                </Button>
              </div>
            )}
            <div
              className={cn(
                'flex min-h-0 flex-1 flex-col overflow-hidden',
                isFullscreen && 'p-4 pt-14',
              )}
            >
              {renderPreview()}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
