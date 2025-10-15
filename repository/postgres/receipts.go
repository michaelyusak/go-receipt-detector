package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"receipt-detector/entity"
	"receipt-detector/helper"
	"receipt-detector/repository"
	"strconv"
)

type receipts struct {
	dbtx repository.DBTX
}

func NewReceipts(dbtx repository.DBTX) *receipts {
	return &receipts{
		dbtx: dbtx,
	}
}

func (r *receipts) NewTx(tx *sql.Tx) repository.Receipts {
	return &receipts{
		dbtx: tx,
	}
}

func (r *receipts) InsertOne(ctx context.Context, receipt entity.Receipt) (int64, error) {
	q := `
		INSERT
		INTO receipts (receipt_name, receipt_date, result_id, device_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING receipt_id
	`

	var receiptId int64

	err := r.dbtx.QueryRowContext(ctx, q, receipt.ReceiptName, receipt.ReceiptDate, receipt.ResultId, receipt.DeviceId, helper.NowUnixMilli()).Scan(&receiptId)
	if err != nil {
		return receiptId, fmt.Errorf("[repository][postgres][receipts][InsertOne][dbtx.ExecContext] %w", err)
	}

	return receiptId, nil
}

func (r *receipts) GetByReceiptId(ctx context.Context, receiptId int64, deviceId string) (*entity.Receipt, error) {
	q := `
		SELECT receipt_id, receipt_name, receipt_date, result_id, created_at, updated_at
		FROM receipts
		WHERE receipt_id = $1
			AND device_id = $2
			AND deleted_at IS NULL
	`

	var receipt entity.Receipt

	err := r.dbtx.QueryRowContext(ctx, q, receiptId, deviceId).Scan(
		&receipt.ReceiptId,
		&receipt.ReceiptName,
		&receipt.ReceiptDate,
		&receipt.ResultId,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[repository][postgres][receipts][GetByReceiptId][dbtx.QueryRowContext] %w", err)
	}

	return &receipt, nil
}

func (r *receipts) UpdateOne(ctx context.Context, newReceipt entity.UpdateReceiptRequest) error {
	q := `
		UPDATE receipts
		SET 
	`

	args := []any{}
	i := 1

	if newReceipt.ReceiptName != nil {
		q += `receipt_name = $` + strconv.Itoa(i) + `, `
		args = append(args, *newReceipt.ReceiptName)
		i++
	}

	if newReceipt.ReceiptDate != nil {
		q += `receipt_date = $` + strconv.Itoa(i) + `, `
		args = append(args, *newReceipt.ReceiptDate)
		i++
	}

	q += `updated_at = $` + strconv.Itoa(i) +
		` WHERE receipt_id = $` + strconv.Itoa(i+1)
	args = append(args, helper.NowUnixMilli())
	args = append(args, newReceipt.ReceiptId)

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("[repository][postgres][receipts][UpdateOne][dbtx.ExecContext] %w", err)
	}

	return nil
}
