package ledger

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type ListParams struct {
	ItemCode      string
	LotInternalID string
	From          *time.Time
	To            *time.Time
	Limit         int
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]LedgerEntry, error) {
	limit := p.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			sl.transaction_id::text,
			sl.timestamp,
			sl.user_id::text,
			sl.transaction_type,
			sl.lot_internal_id::text,
			l.item_code,
			sl.qty_changed::text,
			sl.doc_ref,
			sl.reason_code
		FROM b2_stock_ledger sl
		JOIN b1_lots l ON l.lot_internal_id = sl.lot_internal_id
		WHERE
			($1 = '' OR l.item_code = $1)
			AND ($2 = '' OR sl.lot_internal_id = $2::uuid)
			AND ($3::timestamptz IS NULL OR sl.timestamp >= $3::timestamptz)
			AND ($4::timestamptz IS NULL OR sl.timestamp <= $4::timestamptz)
		ORDER BY sl.timestamp DESC
		LIMIT $5
	`, p.ItemCode, p.LotInternalID, p.From, p.To, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]LedgerEntry, 0, limit)
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(
			&e.TransactionID,
			&e.Timestamp,
			&e.UserID,
			&e.TransactionType,
			&e.LotInternalID,
			&e.ItemCode,
			&e.QtyChanged,
			&e.DocRef,
			&e.ReasonCode,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

