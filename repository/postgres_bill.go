package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"receipt-detector/entity"
	"receipt-detector/helper"
	"strconv"
)

type billPostgresRepo struct {
	dbtx DBTX
}

func NewBillPostgresRepo(dbtx DBTX) *billPostgresRepo {
	return &billPostgresRepo{
		dbtx: dbtx,
	}
}

func (r *billPostgresRepo) InsertOne(ctx context.Context, bill entity.Bill) (int64, error) {
	q := `
		INSERT
		INTO bills (bill_name, bill_date, result_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING bill_id
	`

	var billId int64

	err := r.dbtx.QueryRowContext(ctx, q, bill.BillName, bill.BillDate, bill.ResultId, helper.NowUnixMilli()).Scan(&billId)
	if err != nil {
		return billId, fmt.Errorf("repository][billPostgresRepo][InsertOne][dbtx.ExecContext] %w", err)
	}

	return billId, nil
}

func (r *billPostgresRepo) GetByBillId(ctx context.Context, billId int64) (*entity.Bill, error) {
	q := `
		SELECT bill_id, bill_name, bill_date, result_id, created_at, updated_at
		FROM bills
		WHERE bill_id = $1
			AND deleted_at IS NULL
	`

	var bill entity.Bill

	err := r.dbtx.QueryRowContext(ctx, q, billId).Scan(
		&bill.BillId,
		&bill.BillName,
		&bill.BillDate,
		&bill.ResultId,
		&bill.CreatedAt,
		&bill.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("repository][billPostgresRepo][GetByBillId][dbtx.QueryRowContext] %w", err)
	}

	return &bill, nil
}

func (r *billPostgresRepo) UpdateBill(ctx context.Context, newBill entity.UpdateBillRequest) error {
	q := `
		UPDATE bills
		SET 
	`

	args := []any{}
	i := 1

	if newBill.BillName != nil {
		q += `bill_name = $` + strconv.Itoa(i) + `, `
		args = append(args, *newBill.BillName)
		i++
	}

	if newBill.BillDate != nil {
		q += `bill_date = $` + strconv.Itoa(i) + `, `
		args = append(args, *newBill.BillDate)
		i++
	}

	q += `updated_at = $` + strconv.Itoa(i) +
		`WHERE bill_id = $` + strconv.Itoa(i+1)
	args = append(args, helper.NowUnixMilli())
	args = append(args, newBill.BillId)

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("repository][billPostgresRepo][UpdateBillName][dbtx.ExecContext] %w", err)
	}

	return nil
}
