import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import {
  Button,
  Card,
  Descriptions,
  Form,
  InputNumber,
  Modal,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate, useParams } from 'react-router-dom'
import {
  cancelWorkOrder,
  completeWorkOrder,
  getAllocationProposal,
  getWorkOrder,
  getWorkOrderAudit,
  reopenWorkOrder,
  reserveWorkOrder,
  startWorkOrder,
  updateWorkOrderActuals,
} from '../../api/workorder'
import { WO_STATUS_COLORS } from '../../constants/workorder'
import type {
  AuditEntry,
  WOAllocation,
  WORequirement,
  WOStatus,
  WorkOrderDetail,
} from '../../types/workorder'

export function WorkOrderDetailPage() {
  const { woNumber = '' } = useParams()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<WorkOrderDetail | null>(null)
  const [audit, setAudit] = useState<AuditEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [actualsForm] = Form.useForm()

  const load = useCallback(async () => {
    if (!woNumber) return
    setLoading(true)
    try {
      const [wo, log] = await Promise.all([
        getWorkOrder(woNumber),
        getWorkOrderAudit(woNumber),
      ])
      setDetail(wo)
      setAudit(log)
      actualsForm.setFieldsValue({
        actual_produced_qty: wo.work_order.actual_produced_qty ?? wo.work_order.target_qty,
        lines: wo.allocations.map((a) => ({
          allocation_id: a.allocation_id,
          actual_used_qty: a.actual_used_qty ?? a.reserved_qty,
          damage_qty: a.damage_qty ?? '0',
        })),
      })
    } catch {
      message.error('Failed to load work order')
    } finally {
      setLoading(false)
    }
  }, [actualsForm, woNumber])

  useEffect(() => {
    void load()
  }, [load])

  const wo = detail?.work_order
  const status = wo?.wo_status

  const reqColumns: ColumnsType<WORequirement> = [
    { title: 'Item', dataIndex: 'required_item_code' },
    { title: 'Needed', dataIndex: 'total_needed_qty' },
    { title: 'Available', dataIndex: 'available_qty' },
    {
      title: 'Shortage',
      dataIndex: 'shortage',
      render: (v: boolean) => (v ? <Tag color="red">SHORT</Tag> : <Tag color="green">OK</Tag>),
    },
  ]

  const allocColumns: ColumnsType<WOAllocation> = [
    { title: 'Item', dataIndex: 'item_code' },
    { title: 'LOT', dataIndex: 'supplier_lot_number' },
    { title: 'Reserved', dataIndex: 'reserved_qty' },
    { title: 'EXP', dataIndex: 'exp_date', render: (v: string | null) => v ?? '—' },
    { title: 'Used', dataIndex: 'actual_used_qty', render: (v) => v ?? '—' },
    { title: 'Damage', dataIndex: 'damage_qty', render: (v) => v ?? '—' },
  ]

  const auditColumns: ColumnsType<AuditEntry> = [
    { title: 'Time', dataIndex: 'timestamp', render: (v: string) => new Date(v).toLocaleString() },
    { title: 'User', dataIndex: 'user_name' },
    { title: 'Action', dataIndex: 'action' },
    { title: 'Old', dataIndex: 'old_value', render: (v) => v ?? '—' },
    { title: 'New', dataIndex: 'new_value', render: (v) => v ?? '—' },
  ]

  const handleReserve = async () => {
    try {
      await getAllocationProposal(woNumber)
      Modal.confirm({
        title: 'Reserve materials?',
        content: 'System will auto-allocate LOTs using FEFO+FIFO.',
        onOk: async () => {
          await reserveWorkOrder(woNumber, {})
          message.success('Materials reserved')
          void load()
        },
      })
    } catch {
      message.error('Cannot reserve — check stock shortages')
    }
  }

  const handleStart = async () => {
    try {
      await startWorkOrder(woNumber)
      message.success('Production started')
      void load()
    } catch {
      message.error('Failed to start production')
    }
  }

  const handleSaveActuals = async () => {
    const values = await actualsForm.validateFields()
    try {
      await updateWorkOrderActuals(woNumber, {
        actual_produced_qty: String(values.actual_produced_qty),
        lines: values.lines.map((l: { allocation_id: string; actual_used_qty: number; damage_qty?: number }) => ({
          allocation_id: l.allocation_id,
          actual_used_qty: String(l.actual_used_qty),
          damage_qty: l.damage_qty != null ? String(l.damage_qty) : undefined,
        })),
      })
      message.success('Actuals saved')
      void load()
    } catch {
      message.error('Failed to save actuals')
    }
  }

  const handleComplete = async () => {
    const values = await actualsForm.validateFields()
    Modal.confirm({
      title: 'Complete work order?',
      content: 'Ledger writes are permanent. Full completion only (partial close is Phase 3).',
      onOk: async () => {
        try {
          await completeWorkOrder(woNumber, {
            actual_produced_qty: String(values.actual_produced_qty),
            lines: values.lines.map((l: { allocation_id: string; actual_used_qty: number; damage_qty?: number }) => ({
              allocation_id: l.allocation_id,
              actual_used_qty: String(l.actual_used_qty),
              damage_qty: l.damage_qty != null ? String(l.damage_qty) : undefined,
            })),
          })
          message.success('Work order completed')
          void load()
        } catch {
          message.error('Completion failed — check actuals and stock')
        }
      },
    })
  }

  const handleCancel = async () => {
    try {
      await cancelWorkOrder(woNumber)
      message.success('Work order cancelled')
      void load()
    } catch {
      message.error('Failed to cancel')
    }
  }

  const handleReopen = async () => {
    try {
      await reopenWorkOrder(woNumber)
      message.success('Work order reopened')
      void load()
    } catch {
      message.error('Failed to reopen')
    }
  }

  const actionButtons = useMemo(() => {
    if (!status) return null
    const buttons: ReactNode[] = []
    if (status === 'DRAFT') {
      buttons.push(
        <Button key="reserve" type="primary" onClick={() => void handleReserve()}>
          Reserve
        </Button>,
        <Button key="cancel" danger onClick={() => void handleCancel()}>
          Cancel
        </Button>,
      )
    }
    if (status === 'RESERVED') {
      buttons.push(
        <Button key="start" type="primary" onClick={() => void handleStart()}>
          Start Production
        </Button>,
        <Button key="cancel" danger onClick={() => void handleCancel()}>
          Cancel
        </Button>,
      )
    }
    if (status === 'IN_PRODUCTION') {
      buttons.push(
        <Button key="save" onClick={() => void handleSaveActuals()}>
          Save Actuals
        </Button>,
        <Button key="complete" type="primary" onClick={() => void handleComplete()}>
          Complete
        </Button>,
        <Button key="cancel" danger onClick={() => void handleCancel()}>
          Cancel
        </Button>,
      )
    }
    if (status === 'CANCELLED') {
      buttons.push(
        <Button key="reopen" type="primary" onClick={() => void handleReopen()}>
          Reopen
        </Button>,
      )
    }
    return buttons
  }, [status])

  if (!wo) {
    return <Typography.Text>Loading…</Typography.Text>
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          {wo.wo_number}
        </Typography.Title>
        <Button onClick={() => navigate('/workorders')}>Back to list</Button>
      </Space>

      <Card loading={loading}>
        <Descriptions column={2}>
          <Descriptions.Item label="FG Code">{wo.target_fg_code}</Descriptions.Item>
          <Descriptions.Item label="Target Qty">{wo.target_qty}</Descriptions.Item>
          <Descriptions.Item label="Status">
            <Tag color={WO_STATUS_COLORS[wo.wo_status as WOStatus]}>{wo.wo_status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Completion %">{wo.completion_pct ? `${wo.completion_pct}%` : '—'}</Descriptions.Item>
        </Descriptions>
        <Space style={{ marginTop: 16 }}>{actionButtons}</Space>
      </Card>

      <Tabs
        items={[
          {
            key: 'requirements',
            label: 'Material Requirements',
            children: (
              <Table
                rowKey="req_id"
                columns={reqColumns}
                dataSource={detail?.requirements ?? []}
                pagination={false}
              />
            ),
          },
          {
            key: 'allocations',
            label: 'LOT Allocations',
            children: (
              <Table
                rowKey="allocation_id"
                columns={allocColumns}
                dataSource={detail?.allocations ?? []}
                pagination={false}
              />
            ),
          },
          {
            key: 'actuals',
            label: 'Actuals',
            disabled: status !== 'IN_PRODUCTION',
            children: (
              <Form form={actualsForm} layout="vertical">
                <Form.Item name="actual_produced_qty" label="Actual Produced Qty" rules={[{ required: true }]}>
                  <InputNumber min={0} style={{ width: 200 }} />
                </Form.Item>
                <Form.List name="lines">
                  {(fields) => (
                    <Table
                      rowKey="key"
                      pagination={false}
                      dataSource={fields.map((f, i) => ({
                        key: f.key,
                        index: i,
                        allocation: detail?.allocations[i],
                      }))}
                      columns={[
                        {
                          title: 'Item',
                          render: (_, row) => row.allocation?.item_code,
                        },
                        {
                          title: 'Reserved',
                          render: (_, row) => row.allocation?.reserved_qty,
                        },
                        {
                          title: 'Actual Used',
                          render: (_, row) => (
                            <>
                              <Form.Item name={[row.index, 'allocation_id']} hidden>
                                <input type="hidden" />
                              </Form.Item>
                              <Form.Item name={[row.index, 'actual_used_qty']} noStyle rules={[{ required: true }]}>
                                <InputNumber min={0} style={{ width: 120 }} />
                              </Form.Item>
                            </>
                          ),
                        },
                        {
                          title: 'Damage',
                          render: (_, row) => (
                            <Form.Item name={[row.index, 'damage_qty']} noStyle>
                              <InputNumber min={0} style={{ width: 120 }} />
                            </Form.Item>
                          ),
                        },
                      ]}
                    />
                  )}
                </Form.List>
              </Form>
            ),
          },
          {
            key: 'audit',
            label: 'Audit Log',
            children: <Table rowKey="log_id" columns={auditColumns} dataSource={audit} />,
          },
        ]}
      />
    </Space>
  )
}
