import { apiClient, type ApiEnvelope } from './client'
import type { ApprovalRequest } from '../types/approval'

async function unwrap<T>(promise: Promise<{ data: ApiEnvelope<T> }>) {
  const { data } = await promise
  return data.data
}

export function listPendingApprovals() {
  return unwrap<ApprovalRequest[]>(apiClient.get('/approvals'))
}

export function listWorkOrderApprovals(woNumber: string) {
  return unwrap<ApprovalRequest[]>(
    apiClient.get(`/workorders/${encodeURIComponent(woNumber)}/approvals`),
  )
}

export function withdrawApproval(approvalID: string) {
  return unwrap<ApprovalRequest>(apiClient.post(`/approvals/${approvalID}/withdraw`))
}

export function approveRequest(approvalID: string, resolution_notes?: string) {
  return unwrap<ApprovalRequest>(
    apiClient.post(`/approvals/${approvalID}/approve`, { resolution_notes }),
  )
}

export function rejectRequest(approvalID: string, resolution_notes: string) {
  return unwrap<ApprovalRequest>(
    apiClient.post(`/approvals/${approvalID}/reject`, { resolution_notes }),
  )
}
