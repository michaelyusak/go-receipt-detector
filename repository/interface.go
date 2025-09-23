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

type Cache interface {
	Set(ctx context.Context, key string, data []byte, duration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)

	SetReceiptDetectionResult(ctx context.Context, detectionResult entity.ReceiptDetectionResult) error
	GetReceiptDetectionResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error)

	SetReceipt(ctx context.Context, receipt entity.Receipt) error
	GetReceipt(ctx context.Context, receiptId int64) (*entity.Receipt, error)

	SetReceiptItems(ctx context.Context, receiptId int64, receiptItems []entity.ReceiptItem) error
	GetReceiptItems(ctx context.Context, receiptId int64) ([]entity.ReceiptItem, error)
}

type Receipts interface {
	NewTx(tx *sql.Tx) Receipts
	InsertOne(ctx context.Context, receipt entity.Receipt) (int64, error)
	GetByReceiptId(ctx context.Context, receiptId int64, deviceId string) (*entity.Receipt, error)
	UpdateOne(ctx context.Context, newReceipt entity.UpdateReceiptRequest) error
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

type ReceiptParticipants interface {
	NewTx(tx *sql.Tx) ReceiptParticipants
	InsertMany(ctx context.Context, participants []entity.ReceiptParticipant) error
	GetByReceiptId(ctx context.Context, receiptId int64) ([]entity.ReceiptParticipant, error)
}

type ParticipantContacts interface {
	NewTx(tx *sql.Tx) ParticipantContacts
	InsertMany(ctx context.Context, contacts []entity.ParticipantContact) error
	GetByParticipantId(ctx context.Context, participantId int64) ([]entity.ParticipantContact, error)
}
