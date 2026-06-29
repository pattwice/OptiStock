package ledger

import "time"

type LedgerEntry struct {
	TransactionID   string    `json:"transaction_id"`
	Timestamp       time.Time `json:"timestamp"`
	UserID          string    `json:"user_id"`
	TransactionType string    `json:"transaction_type"`
	LotInternalID   string    `json:"lot_internal_id"`
	ItemCode        string    `json:"item_code"`
	QtyChanged      string    `json:"qty_changed"`
	DocRef          string    `json:"doc_ref"`
	ReasonCode      *string   `json:"reason_code,omitempty"`
}

