package repository

import (
	"context"
	"database/sql"
	"mime/multipart"
	"receipt-detector/entity"
	"time"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type ReceiptDetectionHistoriesRepository interface {
	InsertOne(ctx context.Context, history entity.ReceiptDetectionHistory) error
	GetByResultId(ctx context.Context, resultId string) (*entity.ReceiptDetectionHistory, error)
}

type ReceiptDetectionResults interface {
	InsertOne(ctx context.Context, result []entity.OcrEngineItemDetail) (string, error)
	GetByResultId(ctx context.Context, resultId string) ([]entity.OcrEngineItemDetail, error)
}

type ReceiptImageRepository interface {
	StoreOne(ctx context.Context, contentType string, fileHeader *multipart.FileHeader) (string, error)
	GetImageUrl(ctx context.Context, filePath string) (string, error)
}

type CacheRepository interface {
	SetCache(ctx context.Context, key string, data []byte, duration time.Duration) error
	GetCache(ctx context.Context, key string) ([]byte, error)

	SetReceiptDetectionResult(ctx context.Context, detectionResult entity.ReceiptDetectionResult) error
	GetReceiptDetectionResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error)

	SetBillCache(ctx context.Context, bill entity.Bill) error
	GetBillCache(ctx context.Context, billId int64) (*entity.Bill, error)

	SetBillItemsCache(ctx context.Context, billId int64, billItems []entity.BillItem) error
	GetBillItemsCache(ctx context.Context, billId int64) ([]entity.BillItem, error)
}

type BillRepository interface {
	InsertOne(ctx context.Context, bill entity.Bill) (int64, error)
	GetByBillId(ctx context.Context, billId int64) (*entity.Bill, error)
	UpdateBill(ctx context.Context, newBill entity.UpdateBillRequest) error
}

type BillItemRepository interface {
	InsertMany(ctx context.Context, billItems []entity.BillItem) error
	GetByBillId(ctx context.Context, billId int64) ([]entity.BillItem, error)
}
