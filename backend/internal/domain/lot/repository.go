package lot

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateLot(ctx context.Context, input CreateLotInput) (*Lot, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO b1_lots (item_code, supplier_lot_number, supplier_name, mfg_date, exp_date, status)
		VALUES ($1,$2,$3,$4::date,$5::date,$6)
		RETURNING lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
	`,
		input.ItemCode, input.SupplierLotNumber, input.SupplierName, input.MfgDate, input.ExpDate, input.Status,
	)

	var l Lot
	if err := row.Scan(&l.LotInternalID, &l.ItemCode, &l.SupplierLotNumber, &l.SupplierName, &l.MfgDate, &l.ExpDate, &l.Status, &l.CreatedAt); err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *Repository) GetLot(ctx context.Context, lotInternalID string) (*Lot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
		FROM b1_lots
		WHERE lot_internal_id = $1::uuid
	`, lotInternalID)

	var l Lot
	if err := row.Scan(&l.LotInternalID, &l.ItemCode, &l.SupplierLotNumber, &l.SupplierName, &l.MfgDate, &l.ExpDate, &l.Status, &l.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *Repository) ListLots(ctx context.Context, itemCode, status string) ([]Lot, error) {
	var rows pgx.Rows
	var err error

	if itemCode != "" && status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
			FROM b1_lots
			WHERE item_code = $1 AND status = $2
			ORDER BY created_at DESC
		`, itemCode, status)
	} else if itemCode != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
			FROM b1_lots
			WHERE item_code = $1
			ORDER BY created_at DESC
		`, itemCode)
	} else if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
			FROM b1_lots
			WHERE status = $1
			ORDER BY created_at DESC
		`, status)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
			FROM b1_lots
			ORDER BY created_at DESC
		`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Lot, 0, 64)
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.LotInternalID, &l.ItemCode, &l.SupplierLotNumber, &l.SupplierName, &l.MfgDate, &l.ExpDate, &l.Status, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateLotStatus(ctx context.Context, lotInternalID, newStatus string) (*Lot, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE b1_lots
		SET status = $2
		WHERE lot_internal_id = $1::uuid
		RETURNING lot_internal_id::text, item_code, supplier_lot_number, supplier_name, mfg_date::text, exp_date::text, status, created_at
	`, lotInternalID, newStatus)

	var l Lot
	if err := row.Scan(&l.LotInternalID, &l.ItemCode, &l.SupplierLotNumber, &l.SupplierName, &l.MfgDate, &l.ExpDate, &l.Status, &l.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

