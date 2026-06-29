CREATE TABLE d1_wo_audit_log (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wo_number TEXT NOT NULL REFERENCES c1_wo_header(wo_number) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action TEXT NOT NULL,
    old_value TEXT NULL,
    new_value TEXT NULL
);

CREATE INDEX idx_d1_wo_number ON d1_wo_audit_log(wo_number);
CREATE INDEX idx_d1_timestamp ON d1_wo_audit_log(timestamp);

CREATE OR REPLACE FUNCTION d1_wo_audit_log_block_mutations()
RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'd1_wo_audit_log is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_d1_audit_no_update
BEFORE UPDATE ON d1_wo_audit_log
FOR EACH ROW EXECUTE FUNCTION d1_wo_audit_log_block_mutations();

CREATE TRIGGER trg_d1_audit_no_delete
BEFORE DELETE ON d1_wo_audit_log
FOR EACH ROW EXECUTE FUNCTION d1_wo_audit_log_block_mutations();
