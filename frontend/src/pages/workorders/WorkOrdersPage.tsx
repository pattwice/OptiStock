import { useCallback, useEffect, useState } from 'react'
import { Button, Select, Space, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { Link, useNavigate } from 'react-router-dom'
import { listWorkOrders } from '../../api/workorder'
import { WO_STATUS_COLORS, WO_STATUSES } from '../../constants/workorder'
import type { WOStatus, WorkOrder } from '../../types/workorder'

export function WorkOrdersPage() {
  const [rows, setRows] = useState<WorkOrder[]>([])
  const [loading, setLoading] = useState(false)
  const [statusFilter, setStatusFilter] = useState<WOStatus | undefined>()
  const navigate = useNavigate()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(await listWorkOrders(statusFilter))
    } catch {
      message.error('Failed to load work orders')
    } finally {
      setLoading(false)
    }
  }, [statusFilter])

  useEffect(() => {
    void load()
  }, [load])

  const columns: ColumnsType<WorkOrder> = [
    {
      title: 'WO Number',
      dataIndex: 'wo_number',
      render: (v: string) => <Link to={`/workorders/${encodeURIComponent(v)}`}>{v}</Link>,
    },
    { title: 'FG Code', dataIndex: 'target_fg_code' },
    { title: 'Target Qty', dataIndex: 'target_qty' },
    {
      title: 'Completion %',
      dataIndex: 'completion_pct',
      render: (v: string | null) => (v ? `${v}%` : '—'),
    },
    {
      title: 'Status',
      dataIndex: 'wo_status',
      render: (s: WOStatus) => <Tag color={WO_STATUS_COLORS[s]}>{s}</Tag>,
    },
    { title: 'Created', dataIndex: 'created_at', render: (v: string) => new Date(v).toLocaleString() },
  ]

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          Work Orders
        </Typography.Title>
        <Button type="primary" onClick={() => navigate('/workorders/new')}>
          Create WO
        </Button>
      </Space>
      <Select
        allowClear
        placeholder="Filter by status"
        style={{ width: 220 }}
        options={WO_STATUSES.map((s) => ({ value: s, label: s }))}
        value={statusFilter}
        onChange={(v) => setStatusFilter(v)}
      />
      <Table rowKey="wo_number" loading={loading} columns={columns} dataSource={rows} />
    </Space>
  )
}
