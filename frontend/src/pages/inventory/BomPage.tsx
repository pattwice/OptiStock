import { useEffect, useState } from 'react'
import {
  Button,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { activateBomVersion, createBomRow, listBom, listItems } from '../../api/inventory'
import type { BomRow, Item } from '../../types/inventory'

export function BomPage() {
  const [parentCode, setParentCode] = useState<string>()
  const [parentItems, setParentItems] = useState<Item[]>([])
  const [componentItems, setComponentItems] = useState<Item[]>([])
  const [rows, setRows] = useState<BomRow[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    listItems()
      .then((data) => {
        setParentItems(data.filter((i) => i.item_type === 'FG' || i.item_type === 'SFG'))
        setComponentItems(data.filter((i) => i.item_type === 'RM' || i.item_type === 'SFG'))
      })
      .catch(() => message.error('Failed to load items'))
  }, [])

  useEffect(() => {
    if (!parentCode) {
      setRows([])
      return
    }
    setLoading(true)
    listBom(parentCode)
      .then(setRows)
      .catch(() => message.error('Failed to load BOM'))
      .finally(() => setLoading(false))
  }, [parentCode])

  const onAdd = async () => {
    const values = await form.validateFields()
    try {
      await createBomRow({
        parent_item_code: parentCode!,
        component_item_code: values.component_item_code,
        qty_per_set: values.qty_per_set,
        bom_version: values.bom_version,
        is_active: false,
      })
      message.success('BOM row added')
      setModalOpen(false)
      form.resetFields()
      const data = await listBom(parentCode!)
      setRows(data)
    } catch {
      message.error('Failed to add BOM row')
    }
  }

  const onActivate = async (version: string) => {
    try {
      await activateBomVersion(parentCode!, version)
      message.success(`Activated version ${version}`)
      setRows(await listBom(parentCode!))
    } catch {
      message.error('Activate failed')
    }
  }

  const columns: ColumnsType<BomRow> = [
    { title: 'Version', dataIndex: 'bom_version', key: 'bom_version' },
    { title: 'Component', dataIndex: 'component_item_code', key: 'component_item_code' },
    { title: 'Qty / Set', dataIndex: 'qty_per_set', key: 'qty_per_set', width: 100 },
    {
      title: 'Active',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 90,
      render: (active: boolean) => (active ? <Tag color="green">Yes</Tag> : <Tag>No</Tag>),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, row) =>
        !row.is_active ? (
          <Button type="link" onClick={() => void onActivate(row.bom_version)}>
            Activate version
          </Button>
        ) : null,
    },
  ]

  const fgOptions = parentItems.map((i) => ({
    value: i.item_code,
    label: `${i.item_code} — ${i.description}`,
  }))
  const componentOptions = componentItems.map((i) => ({
    value: i.item_code,
    label: `${i.item_code} — ${i.description}`,
  }))

  return (
    <div>
      <Space style={{ marginBottom: 16, width: '100%', justifyContent: 'space-between' }}>
        <Typography.Title level={3} style={{ margin: 0 }}>
          BOM Ledger
        </Typography.Title>
        <Space>
          <Select
            showSearch
            placeholder="Parent item (FG/SFG)"
            style={{ width: 320 }}
            options={fgOptions}
            value={parentCode}
            onChange={setParentCode}
            optionFilterProp="label"
          />
          <Button type="primary" disabled={!parentCode} onClick={() => setModalOpen(true)}>
            Add row
          </Button>
        </Space>
      </Space>

      <Table rowKey="bom_id" loading={loading} columns={columns} dataSource={rows} pagination={false} />

      <Modal title="Add BOM row" open={modalOpen} onCancel={() => setModalOpen(false)} onOk={() => void onAdd()}>
        <Form form={form} layout="vertical">
          <Form.Item name="bom_version" label="BOM Version" rules={[{ required: true }]}>
            <Input placeholder="e.g. v1" />
          </Form.Item>
          <Form.Item name="component_item_code" label="Component" rules={[{ required: true }]}>
            <Select showSearch options={componentOptions} optionFilterProp="label" />
          </Form.Item>
          <Form.Item name="qty_per_set" label="Qty per set" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
