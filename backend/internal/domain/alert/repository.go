package alert

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LowStockRow struct {
	ItemCode        string
	Description     string
	AvailableQty    string
	MinStockLevel   string
}

type NearExpiryRow struct {
	LotInternalID     string
	ItemCode          string
	Description       string
	SupplierLotNumber string
	ExpDate           string
	DaysUntilExpiry   int
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ConfigEnabled(ctx context.Context, key string) (bool, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT config_value FROM e1_system_config WHERE config_key = $1`, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return value == "TRUE" || value == "true" || value == "1", nil
}

func (r *Repository) NearExpiryDaysDefault(ctx context.Context) (int, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT config_value FROM e1_system_config WHERE config_key = 'NEAR_EXPIRY_DAYS_DEFAULT'`).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return 30, nil
	}
	if err != nil {
		return 0, err
	}
	var days int
	_, _ = fmt.Sscanf(value, "%d", &days)
	if days <= 0 {
		return 30, nil
	}
	return days, nil
}

func (r *Repository) CheckLowStock(ctx context.Context, itemCode string) (*LowStockRow, bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT i.item_code, i.description, COALESCE(s.available_qty, 0)::text, i.min_stock_level::text
		FROM a1_items i
		LEFT JOIN v_item_available_stock s ON s.item_code = i.item_code
		WHERE i.item_code = $1
	`, itemCode)
	var out LowStockRow
	if err := row.Scan(&out.ItemCode, &out.Description, &out.AvailableQty, &out.MinStockLevel); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	avail, ok1 := parseDecimal(out.AvailableQty)
	minLvl, ok2 := parseDecimal(out.MinStockLevel)
	if !ok1 || !ok2 {
		return nil, false, nil
	}
	if avail.Cmp(minLvl) < 0 {
		return &out, true, nil
	}
	return &out, false, nil
}

func (r *Repository) ItemCodeForLot(ctx context.Context, lotID string) (string, error) {
	var code string
	err := r.pool.QueryRow(ctx, `SELECT item_code FROM b1_lots WHERE lot_internal_id = $1::uuid`, lotID).Scan(&code)
	return code, err
}

func (r *Repository) CheckNearExpiryLot(ctx context.Context, lotID string) (*NearExpiryRow, bool, error) {
	defaultDays, err := r.NearExpiryDaysDefault(ctx)
	if err != nil {
		return nil, false, err
	}
	row := r.pool.QueryRow(ctx, `
		SELECT
			l.lot_internal_id::text,
			l.item_code,
			i.description,
			l.supplier_lot_number,
			l.exp_date::text,
			(l.exp_date - CURRENT_DATE)::int AS days_until
		FROM b1_lots l
		JOIN a1_items i ON i.item_code = l.item_code
		WHERE l.lot_internal_id = $1::uuid
		  AND l.status = 'Active'
		  AND l.exp_date IS NOT NULL
	`, lotID)
	var out NearExpiryRow
	if err := row.Scan(&out.LotInternalID, &out.ItemCode, &out.Description, &out.SupplierLotNumber, &out.ExpDate, &out.DaysUntilExpiry); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var threshold *int
	_ = r.pool.QueryRow(ctx, `SELECT expiry_threshold_days FROM a1_items WHERE item_code = $1`, out.ItemCode).Scan(&threshold)
	limit := defaultDays
	if threshold != nil {
		limit = *threshold
	}
	if out.DaysUntilExpiry <= limit {
		return &out, true, nil
	}
	return &out, false, nil
}

func (r *Repository) ListNearExpiryLots(ctx context.Context) ([]NearExpiryRow, error) {
	defaultDays, err := r.NearExpiryDaysDefault(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT
			l.lot_internal_id::text,
			l.item_code,
			i.description,
			l.supplier_lot_number,
			l.exp_date::text,
			(l.exp_date - CURRENT_DATE)::int AS days_until
		FROM b1_lots l
		JOIN a1_items i ON i.item_code = l.item_code
		WHERE l.status = 'Active'
		  AND l.exp_date IS NOT NULL
		  AND (l.exp_date - CURRENT_DATE) <= COALESCE(i.expiry_threshold_days, $1)
		ORDER BY l.exp_date ASC
	`, defaultDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]NearExpiryRow, 0, 32)
	for rows.Next() {
		var row NearExpiryRow
		if err := rows.Scan(&row.LotInternalID, &row.ItemCode, &row.Description, &row.SupplierLotNumber, &row.ExpDate, &row.DaysUntilExpiry); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *Repository) AffectedWOOwners(ctx context.Context, lotID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (w.wo_number) d.user_id::text
		FROM c2_wo_allocations c2
		JOIN c1_5_wo_requirements r ON r.req_id = c2.req_id
		JOIN c1_wo_header w ON w.wo_number = r.wo_number
		JOIN d1_wo_audit_log d ON d.wo_number = w.wo_number
		WHERE c2.lot_internal_id = $1::uuid
		  AND w.wo_status IN ('RESERVED', 'IN_PRODUCTION', 'PENDING_APPROVAL')
		ORDER BY w.wo_number, d.timestamp DESC
	`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0, 8)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) AffectedWONumbers(ctx context.Context, lotID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT w.wo_number
		FROM c2_wo_allocations c2
		JOIN c1_5_wo_requirements r ON r.req_id = c2.req_id
		JOIN c1_wo_header w ON w.wo_number = r.wo_number
		WHERE c2.lot_internal_id = $1::uuid
		  AND w.wo_status IN ('RESERVED', 'IN_PRODUCTION', 'PENDING_APPROVAL')
	`, lotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nums := make([]string, 0, 8)
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}
	return nums, rows.Err()
}
