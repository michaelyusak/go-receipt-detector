package entity

type ReceiptDetectionDocument struct {
	Result []OcrEngineItemDetail
}

type ReceiptDetectionResult struct {
	ResultId string                `json:"result_id"`
	ImageUrl string                `json:"image_url,omitempty"`
	Result   []OcrEngineItemDetail `json:"result"`
}
