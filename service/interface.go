package service

import (
	"context"
	"mime/multipart"
	"receipt-detector/entity"
)

type ReceiptDetection interface {
	DetectAndStoreReceipt(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*entity.ReceiptDetectionResult, error)
	GetResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error)
}

type Bill interface {
	CreateOne(ctx context.Context, bill entity.Bill, detectionResult entity.DetectionResult) (int64, error)
	GetByBillId(ctx context.Context, billId int64) (*entity.Bill, []entity.BillItem, error)
	UpdateBill(ctx context.Context, newBill entity.UpdateBillRequest) error
}
