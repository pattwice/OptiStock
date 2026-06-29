CREATE TABLE c1_5_wo_requirements (
    req_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wo_number TEXT NOT NULL REFERENCES c1_wo_header(wo_number) ON DELETE CASCADE,
    required_item_code TEXT NOT NULL REFERENCES a1_items(item_code) ON DELETE RESTRICT,
    total_needed_qty NUMERIC(18, 6) NOT NULL CHECK (total_needed_qty >= 0),
    UNIQUE (wo_number, required_item_code)
);

CREATE INDEX idx_c1_5_wo_number ON c1_5_wo_requirements(wo_number);
CREATE INDEX idx_c1_5_item_code ON c1_5_wo_requirements(required_item_code);
