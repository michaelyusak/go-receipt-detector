package repository

import (
	"context"
	"database/sql"
	"errors"
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
	q := `
		INSERT 
		INTO receipt_detection_histories (image_path, result_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.dbtx.ExecContext(ctx, q, history.ImagePath, history.ResultId, helper.NowUnixMilli())
	if err != nil {
		return fmt.Errorf("[repository][receiptDetectionHistoriesRepo][InsertOne][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *receiptDetectionHistoriesPostgresRepo) GetByResultId(ctx context.Context, resultId string) (*entity.ReceiptDetectionHistory, error) {
	q := `
		SELECT receipt_detection_history_id, image_path, result_id, revision_id, is_approved, is_reviewed, created_at, updated_at
		FROM receipt_detection_histories
		WHERE result_id = $1
			OR revision_id = $1
			AND deleted_at IS NULL
	`

	var receiptDetectionHistory entity.ReceiptDetectionHistory

	err := r.dbtx.QueryRowContext(ctx, q, resultId).Scan(
		&receiptDetectionHistory.HistoryId,
		&receiptDetectionHistory.ImagePath,
		&receiptDetectionHistory.ResultId,
		&receiptDetectionHistory.RevisionId,
		&receiptDetectionHistory.IsApproced,
		&receiptDetectionHistory.IsReviewed,
		&receiptDetectionHistory.CreatedAt,
		&receiptDetectionHistory.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[repository][receiptDetectionHistoriesRepo][GetByResultId][dbtx.QueryRowContext] %w", err)
	}

	return &receiptDetectionHistory, nil
}
