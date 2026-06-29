import { useCallback, useEffect, useState } from 'react'
import { Button, Input, Select, Space, Table, Tag, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { listStockOnHand } from '../../api/inventory'
import { LOT_STATUS_COLORS, LOT_STATUSES } from '../../constants/inventory'
import type { LotStatus, StockOnHandRow } from '../../types/inventory'
import { downloadCsv } from '../../utils/csv'

export function StockOnHandPage() {
  const [rows, setRows] = useState<StockOnHandRow[]>([])
  const [loading, setLoading] = useState(false)
  const [itemCode, setItemCode] = useState('')
  const [status, setStatus] = useState<LotStatus | undefined>()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(
        await listStockOnHand({
          item_code: itemCode || undefined,
          status,
        }),
      )
    } catch {
      message.error('Failed to load stock')
    } finally {
      setLoading(false)
    }
  }, [itemCode, status])

  useEffect(() => {
    void load()
  }, [load])

  const exportCsv = () => {
    downloadCsv(
      'stock-on-hand.csv',
      [
        'Item',
        'Supplier Lot',
        'Status',
        'Physical',
        'Reserved',
        'Available',
        'MFG',
        'EXP',
      ],
      rows.map((r) => [
        r.item_code,
        r.supplier_lot_number,
        r.status,
        r.physical_qty,
        r.reserved_qty,
        r.available_qty,
        r.mfg_date,
        r.exp_date ?? '',
      ]),
    )
  }

  const columns: ColumnsType<StockOnHandRow> = [
    { title: 'Item', dataIndex: 'item_code', key: 'item_code', width: 120 },
    { title: 'Supplier Lot', dataIndex: 'supplier_lot_number', key: 'supplier_lot_number' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (s: LotStatus) => <Tag color={LOT_STATUS_COLORS[s]}>{s}</Tag>,
    },
    { title: 'Physical', dataIndex: 'physical_qty', key: 'physical_qty', width: 100 },
    { title: 'Reserved', dataIndex: 'reserved_qty', key: 'reserved_qty', width: 100 },
    { title: 'Available', dataIndex: 'available_qty', key: 'available_qty', width: 100 },
    { title: 'MFG', dataIndex: 'mfg_date', key: 'mfg_date', width: 110 },
    { title: 'EXP', dataIndex: 'exp_date', key: 'exp_date', width: 110 },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          Stock On Hand
        </Typography.Title>
        <Space>
          <Input
            allowClear
            placeholder="Item code"
            style={{ width: 160 }}
            value={itemCode}
            onChange={(e) => setItemCode(e.target.value)}
            onPressEnter={() => void load()}
          />
          <Select
            allowClear
            placeholder="Status"
            style={{ width: 140 }}
            options={LOT_STATUSES.map((s) => ({ value: s, label: s }))}
            value={status}
            onChange={setStatus}
          />
          <Button onClick={() => void load()}>Refresh</Button>
          <Button onClick={exportCsv} disabled={rows.length === 0}>
            Export CSV
          </Button>
        </Space>
      </Space>
      <Table rowKey="lot_internal_id" loading={loading} columns={columns} dataSource={rows} />
    </div>
  )
}
