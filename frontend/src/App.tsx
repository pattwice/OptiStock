import { useEffect } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './components/AppLayout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { AdjustmentsPage } from './pages/inventory/AdjustmentsPage'
import { BomPage } from './pages/inventory/BomPage'
import { ItemsPage } from './pages/inventory/ItemsPage'
import { LotsPage } from './pages/inventory/LotsPage'
import { ReceivingPage } from './pages/inventory/ReceivingPage'
import { StockOnHandPage } from './pages/inventory/StockOnHandPage'
import { WorkOrderCreatePage } from './pages/workorders/WorkOrderCreatePage'
import { WorkOrderDetailPage } from './pages/workorders/WorkOrderDetailPage'
import { WorkOrdersPage } from './pages/workorders/WorkOrdersPage'
import { ApprovalsPage } from './pages/approvals/ApprovalsPage'
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
            <Route path="/inventory/items" element={<ItemsPage />} />
            <Route path="/inventory/bom" element={<BomPage />} />
            <Route path="/inventory/lots" element={<LotsPage />} />
            <Route path="/inventory/receiving" element={<ReceivingPage />} />
            <Route path="/inventory/adjustments" element={<AdjustmentsPage />} />
            <Route path="/inventory/stock" element={<StockOnHandPage />} />
            <Route path="/workorders" element={<WorkOrdersPage />} />
            <Route path="/workorders/new" element={<WorkOrderCreatePage />} />
            <Route path="/workorders/:woNumber" element={<WorkOrderDetailPage />} />
            <Route path="/approvals" element={<ApprovalsPage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
