import { request } from './api'

export type RegisterPayload = {
  email: string
  username: string
  name: string
  password: string
  role: 'teacher' | 'student'
}

export type LoginPayload = {
  account: string
  password: string
}

type LoginData = {
  access_token: string
  refresh_token: string
  user: {
    id: string
    username: string
    role: string
  }
}

export async function register(payload: RegisterPayload) {
  return request<{ id: string }>('/auth/register', {
    method: 'POST',
    body: payload,
  })
}

export async function login(payload: LoginPayload) {
  return request<LoginData>('/auth/login', {
    method: 'POST',
    body: payload,
  })
}
