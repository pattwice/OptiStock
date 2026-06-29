package workorder

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"optistock/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateWOInput) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}
	input.WONumber = strings.TrimSpace(input.WONumber)
	input.TargetFGCode = strings.TrimSpace(input.TargetFGCode)
	input.TargetQty = strings.TrimSpace(input.TargetQty)
	if input.WONumber == "" || input.TargetFGCode == "" || input.TargetQty == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "wo_number, target_fg_code, target_qty are required")
	}
	if !isPositive(input.TargetQty) {
		return nil, apperror.WithMessage(apperror.ErrValidation, "target_qty must be positive")
	}

	itemType, err := s.repo.GetItemType(ctx, input.TargetFGCode)
	if err != nil {
		return nil, err
	}
	if itemType == "" {
		return nil, apperror.WithMessage(apperror.ErrNotFound, "target item not found")
	}
	if itemType != "FG" && itemType != "SFG" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "target_fg_code must be FG or SFG")
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.CreateHeader(ctx, tx, input)
	if err != nil {
		return nil, err
	}
	if err := runBOMExplosionTx(ctx, tx, s.repo, wo.WONumber, wo.TargetFGCode, wo.TargetQty); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, wo.WONumber)
}

func (s *Service) List(ctx context.Context, status string) ([]WorkOrder, error) {
	return s.repo.ListWorkOrders(ctx, status)
}

func (s *Service) Get(ctx context.Context, woNumber string) (*WorkOrderDetail, error) {
	wo, err := s.repo.GetHeader(ctx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	reqs, err := s.repo.ListRequirements(ctx, woNumber)
	if err != nil {
		return nil, err
	}
	allocs, err := s.repo.ListAllocations(ctx, woNumber)
	if err != nil {
		return nil, err
	}
	return &WorkOrderDetail{WorkOrder: *wo, Requirements: reqs, Allocations: allocs}, nil
}

func (s *Service) AllocationProposal(ctx context.Context, woNumber string) ([]AllocationProposal, error) {
	wo, err := s.repo.GetHeader(ctx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusDraft {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "allocation proposal only available in DRAFT")
	}

	reqs, err := s.repo.ListRequirements(ctx, woNumber)
	if err != nil {
		return nil, err
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	proposals, err := proposeAllAllocations(ctx, tx, s.repo, reqs)
	if err != nil {
		return nil, err
	}
	_ = tx.Rollback(ctx)
	return proposals, nil
}

func (s *Service) Reserve(ctx context.Context, userID, woNumber string, input ReserveInput) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusDraft {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "only DRAFT work orders can be reserved")
	}

	reqs, err := s.repo.ListRequirements(ctx, woNumber)
	if err != nil {
		return nil, err
	}

	allocations := input.Allocations
	if len(allocations) == 0 {
		proposals, err := proposeAllAllocations(ctx, tx, s.repo, reqs)
		if err != nil {
			return nil, err
		}
		for _, p := range proposals {
			if p.Shortage {
				return nil, apperror.WithDetails(apperror.ErrInsufficientStock, proposals)
			}
			allocations = append(allocations, p.Proposed...)
		}
	}

	if err := applyAllocations(ctx, tx, s.repo, allocations); err != nil {
		return nil, err
	}

	shortages, err := validateRequirementCoverage(ctx, tx, s.repo, reqs)
	if err != nil {
		return nil, err
	}
	if len(shortages) > 0 {
		return nil, apperror.WithDetails(apperror.ErrInsufficientStock, shortages)
	}

	oldStatus := wo.WOStatus
	if err := s.repo.UpdateStatus(ctx, tx, woNumber, StatusReserved); err != nil {
		return nil, err
	}
	action := "Status_Change"
	if input.ManualOverride {
		action = "Manual_Override"
	}
	if err := s.repo.InsertAudit(ctx, tx, woNumber, userID, action, string(oldStatus), string(StatusReserved)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) StartProduction(ctx context.Context, userID, woNumber string) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusReserved {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "only RESERVED work orders can start production")
	}

	oldStatus := wo.WOStatus
	if err := s.repo.UpdateStatus(ctx, tx, woNumber, StatusInProduction); err != nil {
		return nil, err
	}
	if err := s.repo.InsertAudit(ctx, tx, woNumber, userID, "Status_Change", string(oldStatus), string(StatusInProduction)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) UpdateActuals(ctx context.Context, userID, woNumber string, input UpdateActualsInput) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusInProduction {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "actuals can only be updated in IN_PRODUCTION")
	}

	for _, line := range input.Lines {
		damage := line.DamageQty
		if damage != nil && isZeroOrEmpty(*damage) {
			damage = nil
		}
		if err := s.repo.UpdateActuals(ctx, tx, line.AllocationID, line.ActualUsedQty, damage); err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(input.ActualProducedQty) != "" {
		pct, ok := mulRat(input.ActualProducedQty, "100")
		if !ok {
			return nil, apperror.WithMessage(apperror.ErrValidation, "invalid actual_produced_qty")
		}
		pct, ok = divRat(pct, wo.TargetQty)
		if !ok {
			return nil, apperror.ErrInternal
		}
		if err := s.repo.UpdateHeaderActuals(ctx, tx, woNumber, input.ActualProducedQty, pct); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) ResolveOverUsage(ctx context.Context, userID, woNumber string, input ResolveOverUsageInput) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}
	if input.ReqID == "" || len(input.Allocations) == 0 {
		return nil, apperror.WithMessage(apperror.ErrValidation, "req_id and allocations are required")
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusInProduction {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "over-usage resolution only in IN_PRODUCTION")
	}

	if err := applyAllocations(ctx, tx, s.repo, input.Allocations); err != nil {
		return nil, err
	}

	action := "Qty_Edit"
	if input.ManualOverride {
		action = "Manual_Override"
	}
	if err := s.repo.InsertAudit(ctx, tx, woNumber, userID, action, input.ReqID, "delta_allocation"); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) Complete(ctx context.Context, userID, woNumber string, input CompleteInput) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}
	actualProduced := strings.TrimSpace(input.ActualProducedQty)
	if actualProduced == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "actual_produced_qty is required")
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusInProduction {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "only IN_PRODUCTION work orders can be completed")
	}

	if cmp, ok := cmpRat(actualProduced, wo.TargetQty); ok && cmp < 0 {
		return nil, apperror.WithDetails(apperror.ErrApprovalRequired, map[string]string{
			"message": "partial completion requires supervisor approval (Phase 3)",
		})
	}
	if cmp, ok := cmpRat(actualProduced, wo.TargetQty); ok && cmp > 0 {
		return nil, apperror.WithMessage(apperror.ErrValidation, "actual_produced_qty cannot exceed target_qty")
	}

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

	docRef := fmt.Sprintf("WO:%s", woNumber)
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
				return nil, apperror.ErrValidation
			}
			if err := s.applyNegativeStockGuard(ctx, tx, a.LOTInternalID, qty); err != nil {
				return nil, err
			}
			if err := s.repo.InsertLedger(ctx, tx, userID, "WO_ISSUE", a.LOTInternalID, qty, docRef, nil); err != nil {
				return nil, err
			}
		}

		if isPositive(dmg) {
			qty, ok := negateQty(dmg)
			if !ok {
				return nil, apperror.ErrValidation
			}
			if err := s.applyNegativeStockGuard(ctx, tx, a.LOTInternalID, qty); err != nil {
				return nil, err
			}
			reason := damageReason
			if err := s.repo.InsertLedger(ctx, tx, userID, "ADJ_OUT", a.LOTInternalID, qty, docRef, &reason); err != nil {
				return nil, err
			}
		}

		remainder, ok := subRat(a.ReservedQty, used)
		if !ok {
			return nil, apperror.ErrInternal
		}
		remainder, ok = subRat(remainder, dmg)
		if !ok {
			return nil, apperror.ErrInternal
		}
		if isPositive(remainder) {
			if err := s.repo.InsertLedger(ctx, tx, userID, "RETURN_TO_STOCK", a.LOTInternalID, remainder, docRef, nil); err != nil {
				return nil, err
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

	fgLotID, err := s.repo.CreateFGLot(ctx, tx, wo.TargetFGCode, woNumber, mfgDate, expDate)
	if err != nil {
		return nil, err
	}
	if err := s.repo.InsertLedger(ctx, tx, userID, "WO_RECEIPT", fgLotID, actualProduced, docRef, nil); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateHeaderActuals(ctx, tx, woNumber, actualProduced, "100.00"); err != nil {
		return nil, err
	}
	if err := s.repo.ZeroReservations(ctx, tx, woNumber); err != nil {
		return nil, err
	}

	oldStatus := wo.WOStatus
	if err := s.repo.UpdateStatus(ctx, tx, woNumber, StatusCompleted); err != nil {
		return nil, err
	}
	if err := s.repo.InsertAudit(ctx, tx, woNumber, userID, "Status_Change", string(oldStatus), string(StatusCompleted)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) Cancel(ctx context.Context, userID, woNumber string) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	switch wo.WOStatus {
	case StatusDraft, StatusReserved, StatusInProduction:
	default:
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "work order cannot be cancelled from current status")
	}

	if err := s.repo.ZeroReservations(ctx, tx, woNumber); err != nil {
		return nil, err
	}
	oldStatus := wo.WOStatus
	if err := s.repo.UpdateStatus(ctx, tx, woNumber, StatusCancelled); err != nil {
		return nil, err
	}
	if err := s.repo.InsertAudit(ctx, tx, woNumber, userID, "Status_Change", string(oldStatus), string(StatusCancelled)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) Reopen(ctx context.Context, userID, woNumber string) (*WorkOrderDetail, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	wo, err := s.repo.GetHeaderForUpdate(ctx, tx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != StatusCancelled {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "only CANCELLED work orders can be reopened")
	}

	if err := s.repo.ClearAllocations(ctx, tx, woNumber); err != nil {
		return nil, err
	}
	oldStatus := wo.WOStatus
	if err := s.repo.UpdateStatus(ctx, tx, woNumber, StatusDraft); err != nil {
		return nil, err
	}
	if err := s.repo.InsertAudit(ctx, tx, woNumber, userID, "Reopen", string(oldStatus), string(StatusDraft)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx, woNumber)
}

func (s *Service) AuditLog(ctx context.Context, woNumber string) ([]AuditEntry, error) {
	wo, err := s.repo.GetHeader(ctx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	return s.repo.ListAudit(ctx, woNumber)
}

func (s *Service) applyNegativeStockGuard(ctx context.Context, tx pgx.Tx, lotID, qtyChanged string) error {
	if err := s.repo.LockLot(ctx, tx, lotID); err != nil {
		return err
	}
	physical, err := s.repo.GetLotPhysicalQty(ctx, tx, lotID)
	if err != nil {
		return err
	}
	next, ok := addRat(physical, qtyChanged)
	if !ok {
		return apperror.ErrInternal
	}
	if cmp, ok := cmpRat(next, "0"); ok && cmp < 0 {
		return apperror.ErrNegativeStockPrevented
	}
	return nil
}

func divRat(a, b string) (string, bool) {
	ra, ok := parseRat(a)
	if !ok {
		return "", false
	}
	rb, ok := parseRat(b)
	if !ok || rb.Sign() == 0 {
		return "", false
	}
	return ratString(new(big.Rat).Quo(ra, rb)), true
}
