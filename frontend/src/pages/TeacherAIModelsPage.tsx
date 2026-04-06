import { useEffect, useState } from 'react'
import {
  createAIModel,
  deleteAIModel,
  listAIModels,
  testAIModelConnection,
  updateAIModel,
  type AIModelItem,
  type AIModelType,
} from '../services/aiModels'
import { useAuth } from '../store/auth'

const MODEL_TYPES: { value: AIModelType; label: string }[] = [
  { value: 'qa', label: '问答' },
  { value: 'embedding', label: '嵌入' },
  { value: 'rerank', label: '重排' },
]

// TeacherAIModelsPage 维护教师的问答 / 嵌入 / 重排模型配置（持久化到服务端）。
export function TeacherAIModelsPage() {
  const { token } = useAuth()
  const accessToken = token ?? ''
  const [items, setItems] = useState<AIModelItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<AIModelItem | null>(null)
  const [form, setForm] = useState({
    name: '',
    model_type: 'embedding' as AIModelType,
    api_base_url: '',
    model_id: '',
    api_key: '',
  })
  const [testBusy, setTestBusy] = useState(false)
  const [testHint, setTestHint] = useState<{ ok: boolean; text: string } | null>(null)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const { items: rows } = await listAIModels(accessToken)
      setItems(rows)
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [accessToken])

  function openCreate() {
    setEditing(null)
    setTestHint(null)
    setForm({
      name: '',
      model_type: 'embedding',
      api_base_url: 'https://api.openai.com/v1/embeddings',
      model_id: 'text-embedding-3-small',
      api_key: '',
    })
    setShowModal(true)
  }

  function openEdit(m: AIModelItem) {
    setEditing(m)
    setTestHint(null)
    setForm({
      name: m.name,
      model_type: m.model_type,
      api_base_url: m.api_base_url,
      model_id: m.model_id,
      api_key: '',
    })
    setShowModal(true)
  }

  const canTestConnection =
    Boolean(form.api_base_url.trim() && form.model_id.trim()) &&
    (Boolean(form.api_key.trim()) || Boolean(editing?.has_api_key))

  async function runTestConnection() {
    setError('')
    setTestHint(null)
    setTestBusy(true)
    try {
      const data = await testAIModelConnection(accessToken, {
        model_type: form.model_type,
        api_base_url: form.api_base_url.trim(),
        model_id: form.model_id.trim(),
        api_key: form.api_key.trim() || undefined,
        existing_model_id:
          editing && !form.api_key.trim() && editing.has_api_key ? editing.id : undefined,
      })
      setTestHint({ ok: data.ok, text: data.message + (data.http_status ? `（HTTP ${data.http_status}）` : '') })
    } catch (e) {
      setError(e instanceof Error ? e.message : '测试失败')
    } finally {
      setTestBusy(false)
    }
  }

  async function submitModal() {
    setError('')
    setMessage('')
    setLoading(true)
    try {
      if (editing) {
        await updateAIModel(accessToken, editing.id, {
          name: form.name,
          api_base_url: form.api_base_url,
          model_id: form.model_id,
          api_key: form.api_key,
        })
        setMessage('模型已更新（未填写 API Key 则保留原密钥）')
      } else {
        await createAIModel(accessToken, {
          name: form.name,
          model_type: form.model_type,
          api_base_url: form.api_base_url,
          model_id: form.model_id,
          api_key: form.api_key,
        })
        setMessage('模型已添加')
      }
      setShowModal(false)
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="course-workspace">
      <header className="card workspace-header">
        <div className="action-row" style={{ justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
          <div>
            <h2>模型管理</h2>
            <p className="subtle">
              配置问答、嵌入、重排模型；保存前可用「测试连接」做最小请求校验（三种类型请求体不同，见表单说明）。
            </p>
          </div>
          <button className="primary" onClick={openCreate}>
            添加模型
          </button>
        </div>
      </header>
      {error && <div className="alert error">{error}</div>}
      {message && <div className="alert success">{message}</div>}

      <section className="card stack">
        {loading && !items.length && <p className="subtle">加载中…</p>}
        <div className="course-list">
          {items.map((m) => (
            <article key={m.id} className="course-item">
              <div>
                <h4>{m.name}</h4>
                <p className="mono subtle">
                  {MODEL_TYPES.find((t) => t.value === m.model_type)?.label ?? m.model_type} · {m.model_id}
                </p>
                <p className="subtle" style={{ fontSize: 13, wordBreak: 'break-all' }}>
                  {m.api_base_url}
                </p>
                <p className="subtle">API Key：{m.has_api_key ? '已配置' : '未配置'}</p>
              </div>
              <div className="action-row">
                <button onClick={() => openEdit(m)}>编辑</button>
                <button
                  className="danger"
                  onClick={async () => {
                    if (!window.confirm(`删除模型「${m.name}」？`)) return
                    try {
                      await deleteAIModel(accessToken, m.id)
                      setMessage('已删除')
                      await load()
                    } catch (e) {
                      setError(e instanceof Error ? e.message : '删除失败')
                    }
                  }}
                >
                  删除
                </button>
              </div>
            </article>
          ))}
        </div>
        {!loading && items.length === 0 && <p className="subtle">暂无模型，请点击「添加模型」。</p>}
      </section>

      {showModal && (
        // 不在遮罩上绑定点击关闭：避免误触空白处、下拉框收起、触控漂移导致表单被关掉。
        <div className="modal-mask" role="presentation">
          <div
            className="modal-card"
            role="dialog"
            aria-modal="true"
            aria-labelledby="ai-model-modal-title"
            style={{ maxWidth: 520 }}
          >
            <div className="modal-card-header">
              <h3 id="ai-model-modal-title">{editing ? '编辑模型' : '添加模型'}</h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label="关闭"
                onClick={() => setShowModal(false)}
              >
                ×
              </button>
            </div>
            <label>
              显示名称
              <input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
            </label>
            {!editing && (
              <label>
                类型
                <select
                  value={form.model_type}
                  onChange={(e) => setForm((f) => ({ ...f, model_type: e.target.value as AIModelType }))}
                >
                  {MODEL_TYPES.map((t) => (
                    <option key={t.value} value={t.value}>
                      {t.label}
                    </option>
                  ))}
                </select>
              </label>
            )}
            <label>
              API 地址（须为完整 URL）
              <input
                value={form.api_base_url}
                onChange={(e) => setForm((f) => ({ ...f, api_base_url: e.target.value }))}
              />
              <span className="subtle" style={{ fontSize: 12, marginTop: 4 }}>
                问答：<code>/v1/chat/completions</code> · 嵌入：<code>/v1/embeddings</code> ·
                重排：常见 <code>/v1/rerank</code>（请求体含 model、query、documents、top_n）
              </span>
            </label>
            <label>
              模型 ID（请求体 model 字段）
              <input value={form.model_id} onChange={(e) => setForm((f) => ({ ...f, model_id: e.target.value }))} />
            </label>
            <label>
              API Key{editing && '（留空表示不修改；测试连接可使用已存密钥）'}
              <input
                type="password"
                autoComplete="off"
                value={form.api_key}
                onChange={(e) => setForm((f) => ({ ...f, api_key: e.target.value }))}
              />
            </label>
            {testHint && (
              <div className={testHint.ok ? 'alert success' : 'alert error'} style={{ marginTop: 8 }}>
                {testHint.text}
              </div>
            )}
            <p className="subtle" style={{ fontSize: 12, margin: 0 }}>
              点击背景不会关闭；请用「取消」或右上角 ×，避免误触丢失已填内容。
            </p>
            <div className="action-row">
              <button type="button" onClick={() => setShowModal(false)}>
                取消
              </button>
              <button type="button" disabled={testBusy || !canTestConnection} onClick={() => void runTestConnection()}>
                {testBusy ? '测试中…' : '测试连接'}
              </button>
              <button
                className="primary"
                disabled={loading || !form.name || !form.api_base_url || !form.model_id || (!editing && !form.api_key)}
                onClick={() => void submitModal()}
              >
                保存
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
