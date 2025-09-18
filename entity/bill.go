package entity

type Bill struct {
	BillId       int64  `json:"bill_id,omitempty"`
	BillName     string `json:"bill_name" binding:"required"`
	BillDate     int64  `json:"bill_date"`
	BillImageUrl string `json:"bill_image_url"`
	ResultId     string `json:"result_id" binding:"required"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    *int64 `json:"updated_at"`
	DeletedAt    *int64 `json:"deleted_at,omitempty"`
}

type CreateBillResponse struct {
	BillId int64 `json:"bill_id"`
}

type DetectionResult []OcrEngineItemDetail

type CreateBillRequest struct {
	Bill            Bill            `json:"bill" binding:"required"`
	DetectionResult DetectionResult `json:"detection_result" binding:"required"`
}

type GetByBillIdResponse struct {
	Bill      Bill       `json:"bill"`
	BillItems []BillItem `json:"bill_items"`
}

type UpdateBillRequest struct {
	BillId   int64
	BillName *string `json:"bill_name" binding:"required_without=BillDate"`
	BillDate *int64  `json:"bill_date" binding:"required_without=BillName"`
}
