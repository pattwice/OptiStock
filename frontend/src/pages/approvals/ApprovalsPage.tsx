import { useCallback, useEffect, useState } from 'react'
import {
  Button,
  Form,
  Input,
  Modal,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { Link } from 'react-router-dom'
import { approveRequest, listPendingApprovals, rejectRequest } from '../../api/approval'
import { APPROVAL_STATUS_COLORS } from '../../constants/approval'
import type { ApprovalRequest } from '../../types/approval'

export function ApprovalsPage() {
  const [rows, setRows] = useState<ApprovalRequest[]>([])
  const [loading, setLoading] = useState(false)
  const [rejectForm] = Form.useForm<{ resolution_notes: string }>()
  const [rejectTarget, setRejectTarget] = useState<ApprovalRequest | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(await listPendingApprovals())
    } catch {
      message.error('Failed to load pending approvals')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const handleApprove = (row: ApprovalRequest) => {
    Modal.confirm({
      title: `Approve partial close for ${row.wo_number}?`,
      content: `Completion at request: ${row.completion_pct_at_request}%`,
      onOk: async () => {
        try {
          await approveRequest(row.approval_id)
          message.success('Approval granted — work order partially completed')
          void load()
        } catch {
          message.error('Failed to approve request')
        }
      },
    })
  }

  const handleRejectSubmit = async () => {
    if (!rejectTarget) return
    const values = await rejectForm.validateFields()
    try {
      await rejectRequest(rejectTarget.approval_id, values.resolution_notes)
      message.success('Request rejected — work order returned to production')
      setRejectTarget(null)
      rejectForm.resetFields()
      void load()
    } catch {
      message.error('Failed to reject request')
    }
  }

  const columns: ColumnsType<ApprovalRequest> = [
    {
      title: 'WO',
      dataIndex: 'wo_number',
      render: (v: string) => <Link to={`/workorders/${encodeURIComponent(v)}`}>{v}</Link>,
    },
    { title: 'FG Code', dataIndex: 'target_fg_code' },
    { title: 'Target', dataIndex: 'target_qty' },
    { title: 'Produced', dataIndex: 'actual_produced_qty', render: (v) => v ?? '—' },
    {
      title: 'Completion %',
      dataIndex: 'completion_pct_at_request',
      render: (v: string) => `${v}%`,
    },
    { title: 'Requester', dataIndex: 'requested_by_name' },
    {
      title: 'Requested',
      dataIndex: 'requested_at',
      render: (v: string) => new Date(v).toLocaleString(),
    },
    {
      title: 'Status',
      dataIndex: 'approval_status',
      render: (v: ApprovalRequest['approval_status']) => (
        <Tag color={APPROVAL_STATUS_COLORS[v]}>{v}</Tag>
      ),
    },
    {
      title: 'Actions',
      render: (_, row) => (
        <Space>
          <Button type="primary" size="small" onClick={() => handleApprove(row)}>
            Approve
          </Button>
          <Button danger size="small" onClick={() => setRejectTarget(row)}>
            Reject
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Typography.Title level={3} style={{ margin: 0 }}>
        Pending Approvals
      </Typography.Title>
      <Table rowKey="approval_id" loading={loading} columns={columns} dataSource={rows} />

      <Modal
        title={rejectTarget ? `Reject ${rejectTarget.wo_number}` : 'Reject request'}
        open={rejectTarget != null}
        onCancel={() => {
          setRejectTarget(null)
          rejectForm.resetFields()
        }}
        onOk={() => void handleRejectSubmit()}
        okText="Reject"
        okButtonProps={{ danger: true }}
      >
        <Form form={rejectForm} layout="vertical">
          <Form.Item
            name="resolution_notes"
            label="Reason (required)"
            rules={[{ required: true, message: 'Rejection reason is required' }]}
          >
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  )
}
