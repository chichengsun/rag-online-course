import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { login as authLogin, logout as authLogout, saveTokens } from '../services/auth'
import type { LoginReq, User } from '../types'

/**
 * 认证状态接口
 */
interface AuthState {
  // 状态
  token: string | null
  user: User | null
  loading: boolean
  error: string | null

  // 方法
  isAuthenticated: () => boolean
  login: (account: string, password: string) => Promise<void>
  logout: () => void
  setAuth: (token: string, user: User) => void
  clearError: () => void
}

/**
 * 认证Store - 使用Zustand管理全局认证状态
 *
 * 使用方式:
 * ```tsx
 * import { useAuthStore } from '@/stores/authStore'
 *
 * // 在组件中
 * const { token, user, isAuthenticated, login, logout, loading, error } = useAuthStore()
 * ```
 *
 * 登录示例:
 * ```tsx
 * const { login, loading, error } = useAuthStore()
 *
 * const handleLogin = async () => {
 *   try {
 *     await login('user@example.com', 'password')
 *     // 登录成功后跳转
 *   } catch (e) {
 *     // error 已保存在 store 中
 *   }
 * }
 * ```
 */
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      loading: false,
      error: null,

      /**
       * 检查是否已认证
       */
      isAuthenticated: () => {
        return !!get().token
      },

      /**
       * 登录
       * @param account 账号（邮箱或用户名）
       * @param password 密码
       */
      login: async (account: string, password: string) => {
        set({ loading: true, error: null })

        try {
          const req: LoginReq = { account, password }
          const resp = await authLogin(req)

          // 保存 token 到 localStorage
          saveTokens(resp.access_token, resp.refresh_token)

          // 更新 store 状态
          set({
            token: resp.access_token,
            user: resp.user,
            loading: false,
            error: null,
          })
        } catch (err) {
          const message = err instanceof Error ? err.message : '登录失败，请检查账号密码'
          set({
            loading: false,
            error: message,
            token: null,
            user: null,
          })
          throw err
        }
      },

      /**
       * 登出
       */
      logout: () => {
        // 调用 authService 的 logout 清除 localStorage
        authLogout()

        // 清除 store 状态
        set({
          token: null,
          user: null,
          loading: false,
          error: null,
        })
      },

      /**
       * 设置认证信息（内部使用或token刷新时调用）
       */
      setAuth: (token: string, user: User) => {
        set({ token, user })
      },

      /**
       * 清除错误状态
       */
      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'rag_online_course_auth',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
      }),
    },
  ),
)
