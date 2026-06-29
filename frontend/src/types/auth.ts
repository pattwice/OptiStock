export type UserRole = 'user' | 'supervisor' | 'admin'

export interface AuthUser {
  id: string
  name: string
  email: string
  role: UserRole
}

export interface LoginResponse {
  user: AuthUser
  access_token: string
  expires_in: number
}
