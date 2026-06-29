package lot

import (
	"context"
	"strings"

	"optistock/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateLot(ctx context.Context, input CreateLotInput) (*Lot, error) {
	if input.ItemCode == "" || input.SupplierLotNumber == "" || input.MfgDate == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "item_code, supplier_lot_number, mfg_date are required")
	}
	if input.Status == "" {
		input.Status = "Active"
	}
	if !isValidStatus(input.Status) {
		return nil, apperror.WithMessage(apperror.ErrValidation, "invalid status")
	}
	return s.repo.CreateLot(ctx, input)
}

func (s *Service) GetLot(ctx context.Context, lotInternalID string) (*Lot, error) {
	if lotInternalID == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "lot_internal_id is required")
	}
	l, err := s.repo.GetLot(ctx, lotInternalID)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, apperror.ErrNotFound
	}
	return l, nil
}

func (s *Service) ListLots(ctx context.Context, itemCode, status string) ([]Lot, error) {
	if status != "" && !isValidStatus(status) {
		return nil, apperror.WithMessage(apperror.ErrValidation, "invalid status")
	}
	return s.repo.ListLots(ctx, itemCode, status)
}

func (s *Service) UpdateLotStatus(ctx context.Context, lotInternalID, newStatus string) (*Lot, error) {
	if lotInternalID == "" || newStatus == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "lot_internal_id and status are required")
	}
	if !isValidStatus(newStatus) {
		return nil, apperror.WithMessage(apperror.ErrValidation, "invalid status")
	}

	current, err := s.repo.GetLot(ctx, lotInternalID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, apperror.ErrNotFound
	}
	if !isValidTransition(current.Status, newStatus) {
		return nil, apperror.WithMessage(apperror.ErrValidation, "invalid status transition")
	}

	updated, err := s.repo.UpdateLotStatus(ctx, lotInternalID, newStatus)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperror.ErrNotFound
	}
	return updated, nil
}

func isValidStatus(s string) bool {
	switch strings.TrimSpace(s) {
	case "Active", "Hold", "Quarantined":
		return true
	default:
		return false
	}
}

func isValidTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "Active":
		return to == "Hold" || to == "Quarantined"
	case "Hold":
		return to == "Active" || to == "Quarantined"
	case "Quarantined":
		return to == "Hold"
	default:
		return false
	}
}

