package entity

type Receipt struct {
	ReceiptId       int64  `json:"receipt_id,omitempty"`
	ReceiptName     string `json:"receipt_name" binding:"required"`
	ReceiptDate     int64  `json:"receipt_date"`
	ReceiptImageUrl string `json:"receipt_image_url"`
	ResultId        string `json:"result_id" binding:"required"`
	DeviceId        string `json:"device_id,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       *int64 `json:"updated_at"`
	DeletedAt       *int64 `json:"deleted_at,omitempty"`
}

type CreateReceiptResponse struct {
	ReceiptId int64 `json:"receipt_id"`
}

type DetectionResult []OcrEngineItemDetail

type CreateReceiptRequest struct {
	Receipt         Receipt         `json:"receipt" binding:"required"`
	DetectionResult DetectionResult `json:"detection_result" binding:"required"`
}

type GetByReceiptIdResponse struct {
	Receipt      Receipt       `json:"receipt"`
	ReceiptItems []ReceiptItem `json:"receipt_items"`
}

type UpdateReceiptRequest struct {
	ReceiptId   int64
	ReceiptName *string `json:"receipt_name" binding:"required_without=ReceiptDate"`
	ReceiptDate *int64  `json:"receipt_date" binding:"required_without=ReceiptName"`
}
