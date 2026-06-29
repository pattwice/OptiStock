export interface AlertMessage {
  id: string
  type: string
  severity: string
  payload: Record<string, unknown>
  timestamp: string
  read: boolean
}

export function alertTitle(alert: AlertMessage): string {
  switch (alert.type) {
    case 'LOW_STOCK':
      return `Low stock: ${String(alert.payload.item_code ?? '')}`
    case 'NEAR_EXPIRY':
      return `Near expiry: ${String(alert.payload.item_code ?? '')}`
    case 'LOT_ON_HOLD':
      return `LOT on hold`
    case 'LOT_QUARANTINED':
      return `LOT quarantined`
    case 'OVER_RESERVATION_BLOCKED':
      return `Reservation blocked: ${String(alert.payload.wo_number ?? '')}`
    case 'APPROVAL_PENDING':
      return `Approval pending: ${String(alert.payload.wo_number ?? '')}`
    case 'APPROVAL_WITHDRAWN':
      return `Approval withdrawn: ${String(alert.payload.wo_number ?? '')}`
    default:
      return alert.type
  }
}

export function alertDescription(alert: AlertMessage): string {
  const p = alert.payload
  switch (alert.type) {
    case 'LOW_STOCK':
      return `Available ${String(p.available_qty ?? '')} below min ${String(p.min_stock_level ?? '')}`
    case 'NEAR_EXPIRY':
      return `LOT ${String(p.supplier_lot_number ?? '')} expires ${String(p.exp_date ?? '')}`
    case 'APPROVAL_PENDING':
      return `${String(p.target_fg_code ?? '')} at ${String(p.completion_pct ?? '')}% by ${String(p.requester_name ?? '')}`
    default:
      return JSON.stringify(p)
  }
}
