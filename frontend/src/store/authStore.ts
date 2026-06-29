import { create } from 'zustand'
import type { AuthUser } from '../types/auth'

interface AuthState {
  accessToken: string | null
  user: AuthUser | null
  initialized: boolean
  setSession: (accessToken: string, user: AuthUser) => void
  clearSession: () => void
  setInitialized: (initialized: boolean) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  user: null,
  initialized: false,
  setSession: (accessToken, user) => set({ accessToken, user }),
  clearSession: () => set({ accessToken: null, user: null }),
  setInitialized: (initialized) => set({ initialized }),
}))
