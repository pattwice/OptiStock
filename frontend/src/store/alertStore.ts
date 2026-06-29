import { create } from 'zustand'
import type { AlertMessage } from '../types/alert'

interface AlertState {
  alerts: AlertMessage[]
  connected: boolean
  addAlert: (alert: Omit<AlertMessage, 'id' | 'read'>) => void
  markAllRead: () => void
  clearAlerts: () => void
  setConnected: (connected: boolean) => void
  unreadCount: () => number
}

export const useAlertStore = create<AlertState>((set, get) => ({
  alerts: [],
  connected: false,
  addAlert: (alert) =>
    set((state) => ({
      alerts: [
        {
          ...alert,
          id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
          read: false,
        },
        ...state.alerts,
      ].slice(0, 100),
    })),
  markAllRead: () =>
    set((state) => ({
      alerts: state.alerts.map((a) => ({ ...a, read: true })),
    })),
  clearAlerts: () => set({ alerts: [] }),
  setConnected: (connected) => set({ connected }),
  unreadCount: () => get().alerts.filter((a) => !a.read).length,
}))
