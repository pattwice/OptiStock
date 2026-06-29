package report

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var stockOnHandHeaders = []string{
	"Item Code", "Description", "Item Type", "LOT", "Supplier", "MFG Date", "EXP Date",
	"LOT Status", "Physical Qty", "Reserved Qty", "Available Qty",
}

var movementLedgerHeaders = []string{
	"Timestamp", "Transaction Type", "Item Code", "LOT", "Doc Ref", "Qty Changed", "Reason Code", "User",
}

var woSummaryHeaders = []string{
	"WO Number", "FG Code", "Target Qty", "Actual Produced Qty", "Completion %", "WO Status",
	"Created Date", "Completed/Cancelled Date",
}

var shortageDamageHeaders = []string{
	"Date", "WO/PO Ref", "Item Code", "Supplier", "Supplier Damage Qty", "Line Waste Qty",
}

var auditTrailHeaders = []string{
	"Timestamp", "WO Number", "User", "Action", "Old Value", "New Value",
}

var partialCompletionHeaders = []string{
	"WO Number", "FG Code", "Target Qty", "Actual at Request", "Completion %", "Requested By",
	"Requested At", "Approval Status", "Resolved By", "Resolved At", "Resolution Notes",
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) StockOnHand(ctx context.Context, f StockOnHandFilters) ([][]string, int, error) {
	nearClause := ""
	if f.NearExpiry {
		nearClause = `AND v.exp_date IS NOT NULL AND (v.exp_date - CURRENT_DATE) <= COALESCE(i.expiry_threshold_days, (
			SELECT COALESCE(NULLIF(config_value, '')::int, 30) FROM e1_system_config WHERE config_key = 'NEAR_EXPIRY_DAYS_DEFAULT'
		))`
	}
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM v_lot_available_stock v
		JOIN a1_items i ON i.item_code = v.item_code
		WHERE ($1 = '' OR i.item_type = $1)
		  AND ($2 = '' OR v.status = $2)
		  AND ($3 = '' OR v.item_code = $3)
		  %s
	`, nearClause)
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.ItemType, f.LotStatus, f.ItemCode).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := pageLimitOffset(f.Page, f.PageSize)
	dataSQL := fmt.Sprintf(`
		SELECT
			v.item_code,
			i.description,
			i.item_type,
			v.supplier_lot_number,
			COALESCE(v.supplier_name, ''),
			v.mfg_date::text,
			COALESCE(v.exp_date::text, ''),
			v.status,
			v.physical_qty::text,
			v.reserved_qty::text,
			v.available_qty::text
		FROM v_lot_available_stock v
		JOIN a1_items i ON i.item_code = v.item_code
		WHERE ($1 = '' OR i.item_type = $1)
		  AND ($2 = '' OR v.status = $2)
		  AND ($3 = '' OR v.item_code = $3)
		  %s
		ORDER BY v.item_code, v.exp_date ASC NULLS LAST
		LIMIT $4 OFFSET $5
	`, nearClause)
	rows, err := r.pool.Query(ctx, dataSQL, f.ItemType, f.LotStatus, f.ItemCode, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, _, err := scanStringRows(rows, 11)
	return out, total, err
}

func (r *Repository) MovementLedger(ctx context.Context, f MovementLedgerFilters) ([][]string, int, error) {
	countSQL := `
		SELECT COUNT(*)
		FROM b2_stock_ledger b
		JOIN b1_lots l ON l.lot_internal_id = b.lot_internal_id
		JOIN users u ON u.id = b.user_id
		WHERE ($1 = '' OR b.timestamp::date >= $1::date)
		  AND ($2 = '' OR b.timestamp::date <= $2::date)
		  AND ($3 = '' OR b.transaction_type = $3)
		  AND ($4 = '' OR l.item_code = $4)
		  AND ($5 = '' OR b.lot_internal_id::text = $5)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.FromDate, f.ToDate, f.TransactionType, f.ItemCode, f.LotInternalID).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := pageLimitOffset(f.Page, f.PageSize)
	dataSQL := `
		SELECT
			b.timestamp::text,
			b.transaction_type,
			l.item_code,
			l.supplier_lot_number,
			COALESCE(b.doc_ref, ''),
			b.qty_changed::text,
			COALESCE(b.reason_code, ''),
			u.name
		FROM b2_stock_ledger b
		JOIN b1_lots l ON l.lot_internal_id = b.lot_internal_id
		JOIN users u ON u.id = b.user_id
		WHERE ($1 = '' OR b.timestamp::date >= $1::date)
		  AND ($2 = '' OR b.timestamp::date <= $2::date)
		  AND ($3 = '' OR b.transaction_type = $3)
		  AND ($4 = '' OR l.item_code = $4)
		  AND ($5 = '' OR b.lot_internal_id::text = $5)
		ORDER BY b.timestamp DESC
		LIMIT $6 OFFSET $7
	`
	rows, err := r.pool.Query(ctx, dataSQL, f.FromDate, f.ToDate, f.TransactionType, f.ItemCode, f.LotInternalID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, _, err := scanStringRows(rows, 8)
	return out, total, err
}

func (r *Repository) WOSummary(ctx context.Context, f WOSummaryFilters) ([][]string, int, error) {
	countSQL := `
		SELECT COUNT(*)
		FROM c1_wo_header w
		WHERE ($1 = '' OR w.wo_status = $1)
		  AND ($2 = '' OR w.created_at::date >= $2::date)
		  AND ($3 = '' OR w.created_at::date <= $3::date)
		  AND ($4 = '' OR w.target_fg_code = $4)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.Status, f.FromDate, f.ToDate, f.TargetFGCode).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := pageLimitOffset(f.Page, f.PageSize)
	dataSQL := `
		SELECT
			w.wo_number,
			w.target_fg_code,
			w.target_qty::text,
			COALESCE(w.actual_produced_qty::text, ''),
			COALESCE(w.completion_pct::text, ''),
			w.wo_status,
			w.created_at::text,
			CASE
				WHEN w.wo_status IN ('COMPLETED', 'COMPLETED_PARTIAL', 'CANCELLED') THEN w.updated_at::text
				ELSE ''
			END
		FROM c1_wo_header w
		WHERE ($1 = '' OR w.wo_status = $1)
		  AND ($2 = '' OR w.created_at::date >= $2::date)
		  AND ($3 = '' OR w.created_at::date <= $3::date)
		  AND ($4 = '' OR w.target_fg_code = $4)
		ORDER BY w.created_at DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.pool.Query(ctx, dataSQL, f.Status, f.FromDate, f.ToDate, f.TargetFGCode, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, _, err := scanStringRows(rows, 8)
	return out, total, err
}

func (r *Repository) ShortageDamage(ctx context.Context, f ShortageDamageFilters) ([][]string, int, error) {
	countSQL := `
		SELECT COUNT(*)
		FROM b2_stock_ledger b
		JOIN b1_lots l ON l.lot_internal_id = b.lot_internal_id
		WHERE b.transaction_type = 'ADJ_OUT'
		  AND ($1 = '' OR b.timestamp::date >= $1::date)
		  AND ($2 = '' OR b.timestamp::date <= $2::date)
		  AND ($3 = '' OR b.reason_code = $3)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.FromDate, f.ToDate, f.ReasonCode).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := pageLimitOffset(f.Page, f.PageSize)
	dataSQL := `
		SELECT
			b.timestamp::date::text,
			COALESCE(b.doc_ref, ''),
			l.item_code,
			COALESCE(l.supplier_name, ''),
			CASE WHEN b.reason_code = 'SUPPLIER_DAMAGE' THEN ABS(b.qty_changed)::text ELSE '0' END,
			CASE WHEN b.reason_code = 'PRODUCTION_DAMAGE' OR b.reason_code = 'Line_Waste' THEN ABS(b.qty_changed)::text ELSE '0' END
		FROM b2_stock_ledger b
		JOIN b1_lots l ON l.lot_internal_id = b.lot_internal_id
		WHERE b.transaction_type = 'ADJ_OUT'
		  AND ($1 = '' OR b.timestamp::date >= $1::date)
		  AND ($2 = '' OR b.timestamp::date <= $2::date)
		  AND ($3 = '' OR b.reason_code = $3)
		ORDER BY b.timestamp DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, dataSQL, f.FromDate, f.ToDate, f.ReasonCode, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, _, err := scanStringRows(rows, 6)
	return out, total, err
}

func (r *Repository) AuditTrail(ctx context.Context, f AuditTrailFilters) ([][]string, int, error) {
	countSQL := `
		SELECT COUNT(*)
		FROM d1_wo_audit_log d
		JOIN users u ON u.id = d.user_id
		WHERE ($1 = '' OR d.wo_number = $1)
		  AND ($2 = '' OR d.action = $2)
		  AND ($3 = '' OR d.user_id::text = $3)
		  AND ($4 = '' OR d.timestamp::date >= $4::date)
		  AND ($5 = '' OR d.timestamp::date <= $5::date)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.WONumber, f.Action, f.UserID, f.FromDate, f.ToDate).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := pageLimitOffset(f.Page, f.PageSize)
	dataSQL := `
		SELECT
			d.timestamp::text,
			d.wo_number,
			u.name,
			d.action,
			COALESCE(d.old_value, ''),
			COALESCE(d.new_value, '')
		FROM d1_wo_audit_log d
		JOIN users u ON u.id = d.user_id
		WHERE ($1 = '' OR d.wo_number = $1)
		  AND ($2 = '' OR d.action = $2)
		  AND ($3 = '' OR d.user_id::text = $3)
		  AND ($4 = '' OR d.timestamp::date >= $4::date)
		  AND ($5 = '' OR d.timestamp::date <= $5::date)
		ORDER BY d.timestamp DESC
		LIMIT $6 OFFSET $7
	`
	rows, err := r.pool.Query(ctx, dataSQL, f.WONumber, f.Action, f.UserID, f.FromDate, f.ToDate, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, _, err := scanStringRows(rows, 6)
	return out, total, err
}

func (r *Repository) PartialCompletion(ctx context.Context, f PartialCompletionFilters) ([][]string, int, error) {
	countSQL := `
		SELECT COUNT(*)
		FROM d2_wo_approvals a
		JOIN c1_wo_header w ON w.wo_number = a.wo_number
		WHERE ($1 = '' OR a.requested_at::date >= $1::date)
		  AND ($2 = '' OR a.requested_at::date <= $2::date)
		  AND ($3 = '' OR a.approval_status = $3)
	`
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, f.FromDate, f.ToDate, f.ApprovalStatus).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := pageLimitOffset(f.Page, f.PageSize)
	dataSQL := `
		SELECT
			a.wo_number,
			w.target_fg_code,
			w.target_qty::text,
			COALESCE(w.actual_produced_qty::text, ''),
			a.completion_pct_at_request::text,
			ru.name,
			a.requested_at::text,
			a.approval_status,
			COALESCE(res.name, ''),
			COALESCE(a.resolved_at::text, ''),
			COALESCE(a.resolution_notes, '')
		FROM d2_wo_approvals a
		JOIN c1_wo_header w ON w.wo_number = a.wo_number
		JOIN users ru ON ru.id = a.requested_by
		LEFT JOIN users res ON res.id = a.resolved_by
		WHERE ($1 = '' OR a.requested_at::date >= $1::date)
		  AND ($2 = '' OR a.requested_at::date <= $2::date)
		  AND ($3 = '' OR a.approval_status = $3)
		ORDER BY a.requested_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, dataSQL, f.FromDate, f.ToDate, f.ApprovalStatus, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, _, err := scanStringRows(rows, 11)
	return out, total, err
}

func pageLimitOffset(page, pageSize int) (int, int) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	if page <= 0 {
		page = 1
	}
	return pageSize, (page - 1) * pageSize
}

func scanStringRows(rows pgx.Rows, cols int) ([][]string, int, error) {
	out := make([][]string, 0, 64)
	for rows.Next() {
		dest := make([]any, cols)
		ptrs := make([]*string, cols)
		for i := range ptrs {
			dest[i] = &ptrs[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, 0, err
		}
		row := make([]string, cols)
		for i, p := range ptrs {
			if p != nil {
				row[i] = *p
			}
		}
		out = append(out, row)
	}
	return out, len(out), rows.Err()
}

func HeadersStockOnHand() []string       { return stockOnHandHeaders }
func HeadersMovementLedger() []string    { return movementLedgerHeaders }
func HeadersWOSummary() []string         { return woSummaryHeaders }
func HeadersShortageDamage() []string    { return shortageDamageHeaders }
func HeadersAuditTrail() []string        { return auditTrailHeaders }
func HeadersPartialCompletion() []string { return partialCompletionHeaders }
