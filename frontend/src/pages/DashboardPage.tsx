import { Card, Col, Row, Statistic, Typography } from 'antd'
import { useAuthStore } from '../store/authStore'

export function DashboardPage() {
  const user = useAuthStore((state) => state.user)

  return (
    <div>
      <Typography.Title level={3}>Dashboard</Typography.Title>
      <Typography.Paragraph type="secondary">
        Welcome back, {user?.name}. Phase 0 foundation is running.
      </Typography.Paragraph>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={8}>
          <Card>
            <Statistic title="Signed in as" value={user?.role ?? '-'} />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card>
            <Statistic title="Inventory modules" value="Phase 1" />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card>
            <Statistic title="Work orders" value="Phase 2" />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
