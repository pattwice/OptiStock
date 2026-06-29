-- Include C2 reservations in available stock calculations.
-- DROP + CREATE: PostgreSQL cannot change view column types via CREATE OR REPLACE.
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
  COALESCE(res.reserved_qty, 0::numeric(18, 6))::numeric(18, 6) AS reserved_qty,
  ps.physical_qty - COALESCE(res.reserved_qty, 0::numeric(18, 6)) AS available_qty
FROM b1_lots l
JOIN v_lot_physical_stock ps ON ps.lot_internal_id = l.lot_internal_id
LEFT JOIN (
  SELECT
    c2.lot_internal_id,
    SUM(c2.reserved_qty)::numeric(18, 6) AS reserved_qty
  FROM c2_wo_allocations c2
  JOIN c1_5_wo_requirements r ON r.req_id = c2.req_id
  JOIN c1_wo_header w ON w.wo_number = r.wo_number
  WHERE w.wo_status IN ('RESERVED', 'IN_PRODUCTION', 'PENDING_APPROVAL')
  GROUP BY c2.lot_internal_id
) res ON res.lot_internal_id = l.lot_internal_id;

CREATE VIEW v_item_available_stock AS
SELECT
  item_code,
  COALESCE(SUM(available_qty), 0) AS available_qty
FROM v_lot_available_stock
GROUP BY item_code;
