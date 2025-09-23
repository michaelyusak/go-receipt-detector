package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"receipt-detector/entity"
	"receipt-detector/helper"
	"receipt-detector/repository"
)

type receiptDetectionHistories struct {
	dbtx repository.DBTX
}

func NewReceiptDetectionHistories(dbtx repository.DBTX) *receiptDetectionHistories {
	return &receiptDetectionHistories{
		dbtx: dbtx,
	}
}

func (r *receiptDetectionHistories) NewTx(tx *sql.Tx) repository.ReceiptDetectionHistories {
	return &receiptDetectionHistories{
		dbtx: tx,
	}
}

func (r *receiptDetectionHistories) InsertOne(ctx context.Context, history entity.ReceiptDetectionHistory) error {
	q := `
		INSERT 
		INTO receipt_detection_histories (image_path, result_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.dbtx.ExecContext(ctx, q, history.ImagePath, history.ResultId, helper.NowUnixMilli())
	if err != nil {
		return fmt.Errorf("[repository][receiptDetectionHistories][InsertOne][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *receiptDetectionHistories) GetByResultId(ctx context.Context, resultId string) (*entity.ReceiptDetectionHistory, error) {
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

		return nil, fmt.Errorf("[repository][receiptDetectionHistories][GetByResultId][dbtx.QueryRowContext] %w", err)
	}

	return &receiptDetectionHistory, nil
}
