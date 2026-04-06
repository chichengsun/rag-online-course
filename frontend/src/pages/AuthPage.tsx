import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login, register, type RegisterPayload } from '../services/auth'
import { useAuth } from '../store/auth'

type Mode = 'login' | 'register'

const initialRegister: RegisterPayload = {
  email: '',
  username: '',
  name: '',
  password: '',
  role: 'teacher',
}

export function AuthPage() {
  const [mode, setMode] = useState<Mode>('login')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [loginForm, setLoginForm] = useState({ account: '', password: '' })
  const [registerForm, setRegisterForm] = useState<RegisterPayload>(initialRegister)
  const navigate = useNavigate()
  const { setAuth } = useAuth()

  const title = useMemo(
    () => (mode === 'login' ? '欢迎回来' : '创建教师账号'),
    [mode],
  )

  async function onSubmitLogin(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const data = await login(loginForm)
      setAuth(data.access_token, data.user)
      navigate('/teacher/courses', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  async function onSubmitRegister(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await register(registerForm)
      setMode('login')
      setLoginForm({ account: registerForm.username, password: registerForm.password })
    } catch (err) {
      setError(err instanceof Error ? err.message : '注册失败，请检查输入')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="auth-layout">
      <section className="brand-pane">
        <h1>RAG Online Course</h1>
        <p>教师课程管理台</p>
        <p className="subtle">支持课程、章节、资源流程一体化管理。</p>
      </section>
      <section className="form-pane card">
        <div className="mode-switch">
          <button
            className={mode === 'login' ? 'active' : ''}
            onClick={() => setMode('login')}
            type="button"
          >
            登录
          </button>
          <button
            className={mode === 'register' ? 'active' : ''}
            onClick={() => setMode('register')}
            type="button"
          >
            注册
          </button>
        </div>
        <h2>{title}</h2>
        {error && <div className="alert error">{error}</div>}

        {mode === 'login' ? (
          <form onSubmit={onSubmitLogin} className="stack">
            <label>
              账号（用户名或邮箱）
              <input
                required
                value={loginForm.account}
                onChange={(e) => setLoginForm((v) => ({ ...v, account: e.target.value }))}
              />
            </label>
            <label>
              密码
              <input
                required
                type="password"
                value={loginForm.password}
                onChange={(e) => setLoginForm((v) => ({ ...v, password: e.target.value }))}
              />
            </label>
            <button disabled={loading} type="submit" className="primary">
              {loading ? '登录中...' : '登录'}
            </button>
          </form>
        ) : (
          <form onSubmit={onSubmitRegister} className="stack">
            <label>
              邮箱
              <input
                required
                type="email"
                value={registerForm.email}
                onChange={(e) => setRegisterForm((v) => ({ ...v, email: e.target.value }))}
              />
            </label>
            <label>
              用户名
              <input
                required
                value={registerForm.username}
                onChange={(e) => setRegisterForm((v) => ({ ...v, username: e.target.value }))}
              />
            </label>
            <label>
              姓名
              <input
                required
                value={registerForm.name}
                onChange={(e) => setRegisterForm((v) => ({ ...v, name: e.target.value }))}
              />
            </label>
            <label>
              密码
              <input
                required
                minLength={6}
                type="password"
                value={registerForm.password}
                onChange={(e) => setRegisterForm((v) => ({ ...v, password: e.target.value }))}
              />
            </label>
            <button disabled={loading} type="submit" className="primary">
              {loading ? '注册中...' : '注册'}
            </button>
          </form>
        )}
      </section>
    </main>
  )
}
