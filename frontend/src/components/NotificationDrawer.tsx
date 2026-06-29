import { Button, Drawer, List, Space, Tag, Typography } from 'antd'
import { useAlertStore } from '../store/alertStore'
import { alertDescription, alertTitle } from '../types/alert'

interface NotificationDrawerProps {
  open: boolean
  onClose: () => void
}

export function NotificationDrawer({ open, onClose }: NotificationDrawerProps) {
  const alerts = useAlertStore((s) => s.alerts)
  const markAllRead = useAlertStore((s) => s.markAllRead)
  const clearAlerts = useAlertStore((s) => s.clearAlerts)
  const connected = useAlertStore((s) => s.connected)

  return (
    <Drawer
      title="Notifications"
      open={open}
      onClose={onClose}
      extra={
        <Space>
          <Tag color={connected ? 'green' : 'default'}>{connected ? 'Live' : 'Offline'}</Tag>
          <Button size="small" onClick={markAllRead}>
            Mark read
          </Button>
          <Button size="small" onClick={clearAlerts}>
            Clear
          </Button>
        </Space>
      }
    >
      {alerts.length === 0 ? (
        <Typography.Text type="secondary">No alerts yet.</Typography.Text>
      ) : (
        <List
          dataSource={alerts}
          renderItem={(item) => (
            <List.Item style={{ opacity: item.read ? 0.65 : 1 }}>
              <List.Item.Meta
                title={alertTitle(item)}
                description={
                  <>
                    <div>{alertDescription(item)}</div>
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      {new Date(item.timestamp).toLocaleString()}
                    </Typography.Text>
                  </>
                }
              />
            </List.Item>
          )}
        />
      )}
    </Drawer>
  )
}
