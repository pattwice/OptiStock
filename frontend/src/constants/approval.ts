import type { ApprovalStatus } from '../types/approval'

export const APPROVAL_STATUS_COLORS: Record<ApprovalStatus, string> = {
  PENDING: 'gold',
  APPROVED: 'green',
  REJECTED: 'red',
  WITHDRAWN: 'default',
}
