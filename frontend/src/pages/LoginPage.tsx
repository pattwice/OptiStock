import { useEffect, useState } from 'react'
import { Alert, Button, Card, Form, Input, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'
import { login } from '../api/auth'
import { useAuthStore } from '../store/authStore'

export function LoginPage() {
  const navigate = useNavigate()
  const accessToken = useAuthStore((state) => state.accessToken)
  const initialized = useAuthStore((state) => state.initialized)
  const setSession = useAuthStore((state) => state.setSession)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (initialized && accessToken) {
      navigate('/', { replace: true })
    }
  }, [accessToken, initialized, navigate])

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true)
    setError(null)
    try {
      const result = await login(values.email, values.password)
      setSession(result.access_token, result.user)
      navigate('/', { replace: true })
    } catch {
      setError('Invalid email or password.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        background: '#f5f5f5',
        padding: 16,
      }}
    >
      <Card style={{ width: 420 }}>
        <Typography.Title level={3}>OptiStock</Typography.Title>
        <Typography.Paragraph type="secondary">
          Sign in to manage inventory and production.
        </Typography.Paragraph>

        {error && <Alert type="error" message={error} showIcon style={{ marginBottom: 16 }} />}

        <Form layout="vertical" onFinish={onFinish} initialValues={{ email: 'admin@optistock.local' }}>
          <Form.Item
            label="Email"
            name="email"
            rules={[{ required: true, message: 'Email is required' }]}
          >
            <Input autoComplete="username" />
          </Form.Item>
          <Form.Item
            label="Password"
            name="password"
            rules={[{ required: true, message: 'Password is required' }]}
          >
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>
            Sign in
          </Button>
        </Form>
      </Card>
    </div>
  )
}
