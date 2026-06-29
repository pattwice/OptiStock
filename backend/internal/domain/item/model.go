package item

import "time"

type Item struct {
	ItemCode            string    `json:"item_code"`
	Description         string    `json:"description"`
	Unit                string    `json:"unit"`
	ItemType            string    `json:"item_type"` // FG | SFG | RM
	MinStockLevel       string    `json:"min_stock_level"`
	ExpiryThresholdDays *int      `json:"expiry_threshold_days,omitempty"`
	ShelfLifeDays       *int      `json:"shelf_life_days,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CreateItemInput struct {
	ItemCode            string `json:"item_code"`
	Description         string `json:"description"`
	Unit                string `json:"unit"`
	ItemType            string `json:"item_type"`
	MinStockLevel       string `json:"min_stock_level"`
	ExpiryThresholdDays *int   `json:"expiry_threshold_days"`
	ShelfLifeDays       *int   `json:"shelf_life_days"`
}

type UpdateItemInput struct {
	Description         *string `json:"description"`
	Unit                *string `json:"unit"`
	ItemType            *string `json:"item_type"`
	MinStockLevel       *string `json:"min_stock_level"`
	ExpiryThresholdDays **int   `json:"expiry_threshold_days"` // pointer-to-pointer allows explicit null
	ShelfLifeDays       **int   `json:"shelf_life_days"`
}

type BOMRow struct {
	BomID             string    `json:"bom_id"`
	ParentItemCode    string    `json:"parent_item_code"`
	ComponentItemCode string    `json:"component_item_code"`
	QtyPerSet         string    `json:"qty_per_set"`
	BomVersion        string    `json:"bom_version"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateBOMRowInput struct {
	ParentItemCode    string `json:"parent_item_code"`
	ComponentItemCode string `json:"component_item_code"`
	QtyPerSet         string `json:"qty_per_set"`
	BomVersion        string `json:"bom_version"`
	IsActive          bool   `json:"is_active"`
}

