import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, Form, Input, Select, Space, Switch, Table, Tabs, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { downloadReportExport, fetchReport, type ReportPageResult } from '../../api/report'

const REPORT_TABS = [
  { key: 'stock-on-hand', label: 'Stock On Hand' },
  { key: 'movement-ledger', label: 'Movement Ledger' },
  { key: 'wo-summary', label: 'WO Summary' },
  { key: 'shortage-damage', label: 'Shortage / Damage' },
  { key: 'audit-trail', label: 'Audit Trail' },
  { key: 'partial-completion', label: 'Partial Completion Log' },
] as const

type ReportKey = (typeof REPORT_TABS)[number]['key']

function buildParams(reportKey: ReportKey, values: Record<string, unknown>) {
  const params: Record<string, string | number | boolean | undefined> = {}
  if (reportKey === 'stock-on-hand') {
    params.item_type = values.item_type as string | undefined
    params.lot_status = values.lot_status as string | undefined
    params.item_code = values.item_code as string | undefined
    params.near_expiry = values.near_expiry ? 'true' : undefined
  }
  if (reportKey === 'movement-ledger') {
    params.from_date = values.from_date as string | undefined
    params.to_date = values.to_date as string | undefined
    params.transaction_type = values.transaction_type as string | undefined
    params.item_code = values.item_code as string | undefined
  }
  if (reportKey === 'wo-summary') {
    params.from_date = values.from_date as string | undefined
    params.to_date = values.to_date as string | undefined
    params.status = values.status as string | undefined
    params.target_fg_code = values.target_fg_code as string | undefined
  }
  if (reportKey === 'shortage-damage') {
    params.from_date = values.from_date as string | undefined
    params.to_date = values.to_date as string | undefined
    params.reason_code = values.reason_code as string | undefined
  }
  if (reportKey === 'audit-trail') {
    params.from_date = values.from_date as string | undefined
    params.to_date = values.to_date as string | undefined
    params.wo_number = values.wo_number as string | undefined
    params.action = values.action as string | undefined
  }
  if (reportKey === 'partial-completion') {
  params.from_date = values.from_date as string | undefined
  params.to_date = values.to_date as string | undefined
    params.approval_status = values.approval_status as string | undefined
  }
  return params
}

function DateRangeFields() {
  return (
    <>
      <Form.Item name="from_date" label="From">
        <Input placeholder="YYYY-MM-DD" />
      </Form.Item>
      <Form.Item name="to_date" label="To">
        <Input placeholder="YYYY-MM-DD" />
      </Form.Item>
    </>
  )
}

function ReportFilters({ reportKey }: { reportKey: ReportKey }) {
  if (reportKey === 'stock-on-hand') {
    return (
      <>
        <Form.Item name="item_type" label="Item Type">
          <Select allowClear options={['FG', 'SFG', 'RM'].map((v) => ({ value: v, label: v }))} />
        </Form.Item>
        <Form.Item name="lot_status" label="LOT Status">
          <Select allowClear options={['Active', 'Hold', 'Quarantined'].map((v) => ({ value: v, label: v }))} />
        </Form.Item>
        <Form.Item name="item_code" label="Item Code">
          <Input />
        </Form.Item>
        <Form.Item name="near_expiry" label="Near Expiry Only" valuePropName="checked">
          <Switch />
        </Form.Item>
      </>
    )
  }
  if (reportKey === 'movement-ledger') {
    return (
      <>
        <DateRangeFields />
        <Form.Item name="transaction_type" label="Transaction Type">
          <Input placeholder="e.g. PO_RECEIPT" />
        </Form.Item>
        <Form.Item name="item_code" label="Item Code">
          <Input />
        </Form.Item>
      </>
    )
  }
  if (reportKey === 'wo-summary') {
    return (
      <>
        <DateRangeFields />
        <Form.Item name="status" label="WO Status">
          <Input />
        </Form.Item>
        <Form.Item name="target_fg_code" label="FG Code">
          <Input />
        </Form.Item>
      </>
    )
  }
  if (reportKey === 'shortage-damage') {
    return (
      <>
        <DateRangeFields />
        <Form.Item name="reason_code" label="Reason Code">
          <Input />
        </Form.Item>
      </>
    )
  }
  if (reportKey === 'audit-trail') {
    return (
      <>
        <DateRangeFields />
        <Form.Item name="wo_number" label="WO Number">
          <Input />
        </Form.Item>
        <Form.Item name="action" label="Action">
          <Input />
        </Form.Item>
      </>
    )
  }
  return (
    <>
      <DateRangeFields />
      <Form.Item name="approval_status" label="Approval Status">
        <Select
          allowClear
          options={['PENDING', 'APPROVED', 'REJECTED', 'WITHDRAWN'].map((v) => ({ value: v, label: v }))}
        />
      </Form.Item>
    </>
  )
}

export function ReportsPage() {
  const [activeKey, setActiveKey] = useState<ReportKey>('stock-on-hand')
  const [form] = Form.useForm()
  const [data, setData] = useState<ReportPageResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const values = form.getFieldsValue()
      const params = buildParams(activeKey, values)
      params.page = page
      params.page_size = pageSize
      setData(await fetchReport(activeKey, params))
    } catch {
      message.error('Failed to load report')
    } finally {
      setLoading(false)
    }
  }, [activeKey, form, page, pageSize])

  useEffect(() => {
    void load()
  }, [load])

  const columns: ColumnsType<Record<string, string>> = useMemo(() => {
    if (!data) return []
    return data.headers.map((header, index) => ({
      title: header,
      dataIndex: String(index),
      sorter: (a, b) => String(a[String(index)] ?? '').localeCompare(String(b[String(index)] ?? '')),
    }))
  }, [data])

  const tableData = useMemo(() => {
    if (!data) return []
    return data.rows.map((row, i) => {
      const record: Record<string, string> = { key: String(i) }
      row.forEach((cell, idx) => {
        record[String(idx)] = cell
      })
      return record
    })
  }, [data])

  const handleExport = async (format: 'csv' | 'xlsx') => {
    try {
      const params = buildParams(activeKey, form.getFieldsValue())
      await downloadReportExport(activeKey, format, params)
    } catch {
      message.error('Export failed')
    }
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Typography.Title level={3} style={{ margin: 0 }}>
        Reports
      </Typography.Title>

      <Tabs
        activeKey={activeKey}
        onChange={(k) => {
          setActiveKey(k as ReportKey)
          setPage(1)
        }}
        items={REPORT_TABS.map((t) => ({ key: t.key, label: t.label }))}
      />

      <Form form={form} layout="inline" onFinish={() => { setPage(1); void load() }}>
        <ReportFilters reportKey={activeKey} />
        <Form.Item>
          <Button type="primary" htmlType="submit">
            Apply Filters
          </Button>
        </Form.Item>
        <Form.Item>
          <Button onClick={() => void handleExport('csv')}>Export CSV</Button>
        </Form.Item>
        <Form.Item>
          <Button onClick={() => void handleExport('xlsx')}>Export Excel</Button>
        </Form.Item>
      </Form>

      <Table
        rowKey="key"
        loading={loading}
        columns={columns}
        dataSource={tableData}
        pagination={{
          current: page,
          pageSize,
          total: data?.total ?? 0,
          showSizeChanger: true,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        }}
        scroll={{ x: true }}
      />
    </Space>
  )
}
