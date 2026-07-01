import { apiClient, type ApiEnvelope } from './client'
import type { AdminUser, CreateUserPayload, SystemConfigEntry, UpdateUserPayload } from '../types/admin'

function unwrap<T>(res: { data: ApiEnvelope<T> }): T {
  return res.data.data
}

export async function listAdminUsers(): Promise<AdminUser[]> {
  const res = await apiClient.get<ApiEnvelope<AdminUser[]>>('/admin/users')
  return unwrap(res)
}

export async function createAdminUser(payload: CreateUserPayload): Promise<AdminUser> {
  const res = await apiClient.post<ApiEnvelope<AdminUser>>('/admin/users', payload)
  return unwrap(res)
}

export async function updateAdminUser(userId: string, payload: UpdateUserPayload): Promise<AdminUser> {
  const res = await apiClient.patch<ApiEnvelope<AdminUser>>(`/admin/users/${userId}`, payload)
  return unwrap(res)
}

export async function listSystemConfig(): Promise<SystemConfigEntry[]> {
  const res = await apiClient.get<ApiEnvelope<SystemConfigEntry[]>>('/admin/config')
  return unwrap(res)
}

export async function patchSystemConfig(updates: Record<string, string>): Promise<SystemConfigEntry[]> {
  const res = await apiClient.patch<ApiEnvelope<SystemConfigEntry[]>>('/admin/config', updates)
  return unwrap(res)
}
