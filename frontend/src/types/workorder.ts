export type WOStatus =
  | 'DRAFT'
  | 'RESERVED'
  | 'IN_PRODUCTION'
  | 'PENDING_APPROVAL'
  | 'COMPLETED'
  | 'COMPLETED_PARTIAL'
  | 'CANCELLED'

export interface WorkOrder {
  wo_number: string
  target_fg_code: string
  target_qty: string
  actual_produced_qty?: string | null
  completion_pct?: string | null
  wo_status: WOStatus
  created_at: string
  updated_at: string
}

export interface WORequirement {
  req_id: string
  wo_number: string
  required_item_code: string
  total_needed_qty: string
  available_qty?: string
  shortage?: boolean
}

export interface WOAllocation {
  allocation_id: string
  req_id: string
  lot_internal_id: string
  reserved_qty: string
  actual_used_qty?: string | null
  damage_qty?: string | null
  item_code?: string
  supplier_lot_number?: string
  exp_date?: string | null
  mfg_date?: string
}

export interface WorkOrderDetail {
  work_order: WorkOrder
  requirements: WORequirement[]
  allocations: WOAllocation[]
}

export interface AllocationProposal {
  req_id: string
  required_item_code: string
  total_needed_qty: string
  proposed: Array<{
    req_id: string
    lot_internal_id: string
    reserved_qty: string
  }>
  shortage: boolean
  shortage_qty?: string
}

export interface AuditEntry {
  log_id: string
  wo_number: string
  user_id: string
  user_name?: string
  timestamp: string
  action: string
  old_value?: string | null
  new_value?: string | null
}
