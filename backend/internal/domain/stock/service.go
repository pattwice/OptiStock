package stock

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListOnHand(ctx context.Context, itemCode, status string) ([]OnHandRow, error) {
	return s.repo.ListOnHand(ctx, itemCode, status)
}
