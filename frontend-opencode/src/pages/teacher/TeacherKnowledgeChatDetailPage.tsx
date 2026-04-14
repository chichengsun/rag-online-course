import { useEffect, useState, useRef, useCallback, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { getMessages, askInSessionStream } from '@/services/chat'
import { getModels } from '@/services/aiModels'
import type { AIModelListItem, ReferenceItem } from '@/types'
import { ChatMessageList, type Reference } from '@/components/ChatMessage'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Slider } from '@/components/ui/slider'
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
} from '@/components/ui/dialog'
import { ArrowLeft, Send, Loader2 } from 'lucide-react'

interface MessageWithStreaming {
  id: string | number
  session_id: string | number
  role: 'user' | 'assistant'
  content: string
  created_at: string
  references?: Reference[]
  isStreaming?: boolean
}

export function TeacherKnowledgeChatDetailPage() {
  const navigate = useNavigate()
  const { sessionId } = useParams<{ sessionId: string }>()

  const [messages, setMessages] = useState<MessageWithStreaming[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [question, setQuestion] = useState('')
  const [asking, setAsking] = useState(false)

  const [topK, setTopK] = useState(8)
  const [semanticMinScore, setSemanticMinScore] = useState(0)
  const [keywordMinScore, setKeywordMinScore] = useState(0)
  const [qaModelId, setQaModelId] = useState<string>('')
  const [qaModels, setQaModels] = useState<AIModelListItem[]>([])

  const [refViewer, setRefViewer] = useState<{ title: string; content: string } | null>(null)

  const messageListRef = useRef<HTMLDivElement>(null)

  const loadMessages = useCallback(async () => {
    if (!sessionId) return
    setLoading(true)
    setError(null)
    try {
      const data = await getMessages(sessionId)
      setMessages(
        data.map((msg) => ({
          ...msg,
          references: msg.references?.map((ref) => ({
            citationNo: ref.citation_no,
            resourceTitle: ref.resource_title,
            chunkIndex: ref.chunk_index,
            snippet: ref.snippet,
            fullContent: ref.full_content,
          })),
        }))
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载会话历史失败')
    } finally {
      setLoading(false)
    }
  }, [sessionId])

  const loadQAModels = useCallback(async () => {
    try {
      const data = await getModels()
      const qaModelList = data.filter((m) => m.model_type === 'qa')
      setQaModels(qaModelList)
      if (qaModelList.length > 0 && !qaModelId) {
        // 接口 id 可能为 number，与 Select 的 string value 统一，否则无法匹配 Item 文案、触发器会只显示数字 id
        setQaModelId(String(qaModelList[0].id))
      }
    } catch (e) {
      console.warn('加载问答模型列表失败（可忽略）', e)
    }
  }, [qaModelId])

  useEffect(() => {
    void loadQAModels()
    void loadMessages()
  }, [loadQAModels, loadMessages])

  useEffect(() => {
    if (messageListRef.current) {
      messageListRef.current.scrollTop = messageListRef.current.scrollHeight
    }
  }, [messages])

  const handleAsk = async () => {
    if (!question.trim() || asking || !sessionId) return

    setAsking(true)
    setError(null)

    const askText = question.trim()
    const tmpUserId = `tmp-user-${Date.now()}`
    const tmpAssistantId = `tmp-assistant-${Date.now()}`

    setMessages((prev) => [
      ...prev,
      {
        id: tmpUserId,
        session_id: sessionId,
        role: 'user',
        content: askText,
        created_at: new Date().toISOString(),
      },
      {
        id: tmpAssistantId,
        session_id: sessionId,
        role: 'assistant',
        content: '',
        created_at: new Date().toISOString(),
        isStreaming: true,
      },
    ])

    setQuestion('')

    try {
      let finalRefs: Reference[] = []

      await askInSessionStream(
        sessionId,
        {
          question: askText,
          top_k: topK,
          qa_model_id: qaModelId ? Number(qaModelId) : undefined,
          semantic_min_score: semanticMinScore,
          keyword_min_score: keywordMinScore,
        },
        {
          onToken: (token) => {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === tmpAssistantId ? { ...m, content: m.content + token } : m
              )
            )
          },
          onReferences: (refs: ReferenceItem[]) => {
            finalRefs = refs.map((r) => ({
              citationNo: r.citation_no,
              resourceTitle: r.resource_title,
              chunkIndex: r.chunk_index,
              snippet: r.snippet,
              fullContent: r.full_content,
            }))
            setMessages((prev) =>
              prev.map((m) =>
                m.id === tmpAssistantId ? { ...m, references: finalRefs } : m
              )
            )
          },
          onDone: () => {
            setMessages((prev) =>
              prev.map((m) =>
                m.id === tmpAssistantId ? { ...m, isStreaming: false } : m
              )
            )
          },
        }
      )

      await loadMessages()
    } catch (err) {
      setMessages((prev) => prev.filter((m) => m.id !== tmpUserId && m.id !== tmpAssistantId))
      setError(err instanceof Error ? err.message : '提问失败')
    } finally {
      setAsking(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!asking && question.trim()) {
        void handleAsk()
      }
    }
  }

  const handleReferenceClick = (reference: Reference) => {
    setRefViewer({
      title: `[${reference.citationNo ?? 1}] ${reference.resourceTitle}`,
      content: reference.fullContent || reference.snippet,
    })
  }

  const getSessionTitle = () => {
    const firstUserMessage = messages.find((m) => m.role === 'user')
    if (!firstUserMessage) return `会话 ${sessionId}`
    return firstUserMessage.content.length > 32
      ? `${firstUserMessage.content.slice(0, 32)}...`
      : firstUserMessage.content
  }

  /** 当前选中的问答模型（id 与下拉 value 均按字符串比较） */
  const selectedQaModel = useMemo(
    () => qaModels.find((m) => String(m.id) === qaModelId),
    [qaModels, qaModelId],
  )

  const qaModelTriggerLabel = selectedQaModel
    ? `${selectedQaModel.name} (${selectedQaModel.model_id})`
    : undefined

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center justify-between border-b border-border bg-background px-4 py-3 lg:px-6">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/teacher/knowledge/chats')}
          >
            <ArrowLeft className="size-4 mr-2" />
            返回列表
          </Button>
          <h1 className="text-xl font-semibold text-foreground">
            {getSessionTitle()}
          </h1>
        </div>
      </div>

      {error && (
        <div className="shrink-0 border-b border-destructive/20 bg-destructive/10 px-4 py-3 lg:px-6">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      <div className="flex min-h-0 flex-1 flex-col gap-0 overflow-hidden lg:flex-row">
        <div className="flex min-h-0 min-w-0 flex-1 flex-col">
          <div ref={messageListRef} className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-3 lg:px-6">
            {loading ? (
              <div className="flex items-center justify-center h-full">
                <Loader2 className="size-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <ChatMessageList
                messages={messages.map((m) => ({
                  id: String(m.id),
                  role: m.role,
                  content: m.content,
                  createdAt: m.created_at,
                  references: m.references,
                  isStreaming: m.isStreaming,
                }))}
                onReferenceClick={handleReferenceClick}
              />
            )}
          </div>

          <div className="shrink-0 border-t border-border bg-background p-4">
            <div className="flex gap-3">
              <Textarea
                placeholder="输入问题（Enter 发送，Shift+Enter 换行）"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                onKeyDown={handleKeyDown}
                disabled={asking}
                className="flex-1 min-h-[60px] max-h-[200px] resize-none"
                rows={2}
              />
              <Button
                onClick={handleAsk}
                disabled={asking || !question.trim()}
                className="self-end"
              >
                {asking ? (
                  <>
                    <Loader2 className="size-4 mr-2 animate-spin" />
                    思考中
                  </>
                ) : (
                  <>
                    <Send className="size-4 mr-2" />
                    发送
                  </>
                )}
              </Button>
            </div>
          </div>
        </div>

        <Card className="flex max-h-[40vh] w-full shrink-0 flex-col border-t border-border lg:max-h-none lg:w-72 lg:border-l lg:border-t-0 lg:shadow-none">
          <CardHeader className="shrink-0 pb-2 pt-3">
            <CardTitle className="text-base">参数调节</CardTitle>
          </CardHeader>
          <CardContent className="min-h-0 flex-1 space-y-4 overflow-y-auto px-4 pb-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium text-foreground">TopK</label>
                <span className="text-sm text-muted-foreground">{topK}</span>
              </div>
              <Slider
                value={[topK]}
                onValueChange={(value: number[]) => setTopK(value[0])}
                min={1}
                max={20}
                step={1}
                className="w-full"
              />
              <p className="text-xs text-muted-foreground">
                检索返回的最大文档数量
              </p>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium text-foreground">语义阈值</label>
                <span className="text-sm text-muted-foreground">
                  {semanticMinScore.toFixed(2)}
                </span>
              </div>
              <Slider
                value={[semanticMinScore]}
                onValueChange={(value: number[]) => setSemanticMinScore(value[0])}
                min={0}
                max={1}
                step={0.01}
                className="w-full"
              />
              <p className="text-xs text-muted-foreground">
                语义相似度最低阈值
              </p>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium text-foreground">关键词阈值</label>
                <span className="text-sm text-muted-foreground">
                  {keywordMinScore.toFixed(2)}
                </span>
              </div>
              <Slider
                value={[keywordMinScore]}
                onValueChange={(value: number[]) => setKeywordMinScore(value[0])}
                min={0}
                max={1}
                step={0.01}
                className="w-full"
              />
              <p className="text-xs text-muted-foreground">
                关键词匹配最低阈值
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-foreground">问答模型</label>
              <Select
                value={qaModelId}
                onValueChange={(value) => value && setQaModelId(value)}
                disabled={qaModels.length === 0}
              >
                <SelectTrigger className="w-full min-w-0">
                  <SelectValue placeholder="选择问答模型">{qaModelTriggerLabel}</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {qaModels.map((model) => (
                    <SelectItem key={String(model.id)} value={String(model.id)}>
                      {model.name} ({model.model_id})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {qaModels.length === 0 && (
                <p className="text-xs text-muted-foreground">
                  暂无可用模型，请先配置 QA 模型
                </p>
              )}
            </div>

            <Button
              variant="outline"
              className="w-full"
              onClick={loadMessages}
              disabled={loading}
            >
              {loading ? (
                <>
                  <Loader2 className="size-4 mr-2 animate-spin" />
                  加载中
                </>
              ) : (
                '刷新历史'
              )}
            </Button>
          </CardContent>
        </Card>
      </div>

      <Dialog open={!!refViewer} onOpenChange={(open) => !open && setRefViewer(null)}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>{refViewer?.title}</DialogTitle>
          </DialogHeader>
          <div className="flex-1 overflow-y-auto">
            <pre className="text-sm text-muted-foreground whitespace-pre-wrap leading-relaxed">
              {refViewer?.content}
            </pre>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}