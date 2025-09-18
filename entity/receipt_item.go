package entity

type ReceiptItem struct {
	ReceiptItemId     int64   `json:"receipt_item_id"`
	ReceiptId         int64   `json:"receipt_id"`
	ItemCategory      string  `json:"item_category"`
	ItemName          string  `json:"item_name"`
	ItemQuantity      *int    `json:"item_quantity"`
	ItemPriceCurrency string  `json:"item_price_currency"`
	ItemPriceNumeric  float64 `json:"item_price_numeric"`
	CreatedAt         int64   `json:"created_at"`
	UpdatedAt         *int64  `json:"updated_at,omitempty"`
	DeletedAt         *int64  `json:"deleted_at,omitempty"`
}
