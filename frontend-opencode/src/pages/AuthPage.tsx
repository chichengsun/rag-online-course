import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { register } from '@/services/auth'
import type { UserRole } from '@/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type Mode = 'login' | 'register'

interface LoginForm {
  account: string
  password: string
}

interface RegisterForm {
  email: string
  username: string
  name: string
  password: string
  role: UserRole
}

/**
 * 认证页面 - 登录和注册
 * 
 * 功能：
 * - 登录表单：账号（邮箱或用户名）+ 密码
 * - 注册表单：邮箱 + 用户名 + 姓名 + 密码 + 角色选择
 * - 表单校验：必填字段、邮箱格式
 * - 错误提示：API 错误显示
 * - 登录成功后根据角色跳转到对应布局
 */
export function AuthPage() {
  const [mode, setMode] = useState<Mode>('login')
  const [loginForm, setLoginForm] = useState<LoginForm>({
    account: '',
    password: '',
  })
  const [registerForm, setRegisterForm] = useState<RegisterForm>({
    email: '',
    username: '',
    name: '',
    password: '',
    role: 'teacher',
  })
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({})

  const navigate = useNavigate()
  const { login, loading, error, clearError } = useAuthStore()

  const title = useMemo(
    () => (mode === 'login' ? '欢迎回来' : '创建账号'),
    [mode],
  )

  const subtitle = useMemo(
    () => (mode === 'login' ? '登录您的账号以继续' : '填写信息创建新账号'),
    [mode],
  )

  /**
   * 验证邮箱格式
   */
  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(email)
  }

  /**
   * 验证登录表单
   */
  const validateLoginForm = (): boolean => {
    const errors: Record<string, string> = {}

    if (!loginForm.account.trim()) {
      errors.account = '请输入账号'
    }

    if (!loginForm.password.trim()) {
      errors.password = '请输入密码'
    }

    setValidationErrors(errors)
    return Object.keys(errors).length === 0
  }

  /**
   * 验证注册表单
   */
  const validateRegisterForm = (): boolean => {
    const errors: Record<string, string> = {}

    if (!registerForm.email.trim()) {
      errors.email = '请输入邮箱'
    } else if (!validateEmail(registerForm.email)) {
      errors.email = '请输入有效的邮箱地址'
    }

    if (!registerForm.username.trim()) {
      errors.username = '请输入用户名'
    }

    if (!registerForm.name.trim()) {
      errors.name = '请输入姓名'
    }

    if (!registerForm.password.trim()) {
      errors.password = '请输入密码'
    } else if (registerForm.password.length < 6) {
      errors.password = '密码至少需要6个字符'
    }

    setValidationErrors(errors)
    return Object.keys(errors).length === 0
  }

  /**
   * 处理登录提交
   */
  const handleSubmitLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()

    if (!validateLoginForm()) {
      return
    }

    try {
      await login(loginForm.account, loginForm.password)
      
      // 登录成功后根据角色跳转
      const user = useAuthStore.getState().user
      if (user?.role === 'teacher') {
        navigate('/teacher/courses', { replace: true })
      } else {
        navigate('/student/courses', { replace: true })
      }
    } catch (err) {
      // 错误已在 store 中处理
      console.error('登录失败:', err)
    }
  }

  /**
   * 处理注册提交
   */
  const handleSubmitRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()

    if (!validateRegisterForm()) {
      return
    }

    try {
      await register(registerForm)
      
      // 注册成功后切换到登录模式并预填表单
      setMode('login')
      setLoginForm({
        account: registerForm.username,
        password: registerForm.password,
      })
      
      // 清空注册表单
      setRegisterForm({
        email: '',
        username: '',
        name: '',
        password: '',
        role: 'teacher',
      })
    } catch (err) {
      // 错误已在 store 中处理
      console.error('注册失败:', err)
    }
  }

  /**
   * 切换模式时清除错误和表单
   */
  const handleSwitchMode = (newMode: Mode) => {
    setMode(newMode)
    clearError()
    setValidationErrors({})
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted/30 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">
            RAG Online Course
          </h1>
          <p className="text-muted-foreground">
            在线课程管理平台
          </p>
        </div>

        <Card className="shadow-lg">
          <CardHeader className="space-y-1 pb-4">
            <div className="flex gap-1 p-1 bg-muted/50 rounded-lg mb-2">
              <button
                type="button"
                onClick={() => handleSwitchMode('login')}
                className={`flex-1 py-2 px-4 text-sm font-medium rounded-md transition-all ${
                  mode === 'login'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                登录
              </button>
              <button
                type="button"
                onClick={() => handleSwitchMode('register')}
                className={`flex-1 py-2 px-4 text-sm font-medium rounded-md transition-all ${
                  mode === 'register'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                注册
              </button>
            </div>
            
            <CardTitle className="text-2xl">{title}</CardTitle>
            <p className="text-sm text-muted-foreground">{subtitle}</p>
          </CardHeader>

          <CardContent>
            {error && (
              <div className="mb-4 p-3 rounded-lg bg-destructive/10 border border-destructive/20">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            )}

            {mode === 'login' ? (
              <form onSubmit={handleSubmitLogin} className="space-y-4">
                <div className="space-y-2">
                  <label htmlFor="login-account" className="text-sm font-medium text-foreground">
                    账号
                  </label>
                  <Input
                    id="login-account"
                    type="text"
                    placeholder="用户名或邮箱"
                    value={loginForm.account}
                    onChange={(e) => setLoginForm({ ...loginForm, account: e.target.value })}
                    disabled={loading}
                    className={validationErrors.account ? 'border-destructive' : ''}
                  />
                  {validationErrors.account && (
                    <p className="text-xs text-destructive">{validationErrors.account}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <label htmlFor="login-password" className="text-sm font-medium text-foreground">
                    密码
                  </label>
                  <Input
                    id="login-password"
                    type="password"
                    placeholder="输入密码"
                    value={loginForm.password}
                    onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })}
                    disabled={loading}
                    className={validationErrors.password ? 'border-destructive' : ''}
                  />
                  {validationErrors.password && (
                    <p className="text-xs text-destructive">{validationErrors.password}</p>
                  )}
                </div>

                <Button
                  type="submit"
                  disabled={loading}
                  className="w-full"
                >
                  {loading ? '登录中...' : '登录'}
                </Button>
              </form>
            ) : (
              <form onSubmit={handleSubmitRegister} className="space-y-4">
                <div className="space-y-2">
                  <label htmlFor="register-email" className="text-sm font-medium text-foreground">
                    邮箱
                  </label>
                  <Input
                    id="register-email"
                    type="email"
                    placeholder="your@email.com"
                    value={registerForm.email}
                    onChange={(e) => setRegisterForm({ ...registerForm, email: e.target.value })}
                    disabled={loading}
                    className={validationErrors.email ? 'border-destructive' : ''}
                  />
                  {validationErrors.email && (
                    <p className="text-xs text-destructive">{validationErrors.email}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <label htmlFor="register-username" className="text-sm font-medium text-foreground">
                    用户名
                  </label>
                  <Input
                    id="register-username"
                    type="text"
                    placeholder="用户名"
                    value={registerForm.username}
                    onChange={(e) => setRegisterForm({ ...registerForm, username: e.target.value })}
                    disabled={loading}
                    className={validationErrors.username ? 'border-destructive' : ''}
                  />
                  {validationErrors.username && (
                    <p className="text-xs text-destructive">{validationErrors.username}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <label htmlFor="register-name" className="text-sm font-medium text-foreground">
                    姓名
                  </label>
                  <Input
                    id="register-name"
                    type="text"
                    placeholder="您的姓名"
                    value={registerForm.name}
                    onChange={(e) => setRegisterForm({ ...registerForm, name: e.target.value })}
                    disabled={loading}
                    className={validationErrors.name ? 'border-destructive' : ''}
                  />
                  {validationErrors.name && (
                    <p className="text-xs text-destructive">{validationErrors.name}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <label htmlFor="register-password" className="text-sm font-medium text-foreground">
                    密码
                  </label>
                  <Input
                    id="register-password"
                    type="password"
                    placeholder="至少6个字符"
                    value={registerForm.password}
                    onChange={(e) => setRegisterForm({ ...registerForm, password: e.target.value })}
                    disabled={loading}
                    className={validationErrors.password ? 'border-destructive' : ''}
                  />
                  {validationErrors.password && (
                    <p className="text-xs text-destructive">{validationErrors.password}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <label htmlFor="register-role" className="text-sm font-medium text-foreground">
                    角色
                  </label>
                  <Select
                    value={registerForm.role}
                    onValueChange={(value) => 
                      setRegisterForm({ ...registerForm, role: value as UserRole })
                    }
                    disabled={loading}
                  >
                    <SelectTrigger id="register-role" className="w-full">
                      <SelectValue placeholder="选择角色" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="teacher">教师</SelectItem>
                      <SelectItem value="student">学生</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <Button
                  type="submit"
                  disabled={loading}
                  className="w-full"
                >
                  {loading ? '注册中...' : '注册'}
                </Button>
              </form>
            )}
          </CardContent>
        </Card>

        <p className="text-center text-sm text-muted-foreground mt-6">
          {mode === 'login' ? (
            <>
              还没有账号？{' '}
              <button
                type="button"
                onClick={() => handleSwitchMode('register')}
                className="text-primary hover:underline font-medium"
              >
                立即注册
              </button>
            </>
          ) : (
            <>
              已有账号？{' '}
              <button
                type="button"
                onClick={() => handleSwitchMode('login')}
                className="text-primary hover:underline font-medium"
              >
                立即登录
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  )
}