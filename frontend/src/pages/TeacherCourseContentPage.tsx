import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  confirmResource,
  createChapter,
  deleteChapter,
  deleteResource,
  initUpload,
  listChapterResources,
  listCourseChapters,
  type ChapterItem,
  type ConfirmResourcePayload,
  type ResourceItem,
} from '../services/course'
import { useAuth } from '../store/auth'

type ResourceType = 'ppt' | 'pdf' | 'txt' | 'video' | 'doc' | 'docx' | 'audio'

export function TeacherCourseContentPage() {
  const { token } = useAuth()
  const { courseId = '' } = useParams()
  const navigate = useNavigate()
  const accessToken = token ?? ''

  const [chapters, setChapters] = useState<ChapterItem[]>([])
  const [resourcesByChapter, setResourcesByChapter] = useState<Record<string, ResourceItem[]>>({})
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [showChapterModal, setShowChapterModal] = useState(false)
  const [showResourceModal, setShowResourceModal] = useState(false)
  const [chapterForm, setChapterForm] = useState({ title: '', sort_order: 1 })
  const [uploadChapterId, setUploadChapterID] = useState('')
  const [resourceFile, setResourceFile] = useState<File | null>(null)
  const [resourceForm, setResourceForm] = useState<ConfirmResourcePayload>({
    title: '',
    resource_type: 'pdf',
    sort_order: 1,
    object_key: '',
    mime_type: 'application/pdf',
    size_bytes: 0,
  })

  async function loadCatalog() {
    if (!courseId) return
    const chapterResp = await listCourseChapters(accessToken, courseId)
    setChapters(chapterResp.items)
    const pairs = await Promise.all(
      chapterResp.items.map(async (chapterItem) => {
        const res = await listChapterResources(accessToken, chapterItem.id)
        return [chapterItem.id, res.items] as const
      }),
    )
    setResourcesByChapter(Object.fromEntries(pairs))
  }

  useEffect(() => {
    if (!courseId) return
    void loadCatalog()
  }, [courseId])

  function inferResourceType(fileName: string): ResourceType {
    const ext = fileName.toLowerCase().split('.').pop() ?? ''
    if (ext === 'pdf') return 'pdf'
    if (ext === 'txt' || ext === 'md' || ext === 'markdown') return 'txt'
    if (ext === 'doc') return 'doc'
    if (ext === 'docx') return 'docx'
    if (ext === 'ppt' || ext === 'pptx') return 'ppt'
    if (['mp3', 'wav', 'ogg', 'oga', 'm4a', 'aac', 'flac', 'wma'].includes(ext)) return 'audio'
    if (['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) return 'video'
    return 'pdf'
  }

  function inferMimeType(file: File): string {
    if (file.type) return file.type
    const ext = file.name.toLowerCase().split('.').pop() ?? ''
    if (ext === 'pdf') return 'application/pdf'
    if (ext === 'txt') return 'text/plain; charset=utf-8'
    if (ext === 'md' || ext === 'markdown') return 'text/markdown; charset=utf-8'
    if (ext === 'ppt') return 'application/vnd.ms-powerpoint'
    if (ext === 'pptx') {
      return 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
    }
    if (ext === 'doc') return 'application/msword'
    if (ext === 'docx') {
      return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    }
    if (ext === 'mp3') return 'audio/mpeg'
    if (ext === 'wav') return 'audio/wav'
    if (ext === 'ogg' || ext === 'oga') return 'audio/ogg'
    if (ext === 'm4a') return 'audio/mp4'
    if (ext === 'aac') return 'audio/aac'
    if (ext === 'flac') return 'audio/flac'
    if (ext === 'wma') return 'audio/x-ms-wma'
    if (['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) return 'video/mp4'
    return 'application/octet-stream'
  }

  /** 下一可用资源排序号：表 chapter_resources 对 (chapter_id, sort_order) 唯一，不可固定为 1。 */
  function nextResourceSortOrder(chapterId: string): number {
    const list = resourcesByChapter[chapterId] ?? []
    let max = 0
    for (const r of list) {
      if (r.sort_order > max) max = r.sort_order
    }
    return max + 1
  }

  function resourceTagLabel(resourceItem: ResourceItem): string {
    const mime = (resourceItem.mime_type ?? '').toLowerCase()
    const url = (resourceItem.object_url ?? '').toLowerCase()

    // markdown / txt
    if (mime.includes('text/markdown') || url.endsWith('.md') || url.endsWith('.markdown')) return 'md'
    if (mime.startsWith('text/plain')) return 'txt'
    if (resourceItem.resource_type === 'txt') return 'txt'

    // office docs
    if (resourceItem.resource_type === 'pdf') return 'pdf'
    if (resourceItem.resource_type === 'doc') return 'doc'
    if (resourceItem.resource_type === 'docx') return 'docx'
    if (resourceItem.resource_type === 'ppt') {
      if (
        mime.includes('presentationml') ||
        mime.includes('openxmlformats-officedocument.presentationml.presentation') ||
        url.endsWith('.pptx')
      )
        return 'pptx'
      return 'ppt'
    }

    // video / audio: 优先取 mime 的 subtype（video/mp4 -> mp4），否则按扩展名兜底。
    if (resourceItem.resource_type === 'video') {
      const major = mime.split(';')[0]
      if (major.startsWith('video/')) return major.split('/')[1] || 'video'
      if (url.endsWith('.mp4')) return 'mp4'
      return 'video'
    }
    if (resourceItem.resource_type === 'audio') {
      const major = mime.split(';')[0]
      if (major.startsWith('audio/')) return major.split('/')[1] || 'audio'
      if (url.endsWith('.mp3')) return 'mp3'
      if (url.endsWith('.wav')) return 'wav'
      if (url.endsWith('.ogg') || url.endsWith('.oga')) return 'ogg'
      if (url.endsWith('.m4a')) return 'm4a'
      if (url.endsWith('.aac')) return 'aac'
      if (url.endsWith('.flac')) return 'flac'
      if (url.endsWith('.wma')) return 'wma'
      return 'audio'
    }

    return resourceItem.resource_type
  }

  function openPreview(resourceItem: ResourceItem) {
    const previewURL = resourceItem.preview_url || resourceItem.object_url
    const originalURL = resourceItem.object_url || previewURL
    if (!previewURL || !originalURL) {
      setError('该资源缺少访问地址，无法预览')
      return
    }
    const query = new URLSearchParams({
      title: resourceItem.title,
      resource_id: resourceItem.id,
      url: previewURL,
      original_url: originalURL,
      mime_type: resourceItem.mime_type ?? '',
      resource_type: resourceItem.resource_type,
    })
    navigate(`/teacher/resources/preview?${query.toString()}`)
  }

  async function handleCreateChapter(): Promise<boolean> {
    if (!courseId) return false
    setLoading(true)
    setError('')
    setMessage('')
    try {
      await createChapter(accessToken, courseId, chapterForm)
      setChapterForm({ title: '', sort_order: 1 })
      await loadCatalog()
      setMessage('章节创建成功')
      return true
    } catch (err) {
      setError(err instanceof Error ? err.message : '创建章节失败')
      return false
    } finally {
      setLoading(false)
    }
  }

  async function handleUploadResource(): Promise<boolean> {
    if (!courseId || !uploadChapterId || !resourceFile) return false
    setLoading(true)
    setError('')
    setMessage('')
    try {
      const placedOrder = resourceForm.sort_order
      const initData = await initUpload(accessToken, uploadChapterId, {
        course_id: Number(courseId),
        file_name: resourceFile.name,
        resource_type: resourceForm.resource_type,
      })
      const putResp = await fetch(initData.upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': inferMimeType(resourceFile) },
        body: resourceFile,
      })
      if (!putResp.ok) throw new Error(`文件上传失败（${putResp.status}）`)
      await confirmResource(accessToken, uploadChapterId, {
        ...resourceForm,
        object_key: initData.object_key,
        mime_type: inferMimeType(resourceFile),
        size_bytes: resourceFile.size,
      })
      await loadCatalog()
      setResourceFile(null)
      setResourceForm({
        title: '',
        resource_type: 'pdf',
        sort_order: placedOrder + 1,
        object_key: '',
        mime_type: 'application/pdf',
        size_bytes: 0,
      })
      setMessage('资源上传并入库成功')
      return true
    } catch (err) {
      setError(err instanceof Error ? err.message : '上传资源失败')
      return false
    } finally {
      setLoading(false)
    }
  }

  if (!courseId) {
    return (
      <div className="course-workspace">
        <div className="card">
          <h3>未选择课程</h3>
          <p className="subtle">请先从课程列表进入章节与资源管理。</p>
          <button onClick={() => navigate('/teacher/courses')}>返回课程列表</button>
        </div>
      </div>
    )
  }

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <h2>章节与资源管理</h2>
        <p className="subtle">当前课程 ID：{courseId}</p>
      </header>
      {message && <div className="alert success">{message}</div>}
      {error && <div className="alert error">{error}</div>}

      <section className="card catalog-panel">
        <div className="list-toolbar">
          <h3>章节目录</h3>
          <button className="primary" onClick={() => setShowChapterModal(true)}>
            + 新增章节
          </button>
        </div>
        {chapters.length === 0 && <p className="subtle">暂无章节</p>}
        {chapters.map((chapterItem) => (
          <section key={chapterItem.id} className="catalog-chapter">
            <div className="catalog-chapter-title">
              <span className="dot" />
              <strong>
                {chapterItem.sort_order}. {chapterItem.title}
              </strong>
              <div className="action-row">
                <button
                  onClick={() => {
                    setUploadChapterID(chapterItem.id)
                    const nextOrder = nextResourceSortOrder(chapterItem.id)
                    setResourceForm((v) => ({ ...v, sort_order: nextOrder }))
                    setShowResourceModal(true)
                  }}
                >
                  + 资源
                </button>
                <button
                  className="danger"
                  onClick={async () => {
                    if (
                      !window.confirm(
                        `确定删除章节「${chapterItem.title}」及其下资源？此操作不可恢复。`,
                      )
                    )
                      return
                    setError('')
                    try {
                      await deleteChapter(accessToken, chapterItem.id)
                      await loadCatalog()
                      setMessage('章节已删除')
                    } catch (err) {
                      setError(err instanceof Error ? err.message : '删除章节失败')
                    }
                  }}
                >
                  删除章节
                </button>
              </div>
            </div>
            <div className="catalog-resource-list">
              {(resourcesByChapter[chapterItem.id] ?? []).map((resourceItem) => (
                <div key={resourceItem.id} className="catalog-resource-row">
                  <span className="resource-tag">{resourceTagLabel(resourceItem)}</span>
                  <button className="catalog-resource-title" onClick={() => openPreview(resourceItem)}>
                    {resourceItem.title}
                  </button>
                  <button
                    className="danger"
                    onClick={async () => {
                      if (!window.confirm(`确定删除资源「${resourceItem.title}」？`)) return
                      setError('')
                      try {
                        await deleteResource(accessToken, resourceItem.id)
                        await loadCatalog()
                        setMessage('资源已删除')
                      } catch (err) {
                        setError(err instanceof Error ? err.message : '删除资源失败')
                      }
                    }}
                  >
                    删除
                  </button>
                </div>
              ))}
              {(resourcesByChapter[chapterItem.id] ?? []).length === 0 && (
                <p className="subtle">该章节暂无资源</p>
              )}
            </div>
          </section>
        ))}
      </section>

      {showChapterModal && (
        <div className="modal-mask" onClick={() => setShowChapterModal(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h3>新增章节</h3>
            <label>
              章节标题
              <input
                value={chapterForm.title}
                onChange={(e) => setChapterForm((v) => ({ ...v, title: e.target.value }))}
              />
            </label>
            <label>
              排序
              <input
                type="number"
                min={1}
                value={chapterForm.sort_order}
                onChange={(e) => setChapterForm((v) => ({ ...v, sort_order: Number(e.target.value) || 1 }))}
              />
            </label>
            <div className="action-row">
              <button onClick={() => setShowChapterModal(false)}>取消</button>
              <button
                className="primary"
                disabled={!chapterForm.title || loading}
                onClick={async () => {
                  const ok = await handleCreateChapter()
                  if (ok) setShowChapterModal(false)
                }}
              >
                确定
              </button>
            </div>
          </div>
        </div>
      )}

      {showResourceModal && (
        <div className="modal-mask" onClick={() => setShowResourceModal(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <h3>上传资源</h3>
            <label>
              文件
              <input
                type="file"
                accept=".txt,.pdf,.doc,.docx,.ppt,.pptx,.md,.markdown,.mp4,.mov,.mkv,.webm,.avi,.mp3,.wav,.ogg,.oga,.m4a,.aac,.flac,.wma,video/*,audio/*"
                onChange={(e) => {
                  const file = e.target.files?.[0] ?? null
                  setResourceFile(file)
                  if (!file) return
                  const title = file.name.replace(/\.[^.]+$/, '')
                  const resourceType = inferResourceType(file.name)
                  setResourceForm((v) => ({
                    ...v,
                    title,
                    resource_type: resourceType,
                    mime_type: inferMimeType(file),
                    size_bytes: file.size,
                  }))
                }}
              />
            </label>
            <label>
              资源标题
              <input
                value={resourceForm.title}
                onChange={(e) => setResourceForm((v) => ({ ...v, title: e.target.value }))}
              />
            </label>
            <label>
              排序（章节内唯一，默认已取下一个可用序号）
              <input
                type="number"
                min={1}
                value={resourceForm.sort_order}
                onChange={(e) =>
                  setResourceForm((v) => ({ ...v, sort_order: Number(e.target.value) || 1 }))
                }
              />
            </label>
            <div className="action-row">
              <button onClick={() => setShowResourceModal(false)}>取消</button>
              <button
                className="primary"
                disabled={!uploadChapterId || !resourceFile || !resourceForm.title || loading}
                onClick={async () => {
                  const ok = await handleUploadResource()
                  if (ok) setShowResourceModal(false)
                }}
              >
                上传
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
