package workorder

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *Repository) CreateHeader(ctx context.Context, tx pgx.Tx, input CreateWOInput) (*WorkOrder, error) {
	row := tx.QueryRow(ctx, `
		INSERT INTO c1_wo_header (wo_number, target_fg_code, target_qty, wo_status)
		VALUES ($1, $2, $3::numeric, 'DRAFT')
		RETURNING wo_number, target_fg_code, target_qty::text, actual_produced_qty::text,
		          completion_pct::text, wo_status, created_at, updated_at
	`, input.WONumber, input.TargetFGCode, input.TargetQty)

	return scanWO(row)
}

func (r *Repository) InsertRequirement(ctx context.Context, tx pgx.Tx, woNumber, itemCode, qty string) (string, error) {
	var reqID string
	err := tx.QueryRow(ctx, `
		INSERT INTO c1_5_wo_requirements (wo_number, required_item_code, total_needed_qty)
		VALUES ($1, $2, $3::numeric)
		RETURNING req_id::text
	`, woNumber, itemCode, qty).Scan(&reqID)
	return reqID, err
}

func (r *Repository) GetItemType(ctx context.Context, itemCode string) (string, error) {
	var itemType string
	err := r.pool.QueryRow(ctx, `SELECT item_type FROM a1_items WHERE item_code = $1`, itemCode).Scan(&itemType)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return itemType, err
}

func (r *Repository) GetItemShelfLife(ctx context.Context, itemCode string) (*int, error) {
	var shelf *int
	err := r.pool.QueryRow(ctx, `SELECT shelf_life_days FROM a1_items WHERE item_code = $1`, itemCode).Scan(&shelf)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return shelf, err
}

type ActiveBOMRow struct {
	ComponentItemCode string
	QtyPerSet         string
	ComponentType     string
}

func (r *Repository) ListActiveBOM(ctx context.Context, parentItemCode string) ([]ActiveBOMRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.component_item_code, b.qty_per_set::text, a.item_type
		FROM a2_bom b
		JOIN a1_items a ON a.item_code = b.component_item_code
		WHERE b.parent_item_code = $1 AND b.is_active = TRUE
		ORDER BY b.component_item_code
	`, parentItemCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ActiveBOMRow, 0, 16)
	for rows.Next() {
		var row ActiveBOMRow
		if err := rows.Scan(&row.ComponentItemCode, &row.QtyPerSet, &row.ComponentType); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *Repository) HasActiveBOM(ctx context.Context, parentItemCode string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM a2_bom WHERE parent_item_code = $1 AND is_active = TRUE
		)
	`, parentItemCode).Scan(&exists)
	return exists, err
}

func (r *Repository) ListWorkOrders(ctx context.Context, status string) ([]WorkOrder, error) {
	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT wo_number, target_fg_code, target_qty::text, actual_produced_qty::text,
			       completion_pct::text, wo_status, created_at, updated_at
			FROM c1_wo_header
			WHERE wo_status = $1
			ORDER BY created_at DESC
		`, status)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT wo_number, target_fg_code, target_qty::text, actual_produced_qty::text,
			       completion_pct::text, wo_status, created_at, updated_at
			FROM c1_wo_header
			ORDER BY created_at DESC
		`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]WorkOrder, 0, 32)
	for rows.Next() {
		wo, err := scanWOFromRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *wo)
	}
	return out, rows.Err()
}

func (r *Repository) GetHeader(ctx context.Context, woNumber string) (*WorkOrder, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT wo_number, target_fg_code, target_qty::text, actual_produced_qty::text,
		       completion_pct::text, wo_status, created_at, updated_at
		FROM c1_wo_header
		WHERE wo_number = $1
	`, woNumber)
	wo, err := scanWO(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return wo, err
}

func (r *Repository) GetHeaderForUpdate(ctx context.Context, tx pgx.Tx, woNumber string) (*WorkOrder, error) {
	row := tx.QueryRow(ctx, `
		SELECT wo_number, target_fg_code, target_qty::text, actual_produced_qty::text,
		       completion_pct::text, wo_status, created_at, updated_at
		FROM c1_wo_header
		WHERE wo_number = $1
		FOR UPDATE
	`, woNumber)
	wo, err := scanWO(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return wo, err
}

func (r *Repository) ListRequirements(ctx context.Context, woNumber string) ([]Requirement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.req_id::text, r.wo_number, r.required_item_code, r.total_needed_qty::text,
		       COALESCE(ias.available_qty, 0)::text
		FROM c1_5_wo_requirements r
		LEFT JOIN v_item_available_stock ias ON ias.item_code = r.required_item_code
		WHERE r.wo_number = $1
		ORDER BY r.required_item_code
	`, woNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Requirement, 0, 32)
	for rows.Next() {
		var req Requirement
		if err := rows.Scan(&req.ReqID, &req.WONumber, &req.RequiredItemCode, &req.TotalNeededQty, &req.AvailableQty); err != nil {
			return nil, err
		}
		if cmp, ok := cmpRat(req.AvailableQty, req.TotalNeededQty); ok && cmp < 0 {
			req.Shortage = true
		}
		out = append(out, req)
	}
	return out, rows.Err()
}

func (r *Repository) ListAllocations(ctx context.Context, woNumber string) ([]Allocation, error) {
	return r.listAllocationsQuery(ctx, r.pool, woNumber)
}

func (r *Repository) ListAllocationsTx(ctx context.Context, tx pgx.Tx, woNumber string) ([]Allocation, error) {
	return r.listAllocationsQuery(ctx, tx, woNumber)
}

func (r *Repository) listAllocationsQuery(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, woNumber string) ([]Allocation, error) {
	rows, err := q.Query(ctx, `
		SELECT c2.allocation_id::text, c2.req_id::text, c2.lot_internal_id::text,
		       c2.reserved_qty::text, c2.actual_used_qty::text, c2.damage_qty::text,
		       l.item_code, l.supplier_lot_number, l.exp_date::text, l.mfg_date::text
		FROM c2_wo_allocations c2
		JOIN c1_5_wo_requirements r ON r.req_id = c2.req_id
		JOIN b1_lots l ON l.lot_internal_id = c2.lot_internal_id
		WHERE r.wo_number = $1
		ORDER BY r.required_item_code, l.exp_date ASC NULLS LAST, l.mfg_date ASC
	`, woNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Allocation, 0, 64)
	for rows.Next() {
		var a Allocation
		if err := rows.Scan(
			&a.AllocationID, &a.ReqID, &a.LOTInternalID, &a.ReservedQty,
			&a.ActualUsedQty, &a.DamageQty, &a.ItemCode, &a.SupplierLot, &a.ExpDate, &a.MfgDate,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type EligibleLot struct {
	LOTInternalID string
	AvailableQty  string
	ExpDate       *string
	MfgDate       string
}

func (r *Repository) ListEligibleLotsForUpdate(ctx context.Context, tx pgx.Tx, itemCode string) ([]EligibleLot, error) {
	rows, err := tx.Query(ctx, `
		SELECT
			l.lot_internal_id::text,
			(ps.physical_qty - COALESCE(res.reserved_qty, 0))::text AS available_qty,
			l.exp_date::text,
			l.mfg_date::text
		FROM b1_lots l
		JOIN v_lot_physical_stock ps ON ps.lot_internal_id = l.lot_internal_id
		LEFT JOIN (
			SELECT c2.lot_internal_id, SUM(c2.reserved_qty) AS reserved_qty
			FROM c2_wo_allocations c2
			JOIN c1_5_wo_requirements req ON req.req_id = c2.req_id
			JOIN c1_wo_header w ON w.wo_number = req.wo_number
			WHERE w.wo_status IN ('RESERVED', 'IN_PRODUCTION', 'PENDING_APPROVAL')
			GROUP BY c2.lot_internal_id
		) res ON res.lot_internal_id = l.lot_internal_id
		WHERE l.item_code = $1
		  AND l.status = 'Active'
		  AND (ps.physical_qty - COALESCE(res.reserved_qty, 0)) > 0
		ORDER BY l.exp_date ASC NULLS LAST, l.mfg_date ASC
		FOR UPDATE OF l SKIP LOCKED
	`, itemCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EligibleLot, 0, 32)
	for rows.Next() {
		var lot EligibleLot
		if err := rows.Scan(&lot.LOTInternalID, &lot.AvailableQty, &lot.ExpDate, &lot.MfgDate); err != nil {
			return nil, err
		}
		out = append(out, lot)
	}
	return out, rows.Err()
}

func (r *Repository) UpsertAllocation(ctx context.Context, tx pgx.Tx, reqID, lotID, qty string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO c2_wo_allocations (req_id, lot_internal_id, reserved_qty)
		VALUES ($1::uuid, $2::uuid, $3::numeric)
		ON CONFLICT (req_id, lot_internal_id)
		DO UPDATE SET reserved_qty = c2_wo_allocations.reserved_qty + EXCLUDED.reserved_qty
	`, reqID, lotID, qty)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, tx pgx.Tx, woNumber string, status WOStatus) error {
	_, err := tx.Exec(ctx, `
		UPDATE c1_wo_header SET wo_status = $2, updated_at = NOW() WHERE wo_number = $1
	`, woNumber, status)
	return err
}

func (r *Repository) InsertAudit(ctx context.Context, tx pgx.Tx, woNumber, userID, action, oldVal, newVal string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO d1_wo_audit_log (wo_number, user_id, action, old_value, new_value)
		VALUES ($1, $2::uuid, $3, $4, $5)
	`, woNumber, userID, action, nullIfEmpty(oldVal), nullIfEmpty(newVal))
	return err
}

func (r *Repository) ListAudit(ctx context.Context, woNumber string) ([]AuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.log_id::text, d.wo_number, d.user_id::text, u.name, d.timestamp, d.action,
		       d.old_value, d.new_value
		FROM d1_wo_audit_log d
		JOIN users u ON u.id = d.user_id
		WHERE d.wo_number = $1
		ORDER BY d.timestamp ASC
	`, woNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AuditEntry, 0, 64)
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.LogID, &e.WONumber, &e.UserID, &e.UserName, &e.Timestamp, &e.Action, &e.OldValue, &e.NewValue); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) ClearAllocations(ctx context.Context, tx pgx.Tx, woNumber string) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM c2_wo_allocations c2
		USING c1_5_wo_requirements r
		WHERE c2.req_id = r.req_id AND r.wo_number = $1
	`, woNumber)
	return err
}

func (r *Repository) ZeroReservations(ctx context.Context, tx pgx.Tx, woNumber string) error {
	_, err := tx.Exec(ctx, `
		UPDATE c2_wo_allocations c2
		SET reserved_qty = 0
		FROM c1_5_wo_requirements r
		WHERE c2.req_id = r.req_id AND r.wo_number = $1
	`, woNumber)
	return err
}

func (r *Repository) UpdateActuals(ctx context.Context, tx pgx.Tx, allocationID, actualUsed string, damage *string) error {
	_, err := tx.Exec(ctx, `
		UPDATE c2_wo_allocations
		SET actual_used_qty = $2::numeric,
		    damage_qty = CASE WHEN $3::boolean THEN $4::numeric ELSE damage_qty END
		WHERE allocation_id = $1::uuid
	`, allocationID, actualUsed, damage != nil, derefStr(damage))
	return err
}

func (r *Repository) UpdateHeaderActuals(ctx context.Context, tx pgx.Tx, woNumber, actualProduced, completionPct string) error {
	_, err := tx.Exec(ctx, `
		UPDATE c1_wo_header
		SET actual_produced_qty = $2::numeric,
		    completion_pct = $3::numeric,
		    updated_at = NOW()
		WHERE wo_number = $1
	`, woNumber, actualProduced, completionPct)
	return err
}

func (r *Repository) GetRequirementByID(ctx context.Context, reqID string) (*Requirement, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT req_id::text, wo_number, required_item_code, total_needed_qty::text, '0'
		FROM c1_5_wo_requirements WHERE req_id = $1::uuid
	`, reqID)
	var req Requirement
	if err := row.Scan(&req.ReqID, &req.WONumber, &req.RequiredItemCode, &req.TotalNeededQty, &req.AvailableQty); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *Repository) GetLotItemCode(ctx context.Context, tx pgx.Tx, lotID string) (string, error) {
	var code string
	err := tx.QueryRow(ctx, `SELECT item_code FROM b1_lots WHERE lot_internal_id = $1::uuid`, lotID).Scan(&code)
	return code, err
}

func (r *Repository) GetLotPhysicalQty(ctx context.Context, tx pgx.Tx, lotID string) (string, error) {
	var qty string
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(qty_changed), 0)::text
		FROM b2_stock_ledger WHERE lot_internal_id = $1::uuid
	`, lotID).Scan(&qty)
	return qty, err
}

func (r *Repository) LockLot(ctx context.Context, tx pgx.Tx, lotID string) error {
	_, err := tx.Exec(ctx, `SELECT 1 FROM b1_lots WHERE lot_internal_id = $1::uuid FOR UPDATE`, lotID)
	return err
}

func (r *Repository) InsertLedger(ctx context.Context, tx pgx.Tx, userID, txType, lotID, qty, docRef string, reason *string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO b2_stock_ledger (user_id, transaction_type, lot_internal_id, qty_changed, doc_ref, reason_code)
		VALUES ($1::uuid, $2, $3::uuid, $4::numeric, $5, $6)
	`, userID, txType, lotID, qty, docRef, reason)
	return err
}

func (r *Repository) CreateFGLot(ctx context.Context, tx pgx.Tx, itemCode, woNumber, mfgDate string, expDate *string) (string, error) {
	var lotID string
	err := tx.QueryRow(ctx, `
		INSERT INTO b1_lots (item_code, supplier_lot_number, supplier_name, mfg_date, exp_date, status)
		VALUES ($1, $2, NULL, $3::date, $4::date, 'Active')
		RETURNING lot_internal_id::text
	`, itemCode, woNumber, mfgDate, expDate).Scan(&lotID)
	return lotID, err
}

func (r *Repository) GetAllocation(ctx context.Context, allocationID string) (*Allocation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT c2.allocation_id::text, c2.req_id::text, c2.lot_internal_id::text,
		       c2.reserved_qty::text, c2.actual_used_qty::text, c2.damage_qty::text,
		       l.item_code, l.supplier_lot_number, l.exp_date::text, l.mfg_date::text
		FROM c2_wo_allocations c2
		JOIN b1_lots l ON l.lot_internal_id = c2.lot_internal_id
		WHERE c2.allocation_id = $1::uuid
	`, allocationID)
	var a Allocation
	if err := row.Scan(
		&a.AllocationID, &a.ReqID, &a.LOTInternalID, &a.ReservedQty,
		&a.ActualUsedQty, &a.DamageQty, &a.ItemCode, &a.SupplierLot, &a.ExpDate, &a.MfgDate,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *Repository) SumReservedForReq(ctx context.Context, tx pgx.Tx, reqID string) (string, error) {
	var sum string
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(reserved_qty), 0)::text FROM c2_wo_allocations WHERE req_id = $1::uuid
	`, reqID).Scan(&sum)
	return sum, err
}

func scanWO(row pgx.Row) (*WorkOrder, error) {
	var wo WorkOrder
	if err := row.Scan(
		&wo.WONumber, &wo.TargetFGCode, &wo.TargetQty, &wo.ActualProducedQty,
		&wo.CompletionPct, &wo.WOStatus, &wo.CreatedAt, &wo.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &wo, nil
}

func scanWOFromRows(rows pgx.Rows) (*WorkOrder, error) {
	var wo WorkOrder
	if err := rows.Scan(
		&wo.WONumber, &wo.TargetFGCode, &wo.TargetQty, &wo.ActualProducedQty,
		&wo.CompletionPct, &wo.WOStatus, &wo.CreatedAt, &wo.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &wo, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}

func statusLabel(old, new WOStatus) string {
	return fmt.Sprintf("%s → %s", old, new)
}
