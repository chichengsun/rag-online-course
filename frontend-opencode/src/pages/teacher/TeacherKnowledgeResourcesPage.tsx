/**
 * 教师知识库资源列表页面
 * 显示课程下可解析资源的分页列表，提供分块与嵌入入口
 */
import { useEffect, useState, useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getKnowledgeResources } from '@/services/knowledge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

/**
 * 知识库资源列表项（扩展 Resource 类型，包含分块统计）
 */
interface KnowledgeResourceItem {
  id: string
  chapter_id: string
  section_id: string
  chapter_title: string
  section_title: string
  title: string
  resource_type: string
  total_chunk_chars: number
  chunk_count: number
  embedded_count: number
}

/**
 * 资源状态类型
 */
type ResourceStatus = 'none' | 'chunked' | 'embedded'

/**
 * 根据分块和嵌入数量判断资源状态
 */
function getResourceStatus(chunkCount: number, embeddedCount: number): ResourceStatus {
  if (embeddedCount > 0) {
    return 'embedded'
  }
  if (chunkCount > 0) {
    return 'chunked'
  }
  return 'none'
}

/**
 * 获取状态标签配置
 */
function getStatusBadge(status: ResourceStatus) {
  switch (status) {
    case 'embedded':
      return { variant: 'success' as const, label: '已嵌入' }
    case 'chunked':
      return { variant: 'info' as const, label: '已分块' }
    default:
      return { variant: 'secondary' as const, label: '未分块' }
  }
}

/**
 * 格式化数字显示
 */
function formatNumber(value: number | string | undefined): number {
  if (value === undefined || value === null) return 0
  return typeof value === 'number' ? value : Number(value)
}

export function TeacherKnowledgeResourcesPage() {
  const { courseId = '' } = useParams<{ courseId: string }>()
  const navigate = useNavigate()
  const { token } = useAuthStore()

  const [resources, setResources] = useState<KnowledgeResourceItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [page, setPage] = useState(1)
  const [pageSize] = useState(10)
  const [total, setTotal] = useState(0)

  /**
   * 加载知识库资源列表
   */
  const loadResources = useCallback(async (targetPage: number = 1) => {
    if (!courseId || !token) return

    setLoading(true)
    setError(null)

    try {
      const data = await getKnowledgeResources(Number(courseId), {
        page: targetPage,
        page_size: pageSize,
      })
      setResources(data.items as unknown as KnowledgeResourceItem[])
      setTotal(Number(data.total))
      setPage(targetPage)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载资源列表失败')
    } finally {
      setLoading(false)
    }
  }, [courseId, token])

  useEffect(() => {
    void loadResources(1)
  }, [loadResources])

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  /**
   * 跳转到分块管理页面
   */
  function handleResourceClick(resource: KnowledgeResourceItem) {
    navigate(`/teacher/knowledge/${courseId}/chunk/${resource.id}`)
  }

  /**
   * 返回课程列表
   */
  function handleGoBack() {
    navigate('/teacher/courses')
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-4 mb-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleGoBack}
          >
            返回课程列表
          </Button>
        </div>
        <h1 className="text-3xl font-bold text-foreground">知识库资源</h1>
        <p className="text-muted-foreground mt-2">
          课程 ID：{courseId} · 仅列出支持解析与分块的文档类型
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
            <div className="text-center">
              <p className="text-muted-foreground">加载中...</p>
            </div>
          </CardContent>
        </Card>
      )}

      {!loading && resources.length === 0 && (
        <Card>
          <CardContent className="py-12">
            <div className="text-center">
              <p className="text-muted-foreground mb-4">
                暂无符合条件的资源
              </p>
              <p className="text-sm text-muted-foreground">
                请先在课程内容管理中上传文档资源
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {!loading && resources.length > 0 && (
        <>
          <div className="grid gap-4">
            {resources.map((resource) => {
              const status = getResourceStatus(
                formatNumber(resource.chunk_count),
                formatNumber(resource.embedded_count)
              )
              const statusBadge = getStatusBadge(status)

              return (
                <Card
                  key={resource.id}
                  className="cursor-pointer hover:border-primary/50 transition-colors"
                  onClick={() => handleResourceClick(resource)}
                >
                  <CardHeader>
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex-1 min-w-0">
                        <CardTitle className="line-clamp-2">
                          {resource.title}
                        </CardTitle>
                        <CardDescription className="mt-1">
                          {resource.chapter_title}
                          {resource.section_title ? ` · ${resource.section_title}` : ''} · {resource.resource_type.toUpperCase()}
                        </CardDescription>
                      </div>
                      <Badge variant={statusBadge.variant}>
                        {statusBadge.label}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <span>分块数：</span>
                        <span className="font-medium text-foreground">
                          {formatNumber(resource.chunk_count)}
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span>已嵌入：</span>
                        <span className="font-medium text-foreground">
                          {formatNumber(resource.embedded_count)}
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <span>字符数：</span>
                        <span className="font-medium text-foreground">
                          {formatNumber(resource.total_chunk_chars).toLocaleString()}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>

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
                      onClick={() => void loadResources(page - 1)}
                    >
                      上一页
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={page >= totalPages || loading}
                      onClick={() => void loadResources(page + 1)}
                    >
                      下一页
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  )
}