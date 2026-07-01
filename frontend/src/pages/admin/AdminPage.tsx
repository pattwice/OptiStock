import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Typography,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  createAdminUser,
  listAdminUsers,
  listSystemConfig,
  patchSystemConfig,
  updateAdminUser,
} from '../../api/admin'
import type { AdminUser, SystemConfigEntry } from '../../types/admin'
import type { UserRole } from '../../types/auth'

const ROLES: UserRole[] = ['user', 'supervisor', 'admin']

export function AdminPage() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [config, setConfig] = useState<SystemConfigEntry[]>([])
  const [loadingUsers, setLoadingUsers] = useState(false)
  const [loadingConfig, setLoadingConfig] = useState(false)
  const [userModalOpen, setUserModalOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null)
  const [userForm] = Form.useForm()
  const [configForm] = Form.useForm()

  const loadUsers = useCallback(async () => {
    setLoadingUsers(true)
    try {
      setUsers(await listAdminUsers())
    } catch {
      message.error('Failed to load users')
    } finally {
      setLoadingUsers(false)
    }
  }, [])

  const loadConfig = useCallback(async () => {
    setLoadingConfig(true)
    try {
      const entries = await listSystemConfig()
      setConfig(entries)
      const values: Record<string, string | number | boolean> = {}
      for (const entry of entries) {
        if (entry.config_key.endsWith('_ENABLED')) {
          values[entry.config_key] = entry.config_value.toUpperCase() === 'TRUE'
        } else if (entry.config_key === 'NEAR_EXPIRY_DAYS_DEFAULT') {
          values[entry.config_key] = Number(entry.config_value)
        } else {
          values[entry.config_key] = entry.config_value
        }
      }
      configForm.setFieldsValue(values)
    } catch {
      message.error('Failed to load system config')
    } finally {
      setLoadingConfig(false)
    }
  }, [configForm])

  useEffect(() => {
    void loadUsers()
    void loadConfig()
  }, [loadUsers, loadConfig])

  const openCreateUser = () => {
    setEditingUser(null)
    userForm.resetFields()
    userForm.setFieldsValue({ role: 'user' })
    setUserModalOpen(true)
  }

  const openEditUser = (user: AdminUser) => {
    setEditingUser(user)
    userForm.setFieldsValue({
      name: user.name,
      email: user.email,
      role: user.role,
    })
    setUserModalOpen(true)
  }

  const onSubmitUser = async () => {
    const values = await userForm.validateFields()
    try {
      if (editingUser) {
        const payload: {
          name: string
          email: string
          role: UserRole
          password?: string
        } = {
          name: values.name,
          email: values.email,
          role: values.role,
        }
        if (values.password) {
          payload.password = values.password
        }
        await updateAdminUser(editingUser.id, payload)
        message.success('User updated')
      } else {
        await createAdminUser({
          name: values.name,
          email: values.email,
          password: values.password,
          role: values.role,
        })
        message.success('User created')
      }
      setUserModalOpen(false)
      void loadUsers()
    } catch {
      message.error('Save failed')
    }
  }

  const onSaveConfig = async () => {
    const values = await configForm.validateFields()
    const updates: Record<string, string> = {
      NEAR_EXPIRY_DAYS_DEFAULT: String(values.NEAR_EXPIRY_DAYS_DEFAULT),
      LOW_STOCK_ALERT_ENABLED: values.LOW_STOCK_ALERT_ENABLED ? 'TRUE' : 'FALSE',
      NEAR_EXPIRY_ALERT_ENABLED: values.NEAR_EXPIRY_ALERT_ENABLED ? 'TRUE' : 'FALSE',
    }
    try {
      setConfig(await patchSystemConfig(updates))
      message.success('System config saved')
    } catch {
      message.error('Failed to save config')
    }
  }

  const userColumns: ColumnsType<AdminUser> = useMemo(
    () => [
      { title: 'Name', dataIndex: 'name', key: 'name' },
      { title: 'Email', dataIndex: 'email', key: 'email' },
      { title: 'Role', dataIndex: 'role', key: 'role', width: 120 },
      {
        title: 'Created',
        dataIndex: 'created_at',
        key: 'created_at',
        width: 180,
        render: (v: string) => new Date(v).toLocaleString(),
      },
      {
        title: 'Actions',
        key: 'actions',
        width: 100,
        render: (_, row) => (
          <Button type="link" onClick={() => openEditUser(row)}>
            Edit
          </Button>
        ),
      },
    ],
    [],
  )

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Typography.Title level={3} style={{ margin: 0 }}>
        Administration
      </Typography.Title>

      <Tabs
        items={[
          {
            key: 'users',
            label: 'Users',
            children: (
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Button type="primary" onClick={openCreateUser}>
                  Create User
                </Button>
                <Table
                  rowKey="id"
                  loading={loadingUsers}
                  columns={userColumns}
                  dataSource={users}
                  pagination={false}
                />
              </Space>
            ),
          },
          {
            key: 'config',
            label: 'System Config',
            children: (
              <Form form={configForm} layout="vertical" style={{ maxWidth: 480 }}>
                <Form.Item
                  name="NEAR_EXPIRY_DAYS_DEFAULT"
                  label="Near-expiry threshold (days)"
                  rules={[{ required: true, message: 'Required' }]}
                >
                  <InputNumber min={1} max={365} style={{ width: '100%' }} />
                </Form.Item>
                <Form.Item
                  name="LOW_STOCK_ALERT_ENABLED"
                  label="Low stock alerts"
                  valuePropName="checked"
                >
                  <Switch />
                </Form.Item>
                <Form.Item
                  name="NEAR_EXPIRY_ALERT_ENABLED"
                  label="Near-expiry alerts"
                  valuePropName="checked"
                >
                  <Switch />
                </Form.Item>
                <Button type="primary" loading={loadingConfig} onClick={() => void onSaveConfig()}>
                  Save Config
                </Button>
                {config.length > 0 && (
                  <Typography.Paragraph type="secondary" style={{ marginTop: 16 }}>
                    Last updated: {new Date(config[0].updated_at).toLocaleString()}
                  </Typography.Paragraph>
                )}
              </Form>
            ),
          },
        ]}
      />

      <Modal
        title={editingUser ? 'Edit User' : 'Create User'}
        open={userModalOpen}
        onCancel={() => setUserModalOpen(false)}
        onOk={() => void onSubmitUser()}
        destroyOnClose
      >
        <Form form={userForm} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Required' }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="email"
            label="Email"
            rules={[{ required: true, type: 'email', message: 'Valid email required' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="password"
            label={editingUser ? 'New Password (optional)' : 'Password'}
            rules={editingUser ? [] : [{ required: true, min: 8, message: 'Min 8 characters' }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="Role" rules={[{ required: true }]}>
            <Select options={ROLES.map((role) => ({ value: role, label: role }))} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  )
}
