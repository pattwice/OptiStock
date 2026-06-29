import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, Radio, Select, Typography, message } from 'antd'
import { adjustment, listLots } from '../../api/inventory'
import { REASON_CODES } from '../../constants/inventory'
import type { Lot } from '../../types/inventory'

export function AdjustmentsPage() {
  const [lots, setLots] = useState<Lot[]>([])
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    listLots()
      .then(setLots)
      .catch(() => message.error('Failed to load lots'))
  }, [])

  const onSubmit = async () => {
    const values = await form.validateFields()
    setLoading(true)
    try {
      await adjustment({
        lot_internal_id: values.lot_internal_id,
        type: values.type,
        qty: values.qty,
        doc_ref: values.doc_ref,
        reason_code: values.reason_code,
      })
      message.success('Adjustment posted')
      form.resetFields()
    } catch (e: unknown) {
      const code =
        e && typeof e === 'object' && 'response' in e
          ? (e as { response?: { data?: { error?: { code?: string; message?: string } } } }).response
              ?.data?.error
          : undefined
      message.error(code?.message ?? code?.code ?? 'Adjustment failed')
    } finally {
      setLoading(false)
    }
  }

  const lotOptions = lots.map((l) => ({
    value: l.lot_internal_id,
    label: `${l.item_code} / ${l.supplier_lot_number} (${l.status})`,
  }))

  return (
    <div>
      <Typography.Title level={3}>Stock Adjustments</Typography.Title>
      <Card style={{ maxWidth: 560 }}>
        <Form form={form} layout="vertical" initialValues={{ type: 'ADJ_IN' }} onFinish={() => void onSubmit()}>
          <Form.Item name="lot_internal_id" label="LOT" rules={[{ required: true }]}>
            <Select showSearch options={lotOptions} optionFilterProp="label" />
          </Form.Item>
          <Form.Item name="type" label="Type" rules={[{ required: true }]}>
            <Radio.Group>
              <Radio value="ADJ_IN">ADJ IN</Radio>
              <Radio value="ADJ_OUT">ADJ OUT</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item name="qty" label="Quantity (positive)" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="doc_ref" label="Doc Ref" rules={[{ required: true }]}>
            <Input placeholder="CC-2026-001" />
          </Form.Item>
          <Form.Item name="reason_code" label="Reason Code" rules={[{ required: true }]}>
            <Select options={REASON_CODES.map((r) => ({ value: r, label: r }))} />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} block>
            Post Adjustment
          </Button>
        </Form>
      </Card>
    </div>
  )
}
