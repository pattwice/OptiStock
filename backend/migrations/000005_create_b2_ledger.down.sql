DROP TRIGGER IF EXISTS trg_b2_ledger_no_update ON b2_stock_ledger;
DROP TRIGGER IF EXISTS trg_b2_ledger_no_delete ON b2_stock_ledger;
DROP FUNCTION IF EXISTS b2_ledger_block_mutations;
DROP TABLE IF EXISTS b2_stock_ledger;
