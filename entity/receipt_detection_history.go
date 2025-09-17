package entity

type ReceiptDetectionHistory struct {
	HistoryId  int64
	ImagePath  string
	ResultId   string
	RevisionId string
	IsApproced bool
	IsReviewed bool
	CreatedAt  int64
	UpdatedAt  *int64
	DeletedAt  *int64
}
