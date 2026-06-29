import type { WOStatus } from '../types/workorder'

export const WO_STATUSES: WOStatus[] = [
  'DRAFT',
  'RESERVED',
  'IN_PRODUCTION',
  'PENDING_APPROVAL',
  'COMPLETED',
  'COMPLETED_PARTIAL',
  'CANCELLED',
]

export const WO_STATUS_COLORS: Record<WOStatus, string> = {
  DRAFT: 'default',
  RESERVED: 'processing',
  IN_PRODUCTION: 'blue',
  PENDING_APPROVAL: 'orange',
  COMPLETED: 'success',
  COMPLETED_PARTIAL: 'gold',
  CANCELLED: 'error',
}
