package receiving

import (
	"context"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5"
	"optistock/internal/domain/alert"
	"optistock/pkg/apperror"
)

const supplierDamageReasonCode = "SUPPLIER_DAMAGE"

type Service struct {
	repo *Repository
	pool interface {
		Begin(context.Context) (pgx.Tx, error)
	}
	alerts alert.Notifier
}

func NewService(repo *Repository, alerts alert.Notifier) *Service {
	return &Service{repo: repo, pool: repo.pool, alerts: alerts}
}

func (s *Service) POReceipt(ctx context.Context, userID string, input POReceiptInput) (*POReceiptResult, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}
	if input.ItemCode == "" || input.SupplierLotNumber == "" || input.MfgDate == "" || input.QtyReceived == "" || input.DocRef == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "item_code, supplier_lot_number, mfg_date, qty_received, doc_ref are required")
	}

	ensure, err := s.repo.EnsureLot(ctx, input.ItemCode, input.SupplierLotNumber, input.SupplierName, input.MfgDate, input.ExpDate)
	if err != nil {
		return nil, apperror.WithMessage(apperror.ErrValidation, err.Error())
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	posted := 0
	if err := s.repo.InsertLedgerEntry(ctx, tx, userID, "PO_RECEIPT", ensure.LotInternalID, input.QtyReceived, input.DocRef, nil); err != nil {
		return nil, err
	}
	posted++

	if input.DamageQty != nil && strings.TrimSpace(*input.DamageQty) != "" {
		reason := supplierDamageReasonCode
		if input.DamageReasonCode != nil && strings.TrimSpace(*input.DamageReasonCode) != "" {
			reason = strings.TrimSpace(*input.DamageReasonCode)
		}

		qtyOut, err := negateNumericString(*input.DamageQty)
		if err != nil {
			return nil, apperror.WithMessage(apperror.ErrValidation, "invalid damage_qty")
		}
		if err := s.applyNegativeStockGuard(ctx, tx, ensure.LotInternalID, qtyOut); err != nil {
			return nil, err
		}
		if err := s.repo.InsertLedgerEntry(ctx, tx, userID, "ADJ_OUT", ensure.LotInternalID, qtyOut, input.DocRef, &reason); err != nil {
			return nil, err
		}
		posted++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	if s.alerts != nil {
		s.alerts.AfterLedgerWrite(ctx, input.ItemCode)
		s.alerts.AfterLotChange(ctx, ensure.LotInternalID)
	}
	return &POReceiptResult{LotInternalID: ensure.LotInternalID, ItemCode: input.ItemCode, Posted: posted}, nil
}

func (s *Service) Adjustment(ctx context.Context, userID string, input AdjustmentInput) error {
	if userID == "" {
		return apperror.ErrUnauthorized
	}
	if input.LotInternalID == "" || input.Type == "" || input.Qty == "" || input.DocRef == "" || input.ReasonCode == "" {
		return apperror.WithMessage(apperror.ErrValidation, "lot_internal_id, type, qty, doc_ref, reason_code are required")
	}
	typ := strings.TrimSpace(input.Type)
	if typ != "ADJ_IN" && typ != "ADJ_OUT" {
		return apperror.WithMessage(apperror.ErrValidation, "type must be ADJ_IN or ADJ_OUT")
	}

	qty := strings.TrimSpace(input.Qty)
	if qty == "" {
		return apperror.WithMessage(apperror.ErrValidation, "qty is required")
	}

	qtyChanged := qty
	if typ == "ADJ_OUT" {
		var err error
		qtyChanged, err = negateNumericString(qty)
		if err != nil {
			return apperror.WithMessage(apperror.ErrValidation, "invalid qty")
		}
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if strings.HasPrefix(qtyChanged, "-") {
		if err := s.applyNegativeStockGuard(ctx, tx, input.LotInternalID, qtyChanged); err != nil {
			return err
		}
	}

	reason := strings.TrimSpace(input.ReasonCode)
	if err := s.repo.InsertLedgerEntry(ctx, tx, userID, typ, input.LotInternalID, qtyChanged, input.DocRef, &reason); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if s.alerts != nil {
		s.alerts.AfterLotLedgerWrite(ctx, input.LotInternalID)
	}
	return nil
}

func (s *Service) applyNegativeStockGuard(ctx context.Context, tx pgx.Tx, lotInternalID string, qtyChanged string) error {
	if err := s.repo.LockLotForUpdate(ctx, tx, lotInternalID); err != nil {
		return err
	}
	physicalStr, err := s.repo.GetLotPhysicalQty(ctx, tx, lotInternalID)
	if err != nil {
		return err
	}

	physical, ok := new(big.Rat).SetString(physicalStr)
	if !ok {
		return apperror.WithMessage(apperror.ErrInternal, "cannot parse physical qty")
	}
	delta, ok := new(big.Rat).SetString(qtyChanged)
	if !ok {
		return apperror.WithMessage(apperror.ErrValidation, "invalid qty")
	}
	next := new(big.Rat).Add(physical, delta)
	if next.Sign() < 0 {
		return apperror.ErrNegativeStockPrevented
	}
	return nil
}

func negateNumericString(v string) (string, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(v))
	if !ok {
		return "", apperror.ErrValidation
	}
	if r.Sign() <= 0 {
		// require positive qty input for "out" adjustment; service will negate
		return "", apperror.ErrValidation
	}
	r.Neg(r)
	return r.FloatString(6), nil
}

