package repository

import (
	"context"
	"database/sql"
	"mime/multipart"
	"receipt-detector/entity"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type ReceiptDetectionHistoriesRepository interface {
	InsertOne(ctx context.Context, history entity.ReceiptDetectionHistory) error
}

type ReceiptDetectionResultsRepository interface {
	InserOne(ctx context.Context, result []entity.OcrEngineItemDetail) (string, error)
}

type ReceiptImageRepository interface {
	StoreOne(ctx context.Context, contentType string, fileHeader *multipart.FileHeader) (string, error)
}
