CREATE TABLE d2_wo_approvals (
    approval_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wo_number TEXT NOT NULL REFERENCES c1_wo_header(wo_number) ON DELETE CASCADE,
    approval_type TEXT NOT NULL CHECK (approval_type IN ('PARTIAL_COMPLETION')),
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completion_pct_at_request NUMERIC(18, 6) NOT NULL,
    approval_status TEXT NOT NULL CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED', 'WITHDRAWN')),
    resolved_by UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    resolved_at TIMESTAMPTZ NULL,
    resolution_notes TEXT NULL
);

CREATE INDEX idx_d2_wo_number ON d2_wo_approvals(wo_number);
CREATE INDEX idx_d2_approval_status ON d2_wo_approvals(approval_status);
CREATE INDEX idx_d2_requested_at ON d2_wo_approvals(requested_at);
