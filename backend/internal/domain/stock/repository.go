package stock

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListOnHand(ctx context.Context, itemCode, status string) ([]OnHandRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			lot_internal_id::text,
			item_code,
			supplier_lot_number,
			supplier_name,
			mfg_date::text,
			exp_date::text,
			status,
			physical_qty::text,
			reserved_qty::text,
			available_qty::text
		FROM v_lot_available_stock
		WHERE ($1 = '' OR item_code = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY item_code ASC, exp_date ASC NULLS LAST, mfg_date ASC
	`, itemCode, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OnHandRow, 0, 64)
	for rows.Next() {
		var row OnHandRow
		if err := rows.Scan(
			&row.LotInternalID,
			&row.ItemCode,
			&row.SupplierLotNumber,
			&row.SupplierName,
			&row.MfgDate,
			&row.ExpDate,
			&row.Status,
			&row.PhysicalQty,
			&row.ReservedQty,
			&row.AvailableQty,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
