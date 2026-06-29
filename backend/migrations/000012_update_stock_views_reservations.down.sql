DROP VIEW IF EXISTS v_item_available_stock;
DROP VIEW IF EXISTS v_lot_available_stock;

CREATE VIEW v_lot_available_stock AS
SELECT
  l.lot_internal_id,
  l.item_code,
  l.supplier_lot_number,
  l.supplier_name,
  l.mfg_date,
  l.exp_date,
  l.status,
  ps.physical_qty,
  0::numeric(18, 6) AS reserved_qty,
  ps.physical_qty AS available_qty
FROM b1_lots l
JOIN v_lot_physical_stock ps ON ps.lot_internal_id = l.lot_internal_id;

CREATE VIEW v_item_available_stock AS
SELECT
  item_code,
  COALESCE(SUM(available_qty), 0) AS available_qty
FROM v_lot_available_stock
GROUP BY item_code;
