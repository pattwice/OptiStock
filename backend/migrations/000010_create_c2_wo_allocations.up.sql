CREATE TABLE c2_wo_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    req_id UUID NOT NULL REFERENCES c1_5_wo_requirements(req_id) ON DELETE CASCADE,
    lot_internal_id UUID NOT NULL REFERENCES b1_lots(lot_internal_id) ON DELETE RESTRICT,
    reserved_qty NUMERIC(18, 6) NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
    actual_used_qty NUMERIC(18, 6) NULL,
    damage_qty NUMERIC(18, 6) NULL,
    UNIQUE (req_id, lot_internal_id)
);

CREATE INDEX idx_c2_req_id ON c2_wo_allocations(req_id);
CREATE INDEX idx_c2_lot ON c2_wo_allocations(lot_internal_id);
