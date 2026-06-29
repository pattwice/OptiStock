import { Navigate, Outlet } from 'react-router-dom'
import { Spin } from 'antd'
import { useAuthStore } from '../store/authStore'

export function ProtectedRoute() {
  const { accessToken, initialized } = useAuthStore()

  if (!initialized) {
    return (
      <div style={{ display: 'grid', placeItems: 'center', minHeight: '100vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!accessToken) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
