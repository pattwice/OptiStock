package item

import (
	"context"

	"optistock/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateItem(ctx context.Context, input CreateItemInput) (*Item, error) {
	if input.ItemCode == "" || input.Description == "" || input.Unit == "" || input.ItemType == "" || input.MinStockLevel == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "item_code, description, unit, item_type, and min_stock_level are required")
	}
	return s.repo.CreateItem(ctx, input)
}

func (s *Service) UpdateItem(ctx context.Context, itemCode string, input UpdateItemInput) (*Item, error) {
	if itemCode == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "item_code is required")
	}
	it, err := s.repo.UpdateItem(ctx, itemCode, input)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, apperror.ErrNotFound
	}
	return it, nil
}

func (s *Service) GetItem(ctx context.Context, itemCode string) (*Item, error) {
	if itemCode == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "item_code is required")
	}
	it, err := s.repo.GetItem(ctx, itemCode)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, apperror.ErrNotFound
	}
	return it, nil
}

func (s *Service) ListItems(ctx context.Context, itemType string) ([]Item, error) {
	return s.repo.ListItems(ctx, itemType)
}

func (s *Service) CreateBOMRow(ctx context.Context, input CreateBOMRowInput) (*BOMRow, error) {
	if input.ParentItemCode == "" || input.ComponentItemCode == "" || input.QtyPerSet == "" || input.BomVersion == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "parent_item_code, component_item_code, qty_per_set, bom_version are required")
	}
	return s.repo.CreateBOMRow(ctx, input)
}

func (s *Service) ListBOMByParent(ctx context.Context, parentItemCode string) ([]BOMRow, error) {
	if parentItemCode == "" {
		return nil, apperror.WithMessage(apperror.ErrValidation, "parent_item_code is required")
	}
	return s.repo.ListBOMByParent(ctx, parentItemCode)
}

func (s *Service) ActivateBOMVersion(ctx context.Context, parentItemCode, bomVersion string) error {
	if parentItemCode == "" || bomVersion == "" {
		return apperror.WithMessage(apperror.ErrValidation, "parent_item_code and bom_version are required")
	}
	affected, err := s.repo.ActivateBOMVersion(ctx, parentItemCode, bomVersion)
	if err != nil {
		return err
	}
	if affected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

