package stock

type OnHandRow struct {
	LotInternalID     string  `json:"lot_internal_id"`
	ItemCode          string  `json:"item_code"`
	SupplierLotNumber string  `json:"supplier_lot_number"`
	SupplierName      *string `json:"supplier_name,omitempty"`
	MfgDate           string  `json:"mfg_date"`
	ExpDate           *string `json:"exp_date,omitempty"`
	Status            string  `json:"status"`
	PhysicalQty       string  `json:"physical_qty"`
	ReservedQty       string  `json:"reserved_qty"`
	AvailableQty      string  `json:"available_qty"`
}
