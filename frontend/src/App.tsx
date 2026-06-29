import { useEffect } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './components/AppLayout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { refreshSession } from './api/auth'
import { useAuthStore } from './store/authStore'

function BootstrapAuth() {
  const setSession = useAuthStore((state) => state.setSession)
  const clearSession = useAuthStore((state) => state.clearSession)
  const setInitialized = useAuthStore((state) => state.setInitialized)

  useEffect(() => {
    let active = true

    refreshSession()
      .then((result) => {
        if (active) {
          setSession(result.access_token, result.user)
        }
      })
      .catch(() => {
        if (active) {
          clearSession()
        }
      })
      .finally(() => {
        if (active) {
          setInitialized(true)
        }
      })

    return () => {
      active = false
    }
  }, [clearSession, setInitialized, setSession])

  return null
}

export default function App() {
  return (
    <BrowserRouter>
      <BootstrapAuth />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route element={<ProtectedRoute />}>
          <Route element={<AppLayout />}>
            <Route path="/" element={<DashboardPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
