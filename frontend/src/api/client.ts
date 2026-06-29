import axios from 'axios'
import { useAuthStore } from '../store/authStore'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

export const apiClient = axios.create({
  baseURL: `${baseURL}/api/v1`,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let refreshPromise: Promise<string | null> | null = null

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config
    if (error.response?.status !== 401 || original._retry) {
      return Promise.reject(error)
    }

    original._retry = true

    if (!refreshPromise) {
      refreshPromise = apiClient
        .post('/auth/refresh')
        .then((res) => {
          const token = res.data.data.access_token as string
          const user = res.data.data.user
          useAuthStore.getState().setSession(token, user)
          return token
        })
        .catch(() => {
          useAuthStore.getState().clearSession()
          return null
        })
        .finally(() => {
          refreshPromise = null
        })
    }

    const token = await refreshPromise
    if (!token) {
      return Promise.reject(error)
    }

    original.headers.Authorization = `Bearer ${token}`
    return apiClient(original)
  },
)

export interface ApiEnvelope<T> {
  success: boolean
  data: T
  error: {
    code: string
    message: string
    details?: unknown
  } | null
}
