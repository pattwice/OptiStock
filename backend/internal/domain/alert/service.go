package alert

import (
	"context"
	"time"

	"optistock/internal/auth"
	"optistock/internal/ws"
)

type Service struct {
	repo *Repository
	hub  *ws.Hub
}

func NewService(repo *Repository, hub *ws.Hub) *Service {
	return &Service{repo: repo, hub: hub}
}

func (s *Service) StartNearExpiryScheduler(ctx context.Context) {
	go func() {
		s.scanNearExpiry(ctx)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.scanNearExpiry(ctx)
			}
		}
	}()
}

func (s *Service) scanNearExpiry(ctx context.Context) {
	enabled, err := s.repo.ConfigEnabled(ctx, "NEAR_EXPIRY_ALERT_ENABLED")
	if err != nil || !enabled {
		return
	}
	rows, err := s.repo.ListNearExpiryLots(ctx)
	if err != nil {
		return
	}
	for _, row := range rows {
		s.publish(ws.AlertMessage{
			Type:     TypeNearExpiry,
			Severity: SeverityWarning,
			Payload: map[string]any{
				"lot_internal_id":     row.LotInternalID,
				"item_code":           row.ItemCode,
				"description":         row.Description,
				"supplier_lot_number": row.SupplierLotNumber,
				"exp_date":            row.ExpDate,
				"days_until_expiry":   row.DaysUntilExpiry,
			},
		}, nil, nil)
	}
}

func (s *Service) AfterLedgerWrite(ctx context.Context, itemCode string) {
	if itemCode == "" {
		return
	}
	enabled, err := s.repo.ConfigEnabled(ctx, "LOW_STOCK_ALERT_ENABLED")
	if err != nil || !enabled {
		return
	}
	row, low, err := s.repo.CheckLowStock(ctx, itemCode)
	if err != nil || !low || row == nil {
		return
	}
	s.publish(ws.AlertMessage{
		Type:     TypeLowStock,
		Severity: SeverityWarning,
		Payload: map[string]any{
			"item_code":       row.ItemCode,
			"description":     row.Description,
			"available_qty":   row.AvailableQty,
			"min_stock_level": row.MinStockLevel,
		},
	}, nil, nil)
}

func (s *Service) AfterLotChange(ctx context.Context, lotID string) {
	enabled, err := s.repo.ConfigEnabled(ctx, "NEAR_EXPIRY_ALERT_ENABLED")
	if err != nil || !enabled {
		return
	}
	row, near, err := s.repo.CheckNearExpiryLot(ctx, lotID)
	if err != nil || !near || row == nil {
		return
	}
	s.publish(ws.AlertMessage{
		Type:     TypeNearExpiry,
		Severity: SeverityWarning,
		Payload: map[string]any{
			"lot_internal_id":     row.LotInternalID,
			"item_code":           row.ItemCode,
			"description":         row.Description,
			"supplier_lot_number": row.SupplierLotNumber,
			"exp_date":            row.ExpDate,
			"days_until_expiry":   row.DaysUntilExpiry,
		},
	}, nil, nil)
}

func (s *Service) OnLotStatusChange(ctx context.Context, lotID, newStatus string) {
	if newStatus != "Hold" && newStatus != "Quarantined" {
		return
	}
	owners, _ := s.repo.AffectedWOOwners(ctx, lotID)
	woNumbers, _ := s.repo.AffectedWONumbers(ctx, lotID)
	itemCode, _ := s.repo.ItemCodeForLot(ctx, lotID)

	alertType := TypeLotOnHold
	if newStatus == "Quarantined" {
		alertType = TypeLotQuarantined
	}
	s.publish(ws.AlertMessage{
		Type:     alertType,
		Severity: SeverityWarning,
		Payload: map[string]any{
			"lot_internal_id": lotID,
			"item_code":       itemCode,
			"status":          newStatus,
			"wo_numbers":      woNumbers,
		},
	}, nil, owners)
}

func (s *Service) OnReservationBlocked(ctx context.Context, woNumber string, shortages any) {
	s.publish(ws.AlertMessage{
		Type:     TypeOverReservationBlocked,
		Severity: SeverityWarning,
		Payload: map[string]any{
			"wo_number": woNumber,
			"shortages": shortages,
		},
	}, nil, nil)
}

func (s *Service) OnApprovalPending(ctx context.Context, woNumber, fgCode, completionPct, requesterName string) {
	s.publish(ws.AlertMessage{
		Type:     TypeApprovalPending,
		Severity: SeverityInfo,
		Payload: map[string]any{
			"wo_number":       woNumber,
			"target_fg_code":  fgCode,
			"completion_pct":  completionPct,
			"requester_name":  requesterName,
		},
	}, []string{string(auth.RoleSupervisor), string(auth.RoleAdmin)}, nil)
}

func (s *Service) OnApprovalWithdrawn(ctx context.Context, woNumber, fgCode, completionPct, requesterName string) {
	s.publish(ws.AlertMessage{
		Type:     TypeApprovalWithdrawn,
		Severity: SeverityInfo,
		Payload: map[string]any{
			"wo_number":      woNumber,
			"target_fg_code": fgCode,
			"completion_pct": completionPct,
			"requester_name": requesterName,
		},
	}, []string{string(auth.RoleSupervisor), string(auth.RoleAdmin)}, nil)
}

func (s *Service) AfterLotLedgerWrite(ctx context.Context, lotID string) {
	itemCode, err := s.repo.ItemCodeForLot(ctx, lotID)
	if err != nil || itemCode == "" {
		return
	}
	s.AfterLedgerWrite(ctx, itemCode)
}

func (s *Service) publish(msg ws.AlertMessage, roles, userIDs []string) {
	if s.hub == nil {
		return
	}
	msg.Timestamp = time.Now().UTC()
	s.hub.Publish(msg, roles, userIDs)
}
