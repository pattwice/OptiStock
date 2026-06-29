export const REASON_CODES = [
  'Audited',
  'Line_Waste',
  'Supplier_Damage',
  'Expired',
  'Cycle_Count',
] as const

export type ReasonCode = (typeof REASON_CODES)[number]

export const ITEM_TYPES = ['FG', 'SFG', 'RM'] as const

export const LOT_STATUSES = ['Active', 'Hold', 'Quarantined'] as const

export const LOT_STATUS_COLORS: Record<string, string> = {
  Active: 'green',
  Hold: 'orange',
  Quarantined: 'red',
}
