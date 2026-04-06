import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { listTeacherCourses, type TeacherCourseItem } from '../services/course'
import { listAIModels, type AIModelItem } from '../services/aiModels'
import {
  askInSessionStream,
  createChatSession,
  deleteChatSession,
  listChatMessages,
  listChatSessions,
  updateChatSessionTitle,
  type ChatMessageItem,
  type ChatSessionItem,
} from '../services/knowledgeChat'
import { useAuth } from '../store/auth'

// TeacherKnowledgeChatDetailPage 展示单会话历史并支持继续提问。
export function TeacherKnowledgeChatDetailPage() {
  const { token } = useAuth()
  const navigate = useNavigate()
  const { sessionId = '' } = useParams()
  const accessToken = token ?? ''

  const [messages, setMessages] = useState<ChatMessageItem[]>([])
  const [sessions, setSessions] = useState<ChatSessionItem[]>([])
  const [courses, setCourses] = useState<TeacherCourseItem[]>([])
  const [qaModels, setQaModels] = useState<AIModelItem[]>([])
  const [qaModelId, setQaModelId] = useState('')
  const [courseId, setCourseId] = useState('')
  const [question, setQuestion] = useState('')
  const [topK, setTopK] = useState(8)
  const [semanticMinScore, setSemanticMinScore] = useState(0)
  const [keywordMinScore, setKeywordMinScore] = useState(0)
  const [loading, setLoading] = useState(false)
  const [asking, setAsking] = useState(false)
  const [error, setError] = useState('')
  const [streamState, setStreamState] = useState<'idle' | 'streaming' | 'done'>('idle')
  const [editingSessionId, setEditingSessionId] = useState<string | null>(null)
  const [editingTitle, setEditingTitle] = useState('')
  const [refViewer, setRefViewer] = useState<{ title: string; content: string } | null>(null)
  const messageStreamRef = useRef<HTMLDivElement | null>(null)

  async function loadMessages() {
    if (!sessionId) return
    setLoading(true)
    setError('')
    try {
      const data = await listChatMessages(accessToken, sessionId, 1, 200)
      setMessages(data.items)
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载会话历史失败')
    } finally {
      setLoading(false)
    }
  }

  async function loadSessions() {
    try {
      const data = await listChatSessions(accessToken, 1, 100, courseId || undefined)
      setSessions(data.items)
    } catch {
      // ignore
    }
  }

  useEffect(() => {
    let cancelled = false
    async function loadCourses() {
      try {
        const data = await listTeacherCourses(accessToken, 1, 100, '', '', 'updated_at', 'desc')
        if (!cancelled) {
          setCourses(data.items)
          if (data.items.length > 0) setCourseId((v) => v || data.items[0].id)
        }
      } catch {
        // ignore
      }
    }
    void loadCourses()
    return () => {
      cancelled = true
    }
  }, [accessToken])

  useEffect(() => {
    let cancelled = false
    async function loadQAModels() {
      try {
        const data = await listAIModels(accessToken)
        if (cancelled) return
        const models = data.items.filter((m) => m.model_type === 'qa')
        setQaModels(models)
        if (models.length > 0) {
          setQaModelId((v) => v || models[0].id)
        }
      } catch {
        // ignore
      }
    }
    void loadQAModels()
    return () => {
      cancelled = true
    }
  }, [accessToken])

  useEffect(() => {
    void loadMessages()
  }, [sessionId, accessToken])

  useEffect(() => {
    void loadSessions()
  }, [courseId, accessToken, sessionId])

  useEffect(() => {
    const node = messageStreamRef.current
    if (!node) return
    node.scrollTop = node.scrollHeight
  }, [messages, loading])

  async function handleAsk() {
    if (!question.trim()) return
    setAsking(true)
    setError('')
    setStreamState('streaming')
    let done = false
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
      },
    ])
    try {
      let finalRefs: Array<{ citation_no?: number; resource_title: string; chunk_index: number; snippet: string; full_content?: string }> = []
      await askInSessionStream(
        accessToken,
        sessionId,
        askText,
        topK,
        qaModelId || undefined,
        semanticMinScore,
        keywordMinScore,
        {
        onToken: (token) => {
          setMessages((prev) =>
            prev.map((m) => (m.id === tmpAssistantId ? { ...m, content: `${m.content}${token}` } : m)),
          )
        },
        onReferences: (refs) => {
          finalRefs = refs.map((r) => ({
            citation_no: r.citation_no,
            resource_title: r.resource_title,
            chunk_index: r.chunk_index,
            snippet: r.snippet,
            full_content: r.full_content,
          }))
          setMessages((prev) =>
            prev.map((m) => (m.id === tmpAssistantId ? { ...m, references_json: finalRefs } : m)),
          )
        },
        onDone: () => {
          done = true
          setStreamState('done')
        },
        onError: (msg) => {
          throw new Error(msg)
        },
      })
      setQuestion('')
      await loadMessages() // 与服务端最终状态对齐
    } catch (e) {
      setMessages((prev) => prev.filter((m) => m.id !== tmpUserId && m.id !== tmpAssistantId))
      setError(e instanceof Error ? e.message : '提问失败')
    } finally {
      setAsking(false)
      if (!done) setStreamState('idle')
    }
  }

  async function handleCreateInCourse() {
    if (!courseId) return
    setError('')
    try {
      const data = await createChatSession(accessToken, courseId, '')
      navigate(`/teacher/knowledge/chats/${data.id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : '创建会话失败')
    }
  }

  async function handleRenameSession(targetSessionId: string) {
    if (!editingTitle.trim()) {
      setError('标题不能为空')
      return
    }
    try {
      await updateChatSessionTitle(accessToken, targetSessionId, editingTitle.trim())
      setEditingSessionId(null)
      setEditingTitle('')
      await loadSessions()
      if (targetSessionId === sessionId) await loadMessages()
    } catch (e) {
      setError(e instanceof Error ? e.message : '修改标题失败')
    }
  }

  async function handleDeleteSession(targetSessionId: string) {
    if (!window.confirm('确定删除该会话及其所有历史消息？')) return
    try {
      await deleteChatSession(accessToken, targetSessionId)
      await loadSessions()
      if (targetSessionId === sessionId) {
        navigate('/teacher/knowledge/chats')
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : '删除会话失败')
    }
  }

  function getCitationNo(ref: NonNullable<ChatMessageItem['references_json']>[number], idx: number) {
    return typeof ref.citation_no === 'number' && ref.citation_no > 0 ? ref.citation_no : idx + 1
  }

  const title = useMemo(() => {
    const currentSession = sessions.find((s) => s.id === sessionId)
    if (currentSession?.title?.trim()) return currentSession.title
    const firstUser = messages.find((m) => m.role === 'user')
    if (!firstUser) return `会话 ${sessionId}`
    return firstUser.content.length > 32 ? `${firstUser.content.slice(0, 32)}...` : firstUser.content
  }, [sessions, messages, sessionId])

  return (
    <div className="chat-page-shell">
      <aside className="chat-left-panel">
        <div className="chat-left-header">
          <h3>对话管理</h3>
          <button onClick={() => navigate('/teacher/knowledge/chats')}>列表页</button>
        </div>
        <label>
          课程（RAG基础）
          <select value={courseId} onChange={(e) => setCourseId(e.target.value)}>
            {courses.map((c) => (
              <option key={c.id} value={c.id}>
                {c.title}
              </option>
            ))}
          </select>
        </label>
        <button className="primary" disabled={!courseId} onClick={() => void handleCreateInCourse()}>
          新建该课程对话
        </button>
        <div className="chat-session-list">
          {sessions.map((s) => (
            <div key={s.id} className={s.id === sessionId ? 'chat-session-item active' : 'chat-session-item'}>
              <button className="chat-session-link" onClick={() => navigate(`/teacher/knowledge/chats/${s.id}`)}>
                <strong>{s.title}</strong>
                <small>{new Date(s.updated_at).toLocaleString()}</small>
              </button>
              <div className="chat-session-actions">
                <button
                  type="button"
                  className="ghost"
                  onClick={() => {
                    setEditingSessionId(s.id)
                    setEditingTitle(s.title)
                  }}
                >
                  重命名
                </button>
                <button type="button" className="danger" onClick={() => void handleDeleteSession(s.id)}>
                  删除
                </button>
              </div>
              {editingSessionId === s.id && (
                <div className="chat-session-edit-row">
                  <input value={editingTitle} onChange={(e) => setEditingTitle(e.target.value)} maxLength={256} />
                  <button type="button" className="primary" onClick={() => void handleRenameSession(s.id)}>
                    保存
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setEditingSessionId(null)
                      setEditingTitle('')
                    }}
                  >
                    取消
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      </aside>

      <section className="chat-main-panel">
        <header className="chat-main-header">
          <h2>{title}</h2>
          <div className="action-row">
            <label style={{ width: 140 }}>
              TopK
              <input type="number" min={1} max={20} value={topK} onChange={(e) => setTopK(Number(e.target.value))} />
            </label>
            <label style={{ width: 170 }}>
              语义阈值
              <input
                type="number"
                min={0}
                max={1}
                step={0.01}
                value={semanticMinScore}
                onChange={(e) => setSemanticMinScore(Number(e.target.value))}
              />
            </label>
            <label style={{ width: 170 }}>
              关键词阈值
              <input
                type="number"
                min={0}
                max={1}
                step={0.01}
                value={keywordMinScore}
                onChange={(e) => setKeywordMinScore(Number(e.target.value))}
              />
            </label>
            <label style={{ width: 220 }}>
              问答模型
              <select value={qaModelId} onChange={(e) => setQaModelId(e.target.value)}>
                {qaModels.length === 0 && <option value="">未配置 QA 模型</option>}
                {qaModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.model_id})
                  </option>
                ))}
              </select>
            </label>
            <button onClick={() => void loadMessages()} disabled={loading}>
              刷新
            </button>
          </div>
        </header>
        {error && <div className="alert error">{error}</div>}
        <div className="chat-stream-state subtle">
          <span className={streamState === 'streaming' ? 'dot-streaming' : streamState === 'done' ? 'dot-done' : 'dot-idle'} />
          {streamState === 'streaming' ? '流式输出中...' : streamState === 'done' ? '输出完成' : '等待提问'}
        </div>

        <div ref={messageStreamRef} className="chat-message-stream">
          {loading && <p className="subtle">加载中...</p>}
          {!loading && messages.length === 0 && <p className="subtle">开始提问吧。</p>}
          {messages.map((m) => (
            <article key={m.id} className={m.role === 'user' ? 'chat-bubble user' : 'chat-bubble assistant'}>
              <div className="chat-bubble-meta">
                <strong>{m.role === 'assistant' ? '助手' : m.role === 'user' ? '你' : '系统'}</strong>
                <span className="mono subtle">{new Date(m.created_at).toLocaleString()}</span>
              </div>
              {m.role === 'assistant' ? (
                <div className="chat-markdown-body">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>{m.content}</ReactMarkdown>
                </div>
              ) : (
                <pre className="chat-bubble-content">{m.content}</pre>
              )}
              {m.role === 'assistant' && Array.isArray(m.references_json) && m.references_json.length > 0 && (
                <div className="chat-reference-blocks">
                  <div className="chat-reference-blocks-title">引用内容</div>
                  <div className="chat-reference-blocks-list">
                    {m.references_json.map((ref, idx) => (
                      <div key={idx} className="chat-reference-item">
                        <span className="mono subtle">
                          [{getCitationNo(ref, idx)}] {ref.resource_title ?? '未知资源'}
                          {typeof ref.chunk_index === 'number' ? ` #${ref.chunk_index + 1}` : ''}
                        </span>
                        <button
                          type="button"
                          className="ghost"
                          onClick={() =>
                            setRefViewer({
                              title: `[${getCitationNo(ref, idx)}] ${ref.resource_title ?? '未知资源'}`,
                              content: ref.full_content || ref.snippet || '',
                            })
                          }
                        >
                          查看片段
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </article>
          ))}
        </div>

        <div className="chat-input-box">
          <textarea
            rows={2}
            placeholder="给当前课程知识库发送消息（Enter 发送，Shift+Enter 换行）"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                if (!asking && question.trim()) void handleAsk()
              }
            }}
          />
          <button className="primary chat-send-btn" disabled={asking || !question.trim()} onClick={() => void handleAsk()}>
            {asking ? '思考中...' : '发送'}
          </button>
        </div>
      </section>

      {refViewer && (
        <div className="modal-mask" role="presentation">
          <div className="modal-card" role="dialog" aria-modal="true" style={{ maxWidth: 860, width: 'min(860px, 95vw)' }}>
            <div className="modal-card-header">
              <h3>{refViewer.title}</h3>
              <button type="button" className="modal-close-btn" onClick={() => setRefViewer(null)}>
                ×
              </button>
            </div>
            <pre className="chat-ref-viewer">{refViewer.content}</pre>
            <div className="action-row" style={{ justifyContent: 'flex-end' }}>
              <button onClick={() => setRefViewer(null)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
