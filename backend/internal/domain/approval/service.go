package approval

import (
	"context"
	"strings"

	"optistock/internal/domain/alert"
	"optistock/internal/domain/workorder"
	"optistock/pkg/apperror"
)

type Service struct {
	repo   *Repository
	woRepo *workorder.Repository
	woSvc  *workorder.Service
	alerts alert.Notifier
}

func NewService(repo *Repository, woRepo *workorder.Repository, woSvc *workorder.Service, alerts alert.Notifier) *Service {
	return &Service{repo: repo, woRepo: woRepo, woSvc: woSvc, alerts: alerts}
}

func (s *Service) ListPending(ctx context.Context) ([]Approval, error) {
	return s.repo.ListPending(ctx)
}

func (s *Service) ListByWO(ctx context.Context, woNumber string) ([]Approval, error) {
	wo, err := s.woRepo.GetHeader(ctx, woNumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	return s.repo.ListByWO(ctx, woNumber)
}

func (s *Service) Withdraw(ctx context.Context, userID, approvalID string) (*Approval, error) {
	if userID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.woRepo.Pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	approval, err := s.repo.GetByIDForUpdate(ctx, tx, approvalID)
	if err != nil {
		return nil, err
	}
	if approval == nil {
		return nil, apperror.ErrNotFound
	}
	if approval.ApprovalStatus != StatusPending {
		return nil, apperror.WithMessage(apperror.ErrValidation, "only pending approvals can be withdrawn")
	}
	if approval.RequestedBy != userID {
		return nil, apperror.ErrForbidden
	}

	wo, err := s.woRepo.GetHeaderForUpdate(ctx, tx, approval.WONumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != workorder.StatusPendingApproval {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "work order is not pending approval")
	}

	if err := s.repo.Resolve(ctx, tx, approvalID, StatusWithdrawn, &userID, nil); err != nil {
		return nil, err
	}
	oldStatus := wo.WOStatus
	if err := s.woRepo.UpdateStatus(ctx, tx, approval.WONumber, workorder.StatusInProduction); err != nil {
		return nil, err
	}
	if err := s.woRepo.InsertAudit(ctx, tx, approval.WONumber, userID, "Approval_Withdrawn", string(oldStatus), string(workorder.StatusInProduction)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	if s.alerts != nil {
		s.alerts.OnApprovalWithdrawn(ctx, approval.WONumber, approval.TargetFGCode, approval.CompletionPctAtRequest, approval.RequestedByName)
	}
	return s.repo.GetByID(ctx, approvalID)
}

func (s *Service) Approve(ctx context.Context, supervisorID, approvalID string, input ResolveInput) (*Approval, error) {
	if supervisorID == "" {
		return nil, apperror.ErrUnauthorized
	}

	tx, err := s.woRepo.Pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	approval, err := s.repo.GetByIDForUpdate(ctx, tx, approvalID)
	if err != nil {
		return nil, err
	}
	if approval == nil {
		return nil, apperror.ErrNotFound
	}
	if approval.ApprovalStatus != StatusPending {
		return nil, apperror.WithMessage(apperror.ErrValidation, "only pending approvals can be approved")
	}

	notes := strings.TrimSpace(input.ResolutionNotes)
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}
	if err := s.repo.Resolve(ctx, tx, approvalID, StatusApproved, &supervisorID, notesPtr); err != nil {
		return nil, err
	}
	if err := s.woSvc.FinalizeApprovedCompletionTx(ctx, tx, supervisorID, approval.WONumber); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	s.woSvc.EmitPostCompletionAlerts(ctx, approval.WONumber)
	return s.repo.GetByID(ctx, approvalID)
}

func (s *Service) Reject(ctx context.Context, supervisorID, approvalID string, input ResolveInput) (*Approval, error) {
	if supervisorID == "" {
		return nil, apperror.ErrUnauthorized
	}
	notes := strings.TrimSpace(input.ResolutionNotes)
	if notes == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "resolution_notes is required on rejection")
	}

	tx, err := s.woRepo.Pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	approval, err := s.repo.GetByIDForUpdate(ctx, tx, approvalID)
	if err != nil {
		return nil, err
	}
	if approval == nil {
		return nil, apperror.ErrNotFound
	}
	if approval.ApprovalStatus != StatusPending {
		return nil, apperror.WithMessage(apperror.ErrValidation, "only pending approvals can be rejected")
	}

	wo, err := s.woRepo.GetHeaderForUpdate(ctx, tx, approval.WONumber)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, apperror.ErrNotFound
	}
	if wo.WOStatus != workorder.StatusPendingApproval {
		return nil, apperror.WithMessage(apperror.ErrWOStatusInvalid, "work order is not pending approval")
	}

	if err := s.repo.Resolve(ctx, tx, approvalID, StatusRejected, &supervisorID, &notes); err != nil {
		return nil, err
	}
	oldStatus := wo.WOStatus
	if err := s.woRepo.UpdateStatus(ctx, tx, approval.WONumber, workorder.StatusInProduction); err != nil {
		return nil, err
	}
	if err := s.woRepo.InsertAudit(ctx, tx, approval.WONumber, supervisorID, "Approval_Resolved", string(oldStatus), string(workorder.StatusInProduction)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, approvalID)
}
