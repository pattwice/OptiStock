package alert

import "context"

// Notifier is implemented by alert.Service for optional injection into domain services.
type Notifier interface {
	AfterLedgerWrite(ctx context.Context, itemCode string)
	AfterLotLedgerWrite(ctx context.Context, lotID string)
	AfterLotChange(ctx context.Context, lotID string)
	OnLotStatusChange(ctx context.Context, lotID, newStatus string)
	OnReservationBlocked(ctx context.Context, woNumber string, shortages any)
	OnApprovalPending(ctx context.Context, woNumber, fgCode, completionPct, requesterName string)
	OnApprovalWithdrawn(ctx context.Context, woNumber, fgCode, completionPct, requesterName string)
}
