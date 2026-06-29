import { useEffect, useRef } from 'react'
import { notification } from 'antd'
import { useAuthStore } from '../store/authStore'
import { useAlertStore } from '../store/alertStore'
import { alertDescription, alertTitle } from '../types/alert'

const RECONNECT_MS = 3000

export function useAlerts() {
  const accessToken = useAuthStore((s) => s.accessToken)
  const initialized = useAuthStore((s) => s.initialized)
  const addAlert = useAlertStore((s) => s.addAlert)
  const setConnected = useAlertStore((s) => s.setConnected)
  const socketRef = useRef<WebSocket | null>(null)
  const timerRef = useRef<number | null>(null)

  useEffect(() => {
    if (!initialized || !accessToken) {
      setConnected(false)
      socketRef.current?.close()
      socketRef.current = null
      return
    }

    let cancelled = false

    const connect = () => {
      if (cancelled) return
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const url = `${proto}//${window.location.host}/api/v1/ws?token=${encodeURIComponent(accessToken)}`
      const ws = new WebSocket(url)
      socketRef.current = ws

      ws.onopen = () => setConnected(true)
      ws.onclose = () => {
        setConnected(false)
        if (!cancelled) {
          timerRef.current = window.setTimeout(connect, RECONNECT_MS)
        }
      }
      ws.onerror = () => ws.close()
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data as string) as {
            type: string
            severity: string
            payload: Record<string, unknown>
            timestamp: string
          }
          addAlert(data)
          notification.open({
            message: alertTitle({ ...data, id: '', read: false }),
            description: alertDescription({ ...data, id: '', read: false }),
            placement: 'bottomRight',
          })
        } catch {
          // ignore malformed payloads
        }
      }
    }

    connect()

    return () => {
      cancelled = true
      if (timerRef.current != null) {
        window.clearTimeout(timerRef.current)
      }
      socketRef.current?.close()
      socketRef.current = null
      setConnected(false)
    }
  }, [accessToken, addAlert, initialized, setConnected])
}
