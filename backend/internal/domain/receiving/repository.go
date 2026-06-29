package receiving

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type EnsureLotResult struct {
	LotInternalID string
	WasCreated    bool
}

func (r *Repository) EnsureLot(ctx context.Context, itemCode, supplierLotNumber string, supplierName *string, mfgDate string, expDate *string) (*EnsureLotResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Try to insert first (fast path).
	var lotID string
	insertErr := tx.QueryRow(ctx, `
		INSERT INTO b1_lots (item_code, supplier_lot_number, supplier_name, mfg_date, exp_date, status)
		VALUES ($1,$2,$3,$4::date,$5::date,'Active')
		RETURNING lot_internal_id::text
	`, itemCode, supplierLotNumber, supplierName, mfgDate, expDate).Scan(&lotID)

	if insertErr == nil {
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &EnsureLotResult{LotInternalID: lotID, WasCreated: true}, nil
	}

	// If unique violation, fetch existing and validate mismatch rules.
	var pgErr *pgconn.PgError
	if !(errors.As(insertErr, &pgErr) && pgErr.Code == "23505") {
		return nil, insertErr
	}

	var existingMfg string
	var existingExp *string
	row := tx.QueryRow(ctx, `
		SELECT lot_internal_id::text, mfg_date::text, exp_date::text
		FROM b1_lots
		WHERE item_code = $1 AND supplier_lot_number = $2
	`, itemCode, supplierLotNumber)
	if err := row.Scan(&lotID, &existingMfg, &existingExp); err != nil {
		return nil, err
	}

	// Re-receipt mismatch rule.
	if existingMfg != mfgDate {
		return nil, fmt.Errorf("lot mismatch: existing mfg_date=%s received mfg_date=%s", existingMfg, mfgDate)
	}
	if (existingExp == nil) != (expDate == nil) {
		return nil, fmt.Errorf("lot mismatch: existing exp_date=%v received exp_date=%v", existingExp, expDate)
	}
	if existingExp != nil && expDate != nil && *existingExp != *expDate {
		return nil, fmt.Errorf("lot mismatch: existing exp_date=%s received exp_date=%s", *existingExp, *expDate)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &EnsureLotResult{LotInternalID: lotID, WasCreated: false}, nil
}

func (r *Repository) InsertLedgerEntry(ctx context.Context, tx pgx.Tx, userID string, txType string, lotInternalID string, qtyChanged string, docRef string, reasonCode *string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO b2_stock_ledger (user_id, transaction_type, lot_internal_id, qty_changed, doc_ref, reason_code)
		VALUES ($1::uuid, $2, $3::uuid, $4::numeric, $5, $6)
	`, userID, txType, lotInternalID, qtyChanged, docRef, reasonCode)
	return err
}

func (r *Repository) LockLotForUpdate(ctx context.Context, tx pgx.Tx, lotInternalID string) error {
	// Lock the LOT row to serialize negative-stock checks per lot.
	_, err := tx.Exec(ctx, `SELECT 1 FROM b1_lots WHERE lot_internal_id = $1::uuid FOR UPDATE`, lotInternalID)
	return err
}

func (r *Repository) GetLotPhysicalQty(ctx context.Context, tx pgx.Tx, lotInternalID string) (string, error) {
	var physical string
	if err := tx.QueryRow(ctx, `
		SELECT physical_qty::text
		FROM v_lot_physical_stock
		WHERE lot_internal_id = $1::uuid
	`, lotInternalID).Scan(&physical); err != nil {
		return "", err
	}
	return physical, nil
}

