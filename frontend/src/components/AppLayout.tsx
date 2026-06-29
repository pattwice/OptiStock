import {
  AppstoreOutlined,
  BellOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  InboxOutlined,
  LogoutOutlined,
  SettingOutlined,
  ShoppingOutlined,
  SwapOutlined,
} from '@ant-design/icons'
import { Layout, Menu, Button, Typography, Space, Badge } from 'antd'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { logout } from '../api/auth'
import { useAuthStore } from '../store/authStore'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: '/', icon: <AppstoreOutlined />, label: <Link to="/">Dashboard</Link> },
  {
    key: 'inventory',
    icon: <ShoppingOutlined />,
    label: 'Inventory',
    children: [
      { key: '/inventory/items', label: <Link to="/inventory/items">Items</Link> },
      { key: '/inventory/bom', label: <Link to="/inventory/bom">BOM</Link> },
      { key: '/inventory/lots', label: <Link to="/inventory/lots">LOTs</Link> },
      { key: '/inventory/receiving', icon: <InboxOutlined />, label: <Link to="/inventory/receiving">Receiving</Link> },
      { key: '/inventory/adjustments', icon: <SwapOutlined />, label: <Link to="/inventory/adjustments">Adjustments</Link> },
      { key: '/inventory/stock', icon: <DatabaseOutlined />, label: <Link to="/inventory/stock">Stock On Hand</Link> },
    ],
  },
  { key: '/workorders', icon: <FileTextOutlined />, label: 'Work Orders', disabled: true },
  { key: '/reports', icon: <FileTextOutlined />, label: 'Reports', disabled: true },
  { key: '/admin', icon: <SettingOutlined />, label: 'Admin', disabled: true },
]

function selectedKeys(pathname: string): string[] {
  if (pathname === '/') return ['/']
  if (pathname.startsWith('/inventory')) return [pathname]
  return [pathname]
}

function openKeys(pathname: string): string[] {
  return pathname.startsWith('/inventory') ? ['inventory'] : []
}

export function AppLayout() {
  const location = useLocation()
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const clearSession = useAuthStore((state) => state.clearSession)

  const handleLogout = async () => {
    try {
      await logout()
    } finally {
      clearSession()
      navigate('/login', { replace: true })
    }
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={64}>
        <div style={{ color: '#fff', fontWeight: 700, padding: '16px 20px' }}>OptiStock</div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys(location.pathname)}
          defaultOpenKeys={openKeys(location.pathname)}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 24px',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Typography.Text strong>Production & Inventory</Typography.Text>
          <Space>
            <Badge dot>
              <Button type="text" icon={<BellOutlined />} aria-label="Notifications" />
            </Badge>
            <Typography.Text>{user?.name}</Typography.Text>
            <Button icon={<LogoutOutlined />} onClick={handleLogout}>
              Logout
            </Button>
          </Space>
        </Header>
        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
