import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Select, Typography, message } from 'antd'
import { listItems, poReceipt } from '../../api/inventory'
import type { Item } from '../../types/inventory'

export function ReceivingPage() {
  const [items, setItems] = useState<Item[]>([])
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    listItems('RM')
      .then(setItems)
      .catch(() => message.error('Failed to load items'))
  }, [])

  const onSubmit = async () => {
    const values = await form.validateFields()
    setLoading(true)
    try {
      const result = await poReceipt({
        item_code: values.item_code,
        supplier_lot_number: values.supplier_lot_number,
        supplier_name: values.supplier_name || undefined,
        mfg_date: values.mfg_date,
        exp_date: values.exp_date || null,
        qty_received: values.qty_received,
        doc_ref: values.doc_ref,
        damage_qty: values.damage_qty || undefined,
        damage_reason_code: values.damage_qty ? 'Supplier_Damage' : undefined,
      })
      message.success(`Received. LOT ${result.lot_internal_id}`)
      form.resetFields()
    } catch (e: unknown) {
      const msg =
        e && typeof e === 'object' && 'response' in e
          ? (e as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error
              ?.message
          : undefined
      message.error(msg ?? 'PO receipt failed')
    } finally {
      setLoading(false)
    }
  }

  const itemOptions = items.map((i) => ({
    value: i.item_code,
    label: `${i.item_code} — ${i.description}`,
  }))

  return (
    <div>
      <Typography.Title level={3}>PO Receipt</Typography.Title>
      <Card style={{ maxWidth: 560 }}>
        <Form form={form} layout="vertical" onFinish={() => void onSubmit()}>
          <Form.Item name="item_code" label="Item" rules={[{ required: true }]}>
            <Select showSearch options={itemOptions} optionFilterProp="label" placeholder="RM item" />
          </Form.Item>
          <Form.Item name="doc_ref" label="PO / Doc Ref" rules={[{ required: true }]}>
            <Input placeholder="PO-12345" />
          </Form.Item>
          <Form.Item name="supplier_lot_number" label="Supplier Lot #" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="supplier_name" label="Supplier Name">
            <Input />
          </Form.Item>
          <Form.Item name="mfg_date" label="MFG Date" rules={[{ required: true }]}>
            <Input type="date" />
          </Form.Item>
          <Form.Item name="exp_date" label="EXP Date">
            <Input type="date" />
          </Form.Item>
          <Form.Item name="qty_received" label="Qty Received" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="damage_qty" label="Supplier Damage Qty (optional)">
            <Input placeholder="Creates ADJ_OUT with Supplier_Damage" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>
            Post Receipt
          </Button>
        </Form>
      </Card>
    </div>
  )
}
