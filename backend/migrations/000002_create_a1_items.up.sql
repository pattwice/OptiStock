CREATE TABLE a1_items (
    item_code TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    unit TEXT NOT NULL,
    item_type TEXT NOT NULL CHECK (item_type IN ('FG', 'SFG', 'RM')),
    min_stock_level NUMERIC(18, 6) NOT NULL DEFAULT 0,
    expiry_threshold_days INTEGER NULL,
    shelf_life_days INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_a1_items_item_type ON a1_items(item_type);
