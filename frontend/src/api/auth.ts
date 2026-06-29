import { apiClient, type ApiEnvelope } from './client'
import type { AuthUser, LoginResponse } from '../types/auth'

export async function login(email: string, password: string) {
  const { data } = await apiClient.post<ApiEnvelope<LoginResponse>>('/auth/login', {
    email,
    password,
  })
  return data.data
}

export async function refreshSession() {
  const { data } = await apiClient.post<ApiEnvelope<LoginResponse>>('/auth/refresh')
  return data.data
}

export async function logout() {
  await apiClient.post('/auth/logout')
}

export async function fetchMe() {
  const { data } = await apiClient.get<ApiEnvelope<AuthUser>>('/me')
  return data.data
}
