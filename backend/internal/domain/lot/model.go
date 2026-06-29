package lot

import "time"

type Lot struct {
	LotInternalID     string     `json:"lot_internal_id"`
	ItemCode          string     `json:"item_code"`
	SupplierLotNumber string     `json:"supplier_lot_number"`
	SupplierName      *string    `json:"supplier_name,omitempty"`
	MfgDate           string     `json:"mfg_date"` // YYYY-MM-DD
	ExpDate           *string    `json:"exp_date,omitempty"`
	Status            string     `json:"status"` // Active | Hold | Quarantined
	CreatedAt         time.Time  `json:"created_at"`
}

type CreateLotInput struct {
	ItemCode          string  `json:"item_code"`
	SupplierLotNumber string  `json:"supplier_lot_number"`
	SupplierName      *string `json:"supplier_name"`
	MfgDate           string  `json:"mfg_date"`
	ExpDate           *string `json:"exp_date"`
	Status            string  `json:"status"` // default Active
}

type UpdateLotStatusInput struct {
	Status string `json:"status"`
}

