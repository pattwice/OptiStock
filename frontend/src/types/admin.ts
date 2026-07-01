import type { UserRole } from './auth'

export interface AdminUser {
  id: string
  name: string
  email: string
  role: UserRole
  created_at: string
}

export interface SystemConfigEntry {
  config_key: string
  config_value: string
  updated_at: string
}

export interface CreateUserPayload {
  name: string
  email: string
  password: string
  role: UserRole
}

export interface UpdateUserPayload {
  name?: string
  email?: string
  password?: string
  role?: UserRole
}
