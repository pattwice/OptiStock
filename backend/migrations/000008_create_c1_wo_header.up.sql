CREATE TABLE c1_wo_header (
    wo_number TEXT PRIMARY KEY,
    target_fg_code TEXT NOT NULL REFERENCES a1_items(item_code) ON DELETE RESTRICT,
    target_qty NUMERIC(18, 6) NOT NULL CHECK (target_qty > 0),
    actual_produced_qty NUMERIC(18, 6) NULL,
    completion_pct NUMERIC(8, 2) NULL,
    wo_status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (
        wo_status IN (
            'DRAFT', 'RESERVED', 'IN_PRODUCTION', 'PENDING_APPROVAL',
            'COMPLETED', 'COMPLETED_PARTIAL', 'CANCELLED'
        )
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_c1_wo_status ON c1_wo_header(wo_status);
CREATE INDEX idx_c1_wo_target_fg ON c1_wo_header(target_fg_code);
