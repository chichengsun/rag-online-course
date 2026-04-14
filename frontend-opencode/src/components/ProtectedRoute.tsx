import { Navigate } from 'react-router-dom'
import type { ReactElement } from 'react'
import { useAuth } from '@/store/auth'

/**
 * 受保护的路由组件
 * - 检查用户是否已认证（token）
 * - 未认证则重定向到 /auth
 */
function ProtectedRoute({ children }: { children: ReactElement }) {
  const { token } = useAuth()

  if (!token) {
    return <Navigate to="/auth" replace />
  }

  return children
}

export { ProtectedRoute }
