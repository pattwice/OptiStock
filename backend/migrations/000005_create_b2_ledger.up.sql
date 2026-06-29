CREATE TABLE b2_stock_ledger (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('PO_RECEIPT', 'WO_ISSUE', 'WO_RECEIPT', 'RETURN_TO_STOCK', 'ADJ_IN', 'ADJ_OUT')),
    lot_internal_id UUID NOT NULL REFERENCES b1_lots(lot_internal_id) ON DELETE RESTRICT,
    qty_changed NUMERIC(18, 6) NOT NULL CHECK (qty_changed <> 0),
    doc_ref TEXT NOT NULL,
    reason_code TEXT NULL
);

CREATE INDEX idx_b2_ledger_timestamp ON b2_stock_ledger(timestamp);
CREATE INDEX idx_b2_ledger_lot ON b2_stock_ledger(lot_internal_id);
CREATE INDEX idx_b2_ledger_user ON b2_stock_ledger(user_id);
CREATE INDEX idx_b2_ledger_type ON b2_stock_ledger(transaction_type);

-- Adjustment reason codes must be provided.
ALTER TABLE b2_stock_ledger
ADD CONSTRAINT chk_b2_ledger_reason_for_adjustments
CHECK (
  (transaction_type IN ('ADJ_IN', 'ADJ_OUT') AND reason_code IS NOT NULL AND reason_code <> '')
  OR (transaction_type NOT IN ('ADJ_IN', 'ADJ_OUT'))
);

-- Append-only ledger safety net: block UPDATE/DELETE.
CREATE OR REPLACE FUNCTION b2_ledger_block_mutations()
RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'b2_stock_ledger is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_b2_ledger_no_update
BEFORE UPDATE ON b2_stock_ledger
FOR EACH ROW EXECUTE FUNCTION b2_ledger_block_mutations();

CREATE TRIGGER trg_b2_ledger_no_delete
BEFORE DELETE ON b2_stock_ledger
FOR EACH ROW EXECUTE FUNCTION b2_ledger_block_mutations();
