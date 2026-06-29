package workorder

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"optistock/pkg/apperror"
)

func (s *Service) persistCompletionActuals(
	ctx context.Context,
	tx pgx.Tx,
	woNumber string,
	input CompleteInput,
) ([]Allocation, error) {
	for _, line := range input.Lines {
		damage := line.DamageQty
		if damage != nil && isZeroOrEmpty(*damage) {
			zero := "0"
			damage = &zero
		}
		if damage == nil {
			zero := "0"
			damage = &zero
		}
		if err := s.repo.UpdateActuals(ctx, tx, line.AllocationID, line.ActualUsedQty, damage); err != nil {
			return nil, err
		}
	}

	allocs, err := s.repo.ListAllocationsTx(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	for _, a := range allocs {
		used := "0"
		if a.ActualUsedQty != nil {
			used = *a.ActualUsedQty
		}
		dmg := "0"
		if a.DamageQty != nil {
			dmg = *a.DamageQty
		}
		sum, ok := addRat(used, dmg)
		if !ok {
			return nil, apperror.ErrInternal
		}
		if cmp, ok := cmpRat(sum, a.ReservedQty); ok && cmp > 0 {
			return nil, apperror.ErrCompletionGuardFailed
		}
	}
	return allocs, nil
}

func (s *Service) executeCompletionLedger(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	wo *WorkOrder,
	actualProduced string,
	allocs []Allocation,
) (string, error) {
	docRef := fmt.Sprintf("WO:%s", wo.WONumber)
	const damageReason = "PRODUCTION_DAMAGE"

	for _, a := range allocs {
		used := "0"
		if a.ActualUsedQty != nil {
			used = *a.ActualUsedQty
		}
		dmg := "0"
		if a.DamageQty != nil {
			dmg = *a.DamageQty
		}

		if isPositive(used) {
			qty, ok := negateQty(used)
			if !ok {
				return "", apperror.ErrValidation
			}
			if err := s.applyNegativeStockGuard(ctx, tx, a.LOTInternalID, qty); err != nil {
				return "", err
			}
			if err := s.repo.InsertLedger(ctx, tx, userID, "WO_ISSUE", a.LOTInternalID, qty, docRef, nil); err != nil {
				return "", err
			}
		}

		if isPositive(dmg) {
			qty, ok := negateQty(dmg)
			if !ok {
				return "", apperror.ErrValidation
			}
			if err := s.applyNegativeStockGuard(ctx, tx, a.LOTInternalID, qty); err != nil {
				return "", err
			}
			reason := damageReason
			if err := s.repo.InsertLedger(ctx, tx, userID, "ADJ_OUT", a.LOTInternalID, qty, docRef, &reason); err != nil {
				return "", err
			}
		}

		remainder, ok := subRat(a.ReservedQty, used)
		if !ok {
			return "", apperror.ErrInternal
		}
		remainder, ok = subRat(remainder, dmg)
		if !ok {
			return "", apperror.ErrInternal
		}
		if isPositive(remainder) {
			if err := s.repo.InsertLedger(ctx, tx, userID, "RETURN_TO_STOCK", a.LOTInternalID, remainder, docRef, nil); err != nil {
				return "", err
			}
		}
	}

	mfgDate := time.Now().UTC().Format("2006-01-02")
	var expDate *string
	if shelf, err := s.repo.GetItemShelfLife(ctx, wo.TargetFGCode); err == nil && shelf != nil {
		t, _ := time.Parse("2006-01-02", mfgDate)
		exp := t.AddDate(0, 0, *shelf).Format("2006-01-02")
		expDate = &exp
	}

	fgLotID, err := s.repo.CreateFGLot(ctx, tx, wo.TargetFGCode, wo.WONumber, mfgDate, expDate)
	if err != nil {
		return "", err
	}
	if err := s.repo.InsertLedger(ctx, tx, userID, "WO_RECEIPT", fgLotID, actualProduced, docRef, nil); err != nil {
		return "", err
	}
	return fgLotID, nil
}

func (s *Service) finalizeCompletion(
	ctx context.Context,
	tx pgx.Tx,
	userID, woNumber string,
	wo *WorkOrder,
	actualProduced, completionPct string,
	finalStatus WOStatus,
	auditAction string,
) error {
	if err := s.repo.UpdateHeaderActuals(ctx, tx, woNumber, actualProduced, completionPct); err != nil {
		return err
	}
	if err := s.repo.ZeroReservations(ctx, tx, woNumber); err != nil {
		return err
	}
	oldStatus := wo.WOStatus
	if err := s.repo.UpdateStatus(ctx, tx, woNumber, finalStatus); err != nil {
		return err
	}
	return s.repo.InsertAudit(ctx, tx, woNumber, userID, auditAction, string(oldStatus), string(finalStatus))
}

// FinalizeApprovedCompletionTx runs ledger writes for a supervisor-approved partial close inside an open transaction.
func (s *Service) FinalizeApprovedCompletionTx(ctx context.Context, tx pgx.Tx, supervisorID, woNumber string) error {
	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return err
	}
	if wo == nil {
		return apperror.ErrNotFound
	}
	if wo.WOStatus != StatusPendingApproval {
		return apperror.WithMessage(apperror.ErrWOStatusInvalid, "work order is not pending approval")
	}
	if wo.ActualProducedQty == nil || wo.CompletionPct == nil {
		return apperror.WithMessage(apperror.ErrValidation, "actual produced quantity missing on work order")
	}

	allocs, err := s.repo.ListAllocationsTx(ctx, tx, woNumber)
	if err != nil {
		return err
	}
	if _, err := s.executeCompletionLedger(ctx, tx, supervisorID, wo, *wo.ActualProducedQty, allocs); err != nil {
		return err
	}
	return s.finalizeCompletion(ctx, tx, supervisorID, woNumber, wo, *wo.ActualProducedQty, *wo.CompletionPct, StatusCompletedPartial, "Approval_Resolved")
}

func (s *Service) EmitPostCompletionAlerts(ctx context.Context, woNumber string) {
	allocs, err := s.repo.ListAllocations(ctx, woNumber)
	if err != nil {
		return
	}
	fgLotID, _ := s.repo.FindFGLotIDByWONumber(ctx, woNumber)
	s.emitPostCompletionAlerts(ctx, woNumber, fgLotID, allocs)
}

func (s *Service) emitPostCompletionAlerts(ctx context.Context, woNumber, fgLotID string, allocs []Allocation) {
	if s.alerts == nil {
		return
	}
	for _, a := range allocs {
		s.alerts.AfterLotLedgerWrite(ctx, a.LOTInternalID)
	}
	if fgLotID != "" {
		s.alerts.AfterLotLedgerWrite(ctx, fgLotID)
	}
	if wo, err := s.repo.GetHeader(ctx, woNumber); err == nil && wo != nil {
		s.alerts.AfterLedgerWrite(ctx, wo.TargetFGCode)
	}
}
