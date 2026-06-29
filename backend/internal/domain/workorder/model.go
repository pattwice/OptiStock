package workorder

import "time"

type WOStatus string

const (
	StatusDraft         WOStatus = "DRAFT"
	StatusReserved      WOStatus = "RESERVED"
	StatusInProduction  WOStatus = "IN_PRODUCTION"
	StatusPendingApproval WOStatus = "PENDING_APPROVAL"
	StatusCompleted     WOStatus = "COMPLETED"
	StatusCompletedPartial WOStatus = "COMPLETED_PARTIAL"
	StatusCancelled     WOStatus = "CANCELLED"
)

type WorkOrder struct {
	WONumber          string    `json:"wo_number"`
	TargetFGCode      string    `json:"target_fg_code"`
	TargetQty         string    `json:"target_qty"`
	ActualProducedQty *string   `json:"actual_produced_qty"`
	CompletionPct     *string   `json:"completion_pct"`
	WOStatus          WOStatus  `json:"wo_status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Requirement struct {
	ReqID            string `json:"req_id"`
	WONumber         string `json:"wo_number"`
	RequiredItemCode string `json:"required_item_code"`
	TotalNeededQty   string `json:"total_needed_qty"`
	AvailableQty     string `json:"available_qty,omitempty"`
	Shortage         bool   `json:"shortage,omitempty"`
}

type Allocation struct {
	AllocationID   string  `json:"allocation_id"`
	ReqID          string  `json:"req_id"`
	LOTInternalID  string  `json:"lot_internal_id"`
	ReservedQty    string  `json:"reserved_qty"`
	ActualUsedQty  *string `json:"actual_used_qty"`
	DamageQty      *string `json:"damage_qty"`
	ItemCode       string  `json:"item_code,omitempty"`
	SupplierLot    string  `json:"supplier_lot_number,omitempty"`
	ExpDate        *string `json:"exp_date,omitempty"`
	MfgDate        *string `json:"mfg_date,omitempty"`
}

type AuditEntry struct {
	LogID     string    `json:"log_id"`
	WONumber  string    `json:"wo_number"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	OldValue  *string   `json:"old_value"`
	NewValue  *string   `json:"new_value"`
}

type WorkOrderDetail struct {
	WorkOrder    WorkOrder     `json:"work_order"`
	Requirements []Requirement `json:"requirements"`
	Allocations  []Allocation  `json:"allocations"`
}

type CreateWOInput struct {
	WONumber     string `json:"wo_number"`
	TargetFGCode string `json:"target_fg_code"`
	TargetQty    string `json:"target_qty"`
}

type ReserveAllocationInput struct {
	ReqID         string `json:"req_id"`
	LOTInternalID string `json:"lot_internal_id"`
	ReservedQty   string `json:"reserved_qty"`
}

type ReserveInput struct {
	Allocations    []ReserveAllocationInput `json:"allocations"`
	ManualOverride bool                     `json:"manual_override"`
}

type AllocationProposal struct {
	ReqID            string                   `json:"req_id"`
	RequiredItemCode string                   `json:"required_item_code"`
	TotalNeededQty   string                   `json:"total_needed_qty"`
	Proposed         []ReserveAllocationInput `json:"proposed"`
	Shortage         bool                     `json:"shortage"`
	ShortageQty      string                   `json:"shortage_qty,omitempty"`
}

type UpdateActualsInput struct {
	ActualProducedQty string              `json:"actual_produced_qty"`
	Lines             []ActualLineInput   `json:"lines"`
}

type ActualLineInput struct {
	AllocationID  string  `json:"allocation_id"`
	ActualUsedQty string  `json:"actual_used_qty"`
	DamageQty     *string `json:"damage_qty"`
}

type ResolveOverUsageInput struct {
	ReqID          string                   `json:"req_id"`
	Allocations    []ReserveAllocationInput `json:"allocations"`
	ManualOverride bool                     `json:"manual_override"`
}

type CompleteInput struct {
	ActualProducedQty string            `json:"actual_produced_qty"`
	Lines             []ActualLineInput `json:"lines"`
}
