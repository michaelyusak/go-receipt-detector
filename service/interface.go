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

type Receipt interface {
	CreateOne(ctx context.Context, bill entity.Receipt, detectionResult entity.DetectionResult) (int64, error)
	GetByReceiptId(ctx context.Context, billId int64) (*entity.Receipt, []entity.ReceiptItem, error)
	UpdateOne(ctx context.Context, newBill entity.UpdateReceiptRequest) error
}
