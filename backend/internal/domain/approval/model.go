package approval

import "time"

type ApprovalType string

const ApprovalTypePartialCompletion ApprovalType = "PARTIAL_COMPLETION"

type ApprovalStatus string

const (
	StatusPending  ApprovalStatus = "PENDING"
	StatusApproved ApprovalStatus = "APPROVED"
	StatusRejected ApprovalStatus = "REJECTED"
	StatusWithdrawn ApprovalStatus = "WITHDRAWN"
)

type Approval struct {
	ApprovalID              string         `json:"approval_id"`
	WONumber                string         `json:"wo_number"`
	ApprovalType            ApprovalType   `json:"approval_type"`
	RequestedBy             string         `json:"requested_by"`
	RequestedByName         string         `json:"requested_by_name,omitempty"`
	RequestedAt             time.Time      `json:"requested_at"`
	CompletionPctAtRequest  string         `json:"completion_pct_at_request"`
	ApprovalStatus          ApprovalStatus `json:"approval_status"`
	ResolvedBy              *string        `json:"resolved_by"`
	ResolvedByName          *string        `json:"resolved_by_name,omitempty"`
	ResolvedAt              *time.Time     `json:"resolved_at"`
	ResolutionNotes         *string        `json:"resolution_notes"`
	TargetFGCode            string         `json:"target_fg_code,omitempty"`
	TargetQty               string         `json:"target_qty,omitempty"`
	ActualProducedQty       *string        `json:"actual_produced_qty,omitempty"`
}

type ResolveInput struct {
	ResolutionNotes string `json:"resolution_notes"`
}
