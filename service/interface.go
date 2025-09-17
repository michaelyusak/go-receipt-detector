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
