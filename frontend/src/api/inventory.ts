import { apiClient, type ApiEnvelope } from './client'
import type { BomRow, Item, ItemType, LedgerEntry, Lot, LotStatus, StockOnHandRow } from '../types/inventory'

async function unwrap<T>(promise: Promise<{ data: ApiEnvelope<T> }>) {
  const { data } = await promise
  return data.data
}

export function listItems(type?: ItemType) {
  return unwrap<Item[]>(
    apiClient.get('/items', { params: type ? { type } : undefined }),
  )
}

export function createItem(body: {
  item_code: string
  description: string
  unit: string
  item_type: ItemType
  min_stock_level: string
  expiry_threshold_days?: number | null
  shelf_life_days?: number | null
}) {
  return unwrap<Item>(apiClient.post('/items', body))
}

export function updateItem(
  itemCode: string,
  body: Partial<{
    description: string
    unit: string
    item_type: ItemType
    min_stock_level: string
    expiry_threshold_days: number | null
    shelf_life_days: number | null
  }>,
) {
  return unwrap<Item>(apiClient.patch(`/items/${encodeURIComponent(itemCode)}`, body))
}

export function listBom(parentItemCode: string) {
  return unwrap<BomRow[]>(apiClient.get('/bom', { params: { parent_item_code: parentItemCode } }))
}

export function createBomRow(body: {
  parent_item_code: string
  component_item_code: string
  qty_per_set: string
  bom_version: string
  is_active?: boolean
}) {
  return unwrap<BomRow>(apiClient.post('/bom', body))
}

export function activateBomVersion(parentItemCode: string, bomVersion: string) {
  return unwrap<{ status: string }>(
    apiClient.post('/bom/activate', { parent_item_code: parentItemCode, bom_version: bomVersion }),
  )
}

export function listLots(params?: { item_code?: string; status?: LotStatus }) {
  return unwrap<Lot[]>(apiClient.get('/lots', { params }))
}

export function updateLotStatus(lotInternalId: string, status: LotStatus) {
  return unwrap<Lot>(apiClient.patch(`/lots/${lotInternalId}/status`, { status }))
}

export function listLedger(params?: { item_code?: string; lot_internal_id?: string; limit?: number }) {
  return unwrap<LedgerEntry[]>(apiClient.get('/ledger', { params }))
}

export function poReceipt(body: {
  item_code: string
  supplier_lot_number: string
  supplier_name?: string
  mfg_date: string
  exp_date?: string | null
  qty_received: string
  doc_ref: string
  damage_qty?: string
  damage_reason_code?: string
}) {
  return unwrap<{ lot_internal_id: string; item_code: string; posted_entries: number }>(
    apiClient.post('/receiving/po', body),
  )
}

export function adjustment(body: {
  lot_internal_id: string
  type: 'ADJ_IN' | 'ADJ_OUT'
  qty: string
  doc_ref: string
  reason_code: string
}) {
  return unwrap<{ status: string }>(apiClient.post('/receiving/adjustment', body))
}

export function listStockOnHand(params?: { item_code?: string; status?: LotStatus }) {
  return unwrap<StockOnHandRow[]>(apiClient.get('/stock/on-hand', { params }))
}
