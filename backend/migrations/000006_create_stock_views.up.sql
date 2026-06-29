-- Physical stock per lot: sum of all ledger movements.
CREATE OR REPLACE VIEW v_lot_physical_stock AS
SELECT
  l.lot_internal_id,
  l.item_code,
  COALESCE(SUM(sl.qty_changed), 0) AS physical_qty
FROM b1_lots l
LEFT JOIN b2_stock_ledger sl ON sl.lot_internal_id = l.lot_internal_id
GROUP BY l.lot_internal_id, l.item_code;

-- Reservations (Phase 1): no workorder tables yet; reserved is 0.
-- This placeholder keeps the view stable until Phase 2 introduces C2 reservations.
CREATE OR REPLACE VIEW v_lot_available_stock AS
SELECT
  l.lot_internal_id,
  l.item_code,
  l.supplier_lot_number,
  l.supplier_name,
  l.mfg_date,
  l.exp_date,
  l.status,
  ps.physical_qty,
  0::numeric(18,6) AS reserved_qty,
  ps.physical_qty AS available_qty
FROM b1_lots l
JOIN v_lot_physical_stock ps ON ps.lot_internal_id = l.lot_internal_id;

CREATE OR REPLACE VIEW v_item_available_stock AS
SELECT
  item_code,
  COALESCE(SUM(available_qty), 0) AS available_qty
FROM v_lot_available_stock
GROUP BY item_code;
