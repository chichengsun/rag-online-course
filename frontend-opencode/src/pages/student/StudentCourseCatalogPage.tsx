import { useEffect, useState, useCallback } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { getCourseCatalog } from '@/services/course'
import type { CatalogChapter, CatalogResource } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from '@/components/ui/collapsible'
import {
  ChevronDown,
  ChevronRight,
  ArrowLeft,
  BookOpen,
  FileText,
  Video,
  Music,
  Presentation,
  CheckCircle2,
  Circle,
} from 'lucide-react'
import { toast } from 'sonner'

const resourceTypeConfig: Record<string, { label: string; icon: React.ReactNode; color: string }> = {
  pdf: { label: 'PDF', icon: <FileText className="w-4 h-4" />, color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' },
  doc: { label: 'DOC', icon: <FileText className="w-4 h-4" />, color: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' },
  docx: { label: 'DOCX', icon: <FileText className="w-4 h-4" />, color: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' },
  ppt: { label: 'PPT', icon: <Presentation className="w-4 h-4" />, color: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' },
  txt: { label: 'TXT', icon: <FileText className="w-4 h-4" />, color: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400' },
  video: { label: '视频', icon: <Video className="w-4 h-4" />, color: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' },
  audio: { label: '音频', icon: <Music className="w-4 h-4" />, color: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' },
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

interface ResourceWithProgress extends CatalogResource {
  isCompleted?: boolean
  progressPercent?: number
}

interface SectionWithProgress {
  id: string
  title: string
  sort_order: number
  resources: ResourceWithProgress[]
  completedCount: number
  totalCount: number
}

interface ChapterWithProgress {
  id: string
  title: string
  sort_order: number
  sections: SectionWithProgress[]
  completedCount: number
  totalCount: number
}

export function StudentCourseCatalogPage() {
  const { token } = useAuthStore()
  const { courseId } = useParams<{ courseId: string }>()
  const navigate = useNavigate()
  const accessToken = token ?? ''

  const [chapters, setChapters] = useState<ChapterWithProgress[]>([])
  const [expandedChapters, setExpandedChapters] = useState<Set<string>>(new Set())
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)

  const loadCatalog = useCallback(async () => {
    if (!courseId) return
    const courseIdNum = parseInt(courseId, 10)
    if (isNaN(courseIdNum)) return

    setLoading(true)
    try {
      const catalogData = await getCourseCatalog(accessToken, courseIdNum)
      
      const chaptersWithProgress: ChapterWithProgress[] = (catalogData.chapters || []).map((chapter: CatalogChapter) => {
        const sectionsWithProgress: SectionWithProgress[] = (chapter.sections || []).map((section) => {
          const resourcesWithProgress: ResourceWithProgress[] = (section.resources || []).map((resource) => ({
            ...resource,
            isCompleted: false,
            progressPercent: 0,
          }))
          
          return {
            ...section,
            resources: resourcesWithProgress,
            completedCount: resourcesWithProgress.filter((r) => r.isCompleted).length,
            totalCount: resourcesWithProgress.length,
          }
        })

        const totalResources = sectionsWithProgress.reduce((sum, s) => sum + s.totalCount, 0)
        const completedResources = sectionsWithProgress.reduce((sum, s) => sum + s.completedCount, 0)

        return {
          ...chapter,
          sections: sectionsWithProgress,
          completedCount: completedResources,
          totalCount: totalResources,
        }
      })

      setChapters(chaptersWithProgress)

      if (chaptersWithProgress.length > 0 && expandedChapters.size === 0) {
        setExpandedChapters(new Set([chaptersWithProgress[0].id]))
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '加载课程目录失败')
    } finally {
      setLoading(false)
    }
  }, [accessToken, courseId, expandedChapters.size])

  useEffect(() => {
    void loadCatalog()
  }, [loadCatalog])

  const toggleChapter = (chapterId: string) => {
    setExpandedChapters((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(chapterId)) {
        newSet.delete(chapterId)
      } else {
        newSet.add(chapterId)
      }
      return newSet
    })
  }

  const toggleSection = (sectionId: string) => {
    setExpandedSections((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(sectionId)) {
        newSet.delete(sectionId)
      } else {
        newSet.add(sectionId)
      }
      return newSet
    })
  }

  const handleResourceClick = (resource: ResourceWithProgress) => {
    navigate(`/student/courses/${courseId}/resource/${resource.id}`)
  }

  const totalResources = chapters.reduce((sum, c) => sum + c.totalCount, 0)
  const completedResources = chapters.reduce((sum, c) => sum + c.completedCount, 0)
  const overallProgress = totalResources > 0 ? Math.round((completedResources / totalResources) * 100) : 0

  if (!courseId) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>未选择课程</CardTitle>
          </CardHeader>
          <CardContent>
            <Button onClick={() => navigate('/student/my-courses')}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              返回我的课程
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => navigate('/student/my-courses')}
            >
              <ArrowLeft className="w-4 h-4 mr-1" />
              返回
            </Button>
          </div>
          <h1 className="text-3xl font-bold text-foreground">
            课程目录
          </h1>
          <p className="text-muted-foreground mt-2">
            完成进度：{completedResources}/{totalResources} 个资源 ({overallProgress}%)
          </p>
        </div>
      </div>

      {totalResources > 0 && (
        <Card>
          <CardContent className="py-4">
            <div className="flex items-center gap-4">
              <div className="flex-1">
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all duration-300"
                    style={{ width: `${overallProgress}%` }}
                  />
                </div>
              </div>
              <span className="text-sm font-medium text-muted-foreground">
                {overallProgress}%
              </span>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="space-y-4">
        {chapters.length === 0 && !loading && (
          <Card>
            <CardContent className="py-12">
              <div className="text-center">
                <BookOpen className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
                <p className="text-muted-foreground">暂无课程内容</p>
              </div>
            </CardContent>
          </Card>
        )}

        {chapters.map((chapter) => {
          const isChapterExpanded = expandedChapters.has(chapter.id)
          const chapterProgress = chapter.totalCount > 0
            ? Math.round((chapter.completedCount / chapter.totalCount) * 100)
            : 0

          return (
            <Collapsible
              key={chapter.id}
              open={isChapterExpanded}
              onOpenChange={() => toggleChapter(chapter.id)}
            >
              <Card>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <CollapsibleTrigger className="flex-1">
                      <Button
                        variant="ghost"
                        className="w-full justify-start gap-3 p-0 h-auto"
                      >
                        {isChapterExpanded ? (
                          <ChevronDown className="w-5 h-5 text-muted-foreground" />
                        ) : (
                          <ChevronRight className="w-5 h-5 text-muted-foreground" />
                        )}
                        <div className="flex flex-col items-start gap-1">
                          <span className="text-lg font-semibold">
                            第 {chapter.sort_order} 章：{chapter.title}
                          </span>
                          <span className="text-sm text-muted-foreground">
                            {chapter.completedCount}/{chapter.totalCount} 已完成 ({chapterProgress}%)
                          </span>
                        </div>
                      </Button>
                    </CollapsibleTrigger>
                    <div className="flex items-center gap-2">
                      {chapter.completedCount === chapter.totalCount && chapter.totalCount > 0 && (
                        <CheckCircle2 className="w-5 h-5 text-green-500" />
                      )}
                    </div>
                  </div>
                </CardHeader>

                <CollapsibleContent>
                  <CardContent className="pt-0">
                    <div className="pl-8 space-y-3">
                      {chapter.sections.length === 0 && (
                        <p className="text-muted-foreground text-sm py-4">
                          该章节下暂无内容
                        </p>
                      )}

                      {chapter.sections.map((section) => {
                        const isSectionExpanded = expandedSections.has(section.id)

                        return (
                          <Collapsible
                            key={section.id}
                            open={isSectionExpanded}
                            onOpenChange={() => toggleSection(section.id)}
                          >
                            <div className="border rounded-lg overflow-hidden">
                              <CollapsibleTrigger className="w-full">
                                <Button
                                  variant="ghost"
                                  className="w-full justify-start gap-3 rounded-none h-auto py-3 px-4"
                                >
                                  {isSectionExpanded ? (
                                    <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
                                  ) : (
                                    <ChevronRight className="w-4 h-4 text-muted-foreground shrink-0" />
                                  )}
                                  <div className="flex flex-col items-start gap-1 flex-1">
                                    <span className="font-medium">
                                      {section.sort_order}. {section.title}
                                    </span>
                                    <span className="text-xs text-muted-foreground">
                                      {section.completedCount}/{section.totalCount} 已完成
                                    </span>
                                  </div>
                                  {section.completedCount === section.totalCount && section.totalCount > 0 && (
                                    <CheckCircle2 className="w-4 h-4 text-green-500 shrink-0" />
                                  )}
                                </Button>
                              </CollapsibleTrigger>

                              <CollapsibleContent>
                                <div className="border-t bg-muted/30">
                                  {section.resources.length === 0 && (
                                    <p className="text-muted-foreground text-sm py-4 px-6">
                                      该节下暂无资源
                                    </p>
                                  )}

                                  {section.resources.map((resource) => {
                                    const typeConfig = resourceTypeConfig[resource.resource_type] || resourceTypeConfig.txt
                                    const isCompleted = resource.isCompleted

                                    return (
                                      <button
                                        key={resource.id}
                                        onClick={() => handleResourceClick(resource)}
                                        className="w-full flex items-center justify-between py-3 px-6 hover:bg-muted/50 transition-colors text-left border-b last:border-b-0"
                                      >
                                        <div className="flex items-center gap-3 flex-1 min-w-0">
                                          {isCompleted ? (
                                            <CheckCircle2 className="w-4 h-4 text-green-500 shrink-0" />
                                          ) : (
                                            <Circle className="w-4 h-4 text-muted-foreground shrink-0" />
                                          )}
                                          <Badge
                                            variant="secondary"
                                            className={`shrink-0 ${typeConfig.color}`}
                                          >
                                            {typeConfig.icon}
                                            <span className="ml-1">{typeConfig.label}</span>
                                          </Badge>
                                          <span className="text-sm truncate">
                                            {resource.title}
                                          </span>
                                        </div>
                                        <div className="flex items-center gap-3 shrink-0">
                                          <span className="text-xs text-muted-foreground">
                                            {formatFileSize(resource.size_bytes)}
                                          </span>
                                          {resource.ai_summary && (
                                            <Badge variant="outline" className="text-xs">
                                              AI摘要
                                            </Badge>
                                          )}
                                        </div>
                                      </button>
                                    )
                                  })}
                                </div>
                              </CollapsibleContent>
                            </div>
                          </Collapsible>
                        )
                      })}
                    </div>
                  </CardContent>
                </CollapsibleContent>
              </Card>
            </Collapsible>
          )
        })}
      </div>
    </div>
  )
}