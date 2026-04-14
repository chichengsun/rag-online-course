import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

type AuthUser = {
  id: string
  username: string
  role: string
}

const TOKEN_KEY = 'rag_online_course_access_token'
const USER_KEY = 'rag_online_course_user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<AuthUser | null>(null)

  const rawUser = localStorage.getItem(USER_KEY)
  if (rawUser) {
    try {
      user.value = JSON.parse(rawUser) as AuthUser
    } catch {
      user.value = null
    }
  }

  const isAuthed = computed(() => !!token.value)

  function setAuth(nextToken: string, nextUser: AuthUser) {
    token.value = nextToken
    user.value = nextUser
    localStorage.setItem(TOKEN_KEY, nextToken)
    localStorage.setItem(USER_KEY, JSON.stringify(nextUser))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  return { token, user, isAuthed, setAuth, logout }
})
