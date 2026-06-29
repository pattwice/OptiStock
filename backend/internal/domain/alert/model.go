package alert

const (
	TypeLowStock               = "LOW_STOCK"
	TypeNearExpiry             = "NEAR_EXPIRY"
	TypeLotOnHold              = "LOT_ON_HOLD"
	TypeLotQuarantined         = "LOT_QUARANTINED"
	TypeOverReservationBlocked = "OVER_RESERVATION_BLOCKED"
	TypeApprovalPending        = "APPROVAL_PENDING"
	TypeApprovalWithdrawn      = "APPROVAL_WITHDRAWN"
)

const (
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)
