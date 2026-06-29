CREATE TABLE b1_lots (
    lot_internal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_code TEXT NOT NULL REFERENCES a1_items(item_code) ON DELETE RESTRICT,
    supplier_lot_number TEXT NOT NULL,
    supplier_name TEXT NULL,
    mfg_date DATE NOT NULL,
    exp_date DATE NULL,
    status TEXT NOT NULL CHECK (status IN ('Active', 'Hold', 'Quarantined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Supplier_LOT_Number is unique per Item_Code.
CREATE UNIQUE INDEX uq_b1_lots_item_supplier_lot ON b1_lots(item_code, supplier_lot_number);

CREATE INDEX idx_b1_lots_item_code ON b1_lots(item_code);
CREATE INDEX idx_b1_lots_status ON b1_lots(status);
CREATE INDEX idx_b1_lots_exp_date ON b1_lots(exp_date);
