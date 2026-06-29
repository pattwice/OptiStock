import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Input, InputNumber, Modal, Select, Space, Table, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { createItem, listItems, updateItem } from '../../api/inventory'
import { ITEM_TYPES } from '../../constants/inventory'
import type { Item, ItemType } from '../../types/inventory'

export function ItemsPage() {
  const [items, setItems] = useState<Item[]>([])
  const [loading, setLoading] = useState(false)
  const [typeFilter, setTypeFilter] = useState<ItemType | undefined>()
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Item | null>(null)
  const [form] = Form.useForm()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setItems(await listItems(typeFilter))
    } catch {
      message.error('Failed to load items')
    } finally {
      setLoading(false)
    }
  }, [typeFilter])

  useEffect(() => {
    void load()
  }, [load])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ item_type: 'RM', min_stock_level: '0' })
    setModalOpen(true)
  }

  const openEdit = (item: Item) => {
    setEditing(item)
    form.setFieldsValue(item)
    setModalOpen(true)
  }

  const onSubmit = async () => {
    const values = await form.validateFields()
    try {
      if (editing) {
        await updateItem(editing.item_code, {
          description: values.description,
          unit: values.unit,
          item_type: values.item_type,
          min_stock_level: String(values.min_stock_level),
          expiry_threshold_days: values.expiry_threshold_days ?? null,
          shelf_life_days: values.shelf_life_days ?? null,
        })
        message.success('Item updated')
      } else {
        await createItem({
          ...values,
          min_stock_level: String(values.min_stock_level),
        })
        message.success('Item created')
      }
      setModalOpen(false)
      void load()
    } catch {
      message.error('Save failed')
    }
  }

  const columns: ColumnsType<Item> = [
    { title: 'Item Code', dataIndex: 'item_code', key: 'item_code' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    { title: 'Type', dataIndex: 'item_type', key: 'item_type', width: 80 },
    { title: 'Unit', dataIndex: 'unit', key: 'unit', width: 80 },
    { title: 'Min Stock', dataIndex: 'min_stock_level', key: 'min_stock_level', width: 100 },
    {
      title: 'Actions',
      key: 'actions',
      width: 100,
      render: (_, row) => (
        <Button type="link" onClick={() => openEdit(row)}>
          Edit
        </Button>
      ),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          Item Master
        </Typography.Title>
        <Space>
          <Select
            allowClear
            placeholder="Filter type"
            style={{ width: 140 }}
            options={ITEM_TYPES.map((t) => ({ value: t, label: t }))}
            value={typeFilter}
            onChange={setTypeFilter}
          />
          <Button type="primary" onClick={openCreate}>
            New Item
          </Button>
        </Space>
      </Space>

      <Table rowKey="item_code" loading={loading} columns={columns} dataSource={items} />

      <Modal
        title={editing ? 'Edit Item' : 'New Item'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => void onSubmit()}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item name="item_code" label="Item Code" rules={[{ required: true }]} hidden={!!editing}>
            <Input disabled={!!editing} />
          </Form.Item>
          <Form.Item name="description" label="Description" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="item_type" label="Type" rules={[{ required: true }]}>
            <Select options={ITEM_TYPES.map((t) => ({ value: t, label: t }))} />
          </Form.Item>
          <Form.Item name="unit" label="Unit" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="min_stock_level" label="Min Stock Level" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="expiry_threshold_days" label="Expiry Threshold (days)">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item name="shelf_life_days" label="Shelf Life (days)">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
