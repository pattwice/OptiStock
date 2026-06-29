export type ItemType = 'FG' | 'SFG' | 'RM'
export type LotStatus = 'Active' | 'Hold' | 'Quarantined'

export interface Item {
  item_code: string
  description: string
  unit: string
  item_type: ItemType
  min_stock_level: string
  expiry_threshold_days?: number | null
  shelf_life_days?: number | null
  created_at: string
  updated_at: string
}

export interface BomRow {
  bom_id: string
  parent_item_code: string
  component_item_code: string
  qty_per_set: string
  bom_version: string
  is_active: boolean
  created_at: string
}

export interface Lot {
  lot_internal_id: string
  item_code: string
  supplier_lot_number: string
  supplier_name?: string | null
  mfg_date: string
  exp_date?: string | null
  status: LotStatus
  created_at: string
}

export interface LedgerEntry {
  transaction_id: string
  timestamp: string
  user_id: string
  transaction_type: string
  lot_internal_id: string
  item_code: string
  qty_changed: string
  doc_ref: string
  reason_code?: string | null
}

export interface StockOnHandRow {
  lot_internal_id: string
  item_code: string
  supplier_lot_number: string
  supplier_name?: string | null
  mfg_date: string
  exp_date?: string | null
  status: LotStatus
  physical_qty: string
  reserved_qty: string
  available_qty: string
}
