import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Input, InputNumber, Select, Space, Typography, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { createWorkOrder } from '../../api/workorder'
import { listItems } from '../../api/inventory'
import type { Item } from '../../types/inventory'

export function WorkOrderCreatePage() {
  const [form] = Form.useForm()
  const [fgItems, setFgItems] = useState<Item[]>([])
  const [submitting, setSubmitting] = useState(false)
  const navigate = useNavigate()

  const loadItems = useCallback(async () => {
    try {
      const [fg, sfg] = await Promise.all([listItems('FG'), listItems('SFG')])
      setFgItems([...fg, ...sfg])
    } catch {
      message.error('Failed to load FG/SFG items')
    }
  }, [])

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  const onSubmit = async () => {
    const values = await form.validateFields()
    setSubmitting(true)
    try {
      const wo = await createWorkOrder({
        wo_number: values.wo_number,
        target_fg_code: values.target_fg_code,
        target_qty: String(values.target_qty),
      })
      message.success('Work order created')
      navigate(`/workorders/${encodeURIComponent(wo.work_order.wo_number)}`)
    } catch {
      message.error('Failed to create work order — check BOM is active for the FG')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%', maxWidth: 480 }}>
      <Typography.Title level={3} style={{ margin: 0 }}>
        Create Work Order
      </Typography.Title>
      <Form form={form} layout="vertical" onFinish={() => void onSubmit()}>
        <Form.Item name="wo_number" label="WO Number" rules={[{ required: true }]}>
          <Input placeholder="WO-2026-001" />
        </Form.Item>
        <Form.Item name="target_fg_code" label="Target FG / SFG" rules={[{ required: true }]}>
          <Select
            showSearch
            optionFilterProp="label"
            options={fgItems.map((i) => ({
              value: i.item_code,
              label: `${i.item_code} — ${i.description}`,
            }))}
          />
        </Form.Item>
        <Form.Item name="target_qty" label="Target Qty" rules={[{ required: true }]}>
          <InputNumber min={0.000001} style={{ width: '100%' }} />
        </Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={submitting}>
            Create
          </Button>
          <Button onClick={() => navigate('/workorders')}>Cancel</Button>
        </Space>
      </Form>
    </Space>
  )
}
