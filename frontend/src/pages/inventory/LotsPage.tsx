import { useCallback, useEffect, useState } from 'react'
import { Select, Space, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { listLots, updateLotStatus } from '../../api/inventory'
import { LOT_STATUS_COLORS, LOT_STATUSES } from '../../constants/inventory'
import { useAuthStore } from '../../store/authStore'
import type { Lot, LotStatus } from '../../types/inventory'

const SUPERVISOR_ROLES = new Set(['supervisor', 'admin'])

export function LotsPage() {
  const user = useAuthStore((s) => s.user)
  const canChangeStatus = user ? SUPERVISOR_ROLES.has(user.role) : false
  const [lots, setLots] = useState<Lot[]>([])
  const [loading, setLoading] = useState(false)
  const [statusFilter, setStatusFilter] = useState<LotStatus | undefined>()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setLots(await listLots(statusFilter ? { status: statusFilter } : undefined))
    } catch {
      message.error('Failed to load lots')
    } finally {
      setLoading(false)
    }
  }, [statusFilter])

  useEffect(() => {
    void load()
  }, [load])

  const changeStatus = async (lotId: string, status: LotStatus) => {
    try {
      await updateLotStatus(lotId, status)
      message.success(`Status → ${status}`)
      void load()
    } catch {
      message.error('Status change failed')
    }
  }

  const columns: ColumnsType<Lot> = [
    { title: 'Item', dataIndex: 'item_code', key: 'item_code', width: 120 },
    { title: 'Supplier Lot', dataIndex: 'supplier_lot_number', key: 'supplier_lot_number' },
    { title: 'Supplier', dataIndex: 'supplier_name', key: 'supplier_name' },
    { title: 'MFG', dataIndex: 'mfg_date', key: 'mfg_date', width: 110 },
    { title: 'EXP', dataIndex: 'exp_date', key: 'exp_date', width: 110 },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (s: LotStatus) => <Tag color={LOT_STATUS_COLORS[s]}>{s}</Tag>,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (_, row) =>
        canChangeStatus ? (
          <Select
            size="small"
            style={{ width: 140 }}
            value={row.status}
            options={LOT_STATUSES.map((s) => ({ value: s, label: s }))}
            onChange={(v) => void changeStatus(row.lot_internal_id, v)}
          />
        ) : (
          <Typography.Text type="secondary">Supervisor only</Typography.Text>
        ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          LOT Master
        </Typography.Title>
        <Select
          allowClear
          placeholder="Filter status"
          style={{ width: 160 }}
          options={LOT_STATUSES.map((s) => ({ value: s, label: s }))}
          value={statusFilter}
          onChange={setStatusFilter}
        />
      </Space>
      <Table rowKey="lot_internal_id" loading={loading} columns={columns} dataSource={lots} />
    </div>
  )
}
