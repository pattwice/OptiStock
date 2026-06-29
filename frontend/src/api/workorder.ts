import { apiClient, type ApiEnvelope } from './client'
import type {
  AllocationProposal,
  AuditEntry,
  WorkOrder,
  WorkOrderDetail,
} from '../types/workorder'
import type { WOStatus } from '../types/workorder'

async function unwrap<T>(promise: Promise<{ data: ApiEnvelope<T> }>) {
  const { data } = await promise
  return data.data
}

export function listWorkOrders(status?: WOStatus) {
  return unwrap<WorkOrder[]>(
    apiClient.get('/workorders', { params: status ? { status } : undefined }),
  )
}

export function createWorkOrder(body: {
  wo_number: string
  target_fg_code: string
  target_qty: string
}) {
  return unwrap<WorkOrderDetail>(apiClient.post('/workorders', body))
}

export function getWorkOrder(woNumber: string) {
  return unwrap<WorkOrderDetail>(apiClient.get(`/workorders/${encodeURIComponent(woNumber)}`))
}

export function getAllocationProposal(woNumber: string) {
  return unwrap<AllocationProposal[]>(
    apiClient.get(`/workorders/${encodeURIComponent(woNumber)}/allocation-proposal`),
  )
}

export function reserveWorkOrder(
  woNumber: string,
  body: {
    allocations?: Array<{ req_id: string; lot_internal_id: string; reserved_qty: string }>
    manual_override?: boolean
  } = {},
) {
  return unwrap<WorkOrderDetail>(
    apiClient.post(`/workorders/${encodeURIComponent(woNumber)}/reserve`, body),
  )
}

export function startWorkOrder(woNumber: string) {
  return unwrap<WorkOrderDetail>(
    apiClient.post(`/workorders/${encodeURIComponent(woNumber)}/start`),
  )
}

export function updateWorkOrderActuals(
  woNumber: string,
  body: {
    actual_produced_qty: string
    lines: Array<{ allocation_id: string; actual_used_qty: string; damage_qty?: string }>
  },
) {
  return unwrap<WorkOrderDetail>(
    apiClient.patch(`/workorders/${encodeURIComponent(woNumber)}/actuals`, body),
  )
}

export function completeWorkOrder(
  woNumber: string,
  body: {
    actual_produced_qty: string
    lines: Array<{ allocation_id: string; actual_used_qty: string; damage_qty?: string }>
  },
) {
  return unwrap<WorkOrderDetail>(
    apiClient.post(`/workorders/${encodeURIComponent(woNumber)}/complete`, body),
  )
}

export function submitWorkOrderForApproval(
  woNumber: string,
  body: {
    actual_produced_qty: string
    lines: Array<{ allocation_id: string; actual_used_qty: string; damage_qty?: string }>
  },
) {
  return unwrap<WorkOrderDetail>(
    apiClient.post(`/workorders/${encodeURIComponent(woNumber)}/submit-for-approval`, body),
  )
}

export function cancelWorkOrder(woNumber: string) {
  return unwrap<WorkOrderDetail>(
    apiClient.post(`/workorders/${encodeURIComponent(woNumber)}/cancel`),
  )
}

export function reopenWorkOrder(woNumber: string) {
  return unwrap<WorkOrderDetail>(
    apiClient.post(`/workorders/${encodeURIComponent(woNumber)}/reopen`),
  )
}

export function getWorkOrderAudit(woNumber: string) {
  return unwrap<AuditEntry[]>(
    apiClient.get(`/workorders/${encodeURIComponent(woNumber)}/audit`),
  )
}
