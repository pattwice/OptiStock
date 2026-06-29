package ledger

import (
	"context"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, itemCode, lotInternalID, fromISO, toISO string, limit int) ([]LedgerEntry, error) {
	var from *time.Time
	var to *time.Time

	if fromISO != "" {
		t, err := time.Parse(time.RFC3339, fromISO)
		if err != nil {
			return nil, err
		}
		from = &t
	}
	if toISO != "" {
		t, err := time.Parse(time.RFC3339, toISO)
		if err != nil {
			return nil, err
		}
		to = &t
	}

	return s.repo.List(ctx, ListParams{
		ItemCode:      itemCode,
		LotInternalID: lotInternalID,
		From:          from,
		To:            to,
		Limit:         limit,
	})
}

