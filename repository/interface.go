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

type ReceiptDetectionHistories interface {
	NewTx(tx *sql.Tx) ReceiptDetectionHistories
	InsertOne(ctx context.Context, history entity.ReceiptDetectionHistory) error
	GetByResultId(ctx context.Context, resultId string) (*entity.ReceiptDetectionHistory, error)
}

type ReceiptDetectionResults interface {
	InsertOne(ctx context.Context, result []entity.OcrEngineItemDetail) (string, error)
	GetByResultId(ctx context.Context, resultId string) ([]entity.OcrEngineItemDetail, error)
}

type ReceiptImages interface {
	StoreOne(ctx context.Context, contentType string, fileHeader *multipart.FileHeader) (string, error)
	GetImageUrl(ctx context.Context, filePath string) (string, error)
}

type CacheRepository interface {
	SetCache(ctx context.Context, key string, data []byte, duration time.Duration) error
	GetCache(ctx context.Context, key string) ([]byte, error)

	SetReceiptDetectionResult(ctx context.Context, detectionResult entity.ReceiptDetectionResult) error
	GetReceiptDetectionResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error)

	SetReceiptCache(ctx context.Context, receipt entity.Receipt) error
	GetReceiptCache(ctx context.Context, receiptId int64) (*entity.Receipt, error)

	SetReceiptItemsCache(ctx context.Context, receiptId int64, receiptItems []entity.ReceiptItem) error
	GetReceiptItemsCache(ctx context.Context, receiptId int64) ([]entity.ReceiptItem, error)
}

type Receipts interface {
	NewTx(tx *sql.Tx) Receipts
	InsertOne(ctx context.Context, receipt entity.Receipt) (int64, error)
	GetByReceiptId(ctx context.Context, receiptId int64, deviceId string) (*entity.Receipt, error)
	UpdateReceipt(ctx context.Context, newReceipt entity.UpdateReceiptRequest) error
}

type ReceiptItems interface {
	NewTx(tx *sql.Tx) ReceiptItems
	InsertMany(ctx context.Context, receiptItems []entity.ReceiptItem) error
	GetByReceiptId(ctx context.Context, receiptId int64) ([]entity.ReceiptItem, error)
}

type Transaction interface {
	Begin() (*sql.Tx, error)
	Rollback() error
	Commit() error
}
