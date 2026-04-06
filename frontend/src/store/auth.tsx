/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

type AuthUser = {
  id: string
  username: string
  role: string
}

type AuthState = {
  token: string | null
  user: AuthUser | null
  setAuth: (token: string, user: AuthUser) => void
  logout: () => void
}

const AuthContext = createContext<AuthState | undefined>(undefined)

const TOKEN_KEY = 'rag_online_course_access_token'
const USER_KEY = 'rag_online_course_user'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<AuthUser | null>(() => {
    const raw = localStorage.getItem(USER_KEY)
    if (!raw) return null
    try {
      return JSON.parse(raw) as AuthUser
    } catch {
      return null
    }
  })

  const value = useMemo<AuthState>(
    () => ({
      token,
      user,
      setAuth: (nextToken, nextUser) => {
        setToken(nextToken)
        setUser(nextUser)
        localStorage.setItem(TOKEN_KEY, nextToken)
        localStorage.setItem(USER_KEY, JSON.stringify(nextUser))
      },
      logout: () => {
        setToken(null)
        setUser(null)
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_KEY)
      },
    }),
    [token, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}
