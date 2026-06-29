package receiving

type POReceiptInput struct {
	ItemCode          string  `json:"item_code"`
	SupplierLotNumber string  `json:"supplier_lot_number"`
	SupplierName      *string `json:"supplier_name"`
	MfgDate           string  `json:"mfg_date"` // YYYY-MM-DD
	ExpDate           *string `json:"exp_date"` // YYYY-MM-DD or null
	QtyReceived       string  `json:"qty_received"`
	DocRef            string  `json:"doc_ref"` // e.g., PO-123

	DamageQty        *string `json:"damage_qty"`         // optional; treated as ADJ_OUT
	DamageReasonCode *string `json:"damage_reason_code"` // optional; default SUPPLIER_DAMAGE when DamageQty is set
}

type AdjustmentInput struct {
	LotInternalID string `json:"lot_internal_id"`
	Type          string `json:"type"` // ADJ_IN | ADJ_OUT
	Qty           string `json:"qty"`  // positive numeric string
	DocRef        string `json:"doc_ref"`
	ReasonCode    string `json:"reason_code"`
}

type POReceiptResult struct {
	LotInternalID string `json:"lot_internal_id"`
	ItemCode      string `json:"item_code"`
	Posted        int    `json:"posted_entries"`
}

