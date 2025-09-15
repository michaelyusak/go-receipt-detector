package repository

import (
	"context"
	"fmt"
	"receipt-detector/entity"
	"receipt-detector/helper"
)

type receiptDetectionHistoriesPostgresRepo struct {
	dbtx DBTX
}

func NewReceiptDetectionHistoriesPostgresRepo(dbtx DBTX) *receiptDetectionHistoriesPostgresRepo {
	return &receiptDetectionHistoriesPostgresRepo{
		dbtx: dbtx,
	}
}

func (r *receiptDetectionHistoriesPostgresRepo) InsertOne(ctx context.Context, history entity.ReceiptDetectionHistory) error {
	sql := `
		INSERT 
		INTO receipt_detection_histories (image_path, result_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.dbtx.ExecContext(ctx, sql, history.ImagePath, history.ResultId, helper.NowUnixMilli())
	if err != nil {
		return fmt.Errorf("[repository][receiptDetectionHistoriesRepo][InsertOne][dbtx.ExecContext] %w", err)
	}

	return nil
}
