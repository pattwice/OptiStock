export type ApprovalStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'WITHDRAWN'

export type ApprovalType = 'PARTIAL_COMPLETION'

export interface ApprovalRequest {
  approval_id: string
  wo_number: string
  approval_type: ApprovalType
  requested_by: string
  requested_by_name?: string
  requested_at: string
  completion_pct_at_request: string
  approval_status: ApprovalStatus
  resolved_by?: string | null
  resolved_by_name?: string | null
  resolved_at?: string | null
  resolution_notes?: string | null
  target_fg_code?: string
  target_qty?: string
  actual_produced_qty?: string | null
}
