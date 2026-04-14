import { request } from './api'
import type { LoginReq, LoginResp, RegisterReq, RefreshReq, RefreshResp } from '../types'

const TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'

export async function register(data: RegisterReq): Promise<LoginResp> {
  return request<LoginResp>('/auth/register', {
    method: 'POST',
    body: data,
  })
}

export async function login(data: LoginReq): Promise<LoginResp> {
  return request<LoginResp>('/auth/login', {
    method: 'POST',
    body: data,
  })
}

export async function refreshToken(refresh_token: string): Promise<LoginResp> {
  const resp = await request<RefreshResp>('/auth/refresh', {
    method: 'POST',
    body: { refresh_token } as RefreshReq,
  })

  return {
    access_token: resp.access_token,
    refresh_token: refresh_token,
    user: resp as unknown as LoginResp['user'],
  } as LoginResp
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function saveTokens(access_token: string, refresh_token: string): void {
  localStorage.setItem(TOKEN_KEY, access_token)
  localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token)
}