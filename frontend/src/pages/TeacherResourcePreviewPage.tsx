import { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { getResourcePreviewURL } from '../services/course'

function guessPreviewKind(resourceType: string, mimeType: string, objectURL: string) {
  const lowerMime = mimeType.toLowerCase()
  const lowerURL = objectURL.toLowerCase()
  if (resourceType === 'video' || lowerMime.startsWith('video/')) return 'video'
  if (resourceType === 'audio' || lowerMime.startsWith('audio/')) return 'audio'
  if (lowerMime.includes('pdf') || lowerURL.endsWith('.pdf')) return 'pdf'
  if (
    resourceType === 'txt' ||
    lowerMime.startsWith('text/') ||
    lowerURL.endsWith('.txt') ||
    lowerURL.endsWith('.md') ||
    lowerURL.endsWith('.markdown')
  ) {
    return 'text'
  }
  if (
    lowerMime.includes('presentation') ||
    lowerMime.includes('powerpoint') ||
    lowerURL.endsWith('.ppt') ||
    lowerURL.endsWith('.pptx')
  ) {
    return 'office'
  }
  if (
    resourceType === 'doc' ||
    resourceType === 'docx' ||
    lowerMime.includes('word') ||
    lowerMime.includes('msword') ||
    lowerURL.endsWith('.doc') ||
    lowerURL.endsWith('.docx')
  ) {
    return 'office'
  }
  return 'generic'
}

export function TeacherResourcePreviewPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const query = useMemo(() => new URLSearchParams(location.search), [location.search])
  const title = query.get('title') ?? '资源预览'
  const objectURL = query.get('url') ?? ''
  const originalURL = query.get('original_url') ?? objectURL
  const resourceId = query.get('resource_id') ?? ''
  const mimeType = query.get('mime_type') ?? ''
  const resourceType = query.get('resource_type') ?? ''
  const kind = guessPreviewKind(resourceType, mimeType, objectURL)
  const { token } = useAuth()
  const initialOfficeLoading = kind === 'office' && !!resourceId
  const [displayURL, setDisplayURL] = useState(() => (initialOfficeLoading ? '' : objectURL))
  const [officeLoading, setOfficeLoading] = useState(() => initialOfficeLoading)
  const [officeError, setOfficeError] = useState('')
  const [textContent, setTextContent] = useState('')
  const [textLoading, setTextLoading] = useState(false)
  const [textError, setTextError] = useState('')

  useEffect(() => {
    if (kind !== 'text') return
    const controller = new AbortController()
    async function loadText() {
      setTextLoading(true)
      setTextError('')
      try {
        const resp = await fetch(objectURL, { signal: controller.signal })
        if (!resp.ok) throw new Error(`文本加载失败（${resp.status}）`)
        const buf = await resp.arrayBuffer()
        const decoded = new TextDecoder('utf-8', { fatal: false }).decode(buf)
        setTextContent(decoded)
      } catch (err) {
        if (controller.signal.aborted) return
        setTextError(err instanceof Error ? err.message : '文本加载失败')
      } finally {
        if (!controller.signal.aborted) setTextLoading(false)
      }
    }
    void loadText()
    return () => controller.abort()
  }, [kind, objectURL])

  // Office 预览：服务端转换 PDF（预览用），确保页面内渲染可行。
  useEffect(() => {
    if (kind !== 'office') return
    if (!resourceId) return
    const tk = token ?? ''
    if (!tk) return
    let mounted = true
    setOfficeError('')
    setOfficeLoading(true)
    void getResourcePreviewURL(tk, resourceId)
      .then((resp) => {
        if (!mounted) return
        if (resp?.preview_url) setDisplayURL(resp.preview_url)
        else setDisplayURL('')
        setOfficeLoading(false)
      })
      .catch((err) => {
        if (!mounted) return
        setDisplayURL('')
        setOfficeError(err instanceof Error ? err.message : 'Office 预览转换失败')
        setOfficeLoading(false)
      })
    return () => {
      mounted = false
    }
  }, [kind, resourceId, token])

  if (!objectURL) {
    return (
      <div className="course-workspace">
        <section className="card resource-preview-page">
          <h2>资源预览</h2>
          <p className="subtle">缺少资源地址，请返回章节列表重新进入。</p>
          <div className="action-row">
            <button onClick={() => navigate(-1)}>返回上一页</button>
            <button className="primary" onClick={() => navigate('/teacher/courses')}>
              返回课程列表
            </button>
          </div>
        </section>
      </div>
    )
  }

  return (
    <div className="course-workspace">
      <section className="card resource-preview-page">
        <div className="resource-preview-header">
          <div>
            <h2>{title}</h2>
            <p className="subtle">类型：{mimeType || resourceType || '未知'}</p>
          </div>
          <div className="action-row">
            <button onClick={() => navigate(-1)}>返回</button>
            <a className="button-link primary-link" href={originalURL} target="_blank" rel="noreferrer">
              新窗口打开
            </a>
          </div>
        </div>

        {kind === 'video' && (
          <video className="resource-preview-frame" controls src={objectURL}>
            您的浏览器不支持视频播放，请使用“新窗口打开”查看。
          </video>
        )}

        {((kind === 'pdf' || kind === 'office' || kind === 'generic') && displayURL) && (
          <iframe className="resource-preview-frame" src={displayURL} title={title} />
        )}

        {kind === 'text' && (
          <div className="resource-text-viewer">
            {textLoading && <p className="subtle">文本加载中...</p>}
            {textError && <p className="alert error">{textError}</p>}
            {!textLoading && !textError && <pre>{textContent}</pre>}
          </div>
        )}

        {kind === 'office' && officeLoading && (
          <p className="subtle">正在将 Office 文档转换为 PDF 预览中...</p>
        )}

        {kind === 'office' && officeError && <p className="alert error">{officeError}</p>}

        {kind === 'office' && !officeLoading && (
          <p className="subtle">
            Office 文档在本地浏览器内的内嵌能力有限，可能会触发下载或无法完整渲染；如内嵌失败请点击「新窗口打开」使用本机 Office 打开。
          </p>
        )}

        {kind === 'audio' && (
          <audio className="resource-audio-frame" controls src={objectURL}>
            您的浏览器不支持音频播放。
          </audio>
        )}
      </section>
    </div>
  )
}
