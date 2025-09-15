package entity

import "fmt"

type OcrEngineResponse[T any] struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Data       T      `json:"data,omitempty"`
}

type PriceDetail struct {
	Currency string  `json:"currency"`
	Numeric  float64 `json:"numeric"`
}

type OcrEngineItemDetailInfo struct {
	Item  string      `json:"item"`
	Qty   int         `json:"qty,omitempty"`
	Price PriceDetail `json:"price"`
}

type OcrEngineItemDetail struct {
	Category string                  `json:"category"`
	Info     OcrEngineItemDetailInfo `json:"info"`
}

func (p PriceDetail) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"currency":%q,"numeric":%.2f}`, p.Currency, p.Numeric)), nil
}
