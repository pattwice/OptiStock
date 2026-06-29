CREATE TABLE a2_bom (
    bom_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_item_code TEXT NOT NULL REFERENCES a1_items(item_code) ON DELETE RESTRICT,
    component_item_code TEXT NOT NULL REFERENCES a1_items(item_code) ON DELETE RESTRICT,
    qty_per_set NUMERIC(18, 6) NOT NULL CHECK (qty_per_set > 0),
    bom_version TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_a2_bom_parent ON a2_bom(parent_item_code);
CREATE INDEX idx_a2_bom_component ON a2_bom(component_item_code);

-- Only one active BOM version per parent at any time.
CREATE UNIQUE INDEX uq_a2_bom_one_active_per_parent
ON a2_bom(parent_item_code)
WHERE is_active = TRUE;
