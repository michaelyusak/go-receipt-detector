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

type billItemsPostgresRepo struct {
	dbtx DBTX
}

func NewBillItemsPostgresRepo(dbtx DBTX) *billItemsPostgresRepo {
	return &billItemsPostgresRepo{
		dbtx: dbtx,
	}
}

func (r *billItemsPostgresRepo) InsertMany(ctx context.Context, billItems []entity.BillItem) error {
	q := `
		INSERT
		INTO bill_items (bill_id, item_category, item_name, item_quantity, item_price_currency, item_price_numeric, created_at)
		VALUES
	`

	now := helper.NowUnixMilli()
	args := []any{}

	for i, billItem := range billItems {
		offset := i * 6

		billIdIdx := strconv.Itoa(offset + 1)
		itemCategoryIdx := strconv.Itoa(offset + 2)
		itemNameIdx := strconv.Itoa(offset + 3)
		itemQuantityIdx := strconv.Itoa(offset + 4)
		itemPriceCurrencyIdx := strconv.Itoa(offset + 5)
		itemPriceNumericIdx := strconv.Itoa(offset + 6)
		createdAt := strconv.Itoa(int(now))

		q += `($` + billIdIdx +
			`, $` + itemCategoryIdx +
			`, $` + itemNameIdx +
			`, $` + itemQuantityIdx +
			`, $` + itemPriceCurrencyIdx +
			`, $` + itemPriceNumericIdx +
			`, ` + createdAt + `)`

		if i < len(billItems)-1 {
			q += `, `
		}

		args = append(args, billItem.BillId)
		args = append(args, billItem.ItemCategory)
		args = append(args, billItem.ItemName)
		args = append(args, billItem.ItemQuantity)
		args = append(args, billItem.ItemPriceCurrency)
		args = append(args, billItem.ItemPriceNumeric)
	}

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("repository][billItemsPostgresRepo][InsertMany][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *billItemsPostgresRepo) GetByBillId(ctx context.Context, billId int64) ([]entity.BillItem, error) {
	q := `
		SELECT 
			bill_item_id,
			bill_id, 
			item_category, 
			item_name, 
			item_quantity, 
			item_price_currency, 
			item_price_numeric, 
			created_at, 
			updated_at
		FROM bill_items
		WHERE bill_id = $1
			AND deleted_at IS NULL
	`

	billItems := []entity.BillItem{}

	rows, err := r.dbtx.QueryContext(ctx, q, billId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return billItems, nil
		}

		return nil, fmt.Errorf("repository][billItemsPostgresRepo][GetByBillId][dbtx.QueryContext] %w", err)
	}

	for rows.Next() {
		var billItem entity.BillItem

		err = rows.Scan(
			&billItem.BillItemId,
			&billItem.BillId,
			&billItem.ItemCategory,
			&billItem.ItemName,
			&billItem.ItemQuantity,
			&billItem.ItemPriceCurrency,
			&billItem.ItemPriceNumeric,
			&billItem.CreatedAt,
			&billItem.UpdatedAt,
		)
		if err != nil {
			return billItems, fmt.Errorf("repository][billItemsPostgresRepo][GetByBillId][rows.Scan] %w", err)
		}

		billItems = append(billItems, billItem)
	}

	return billItems, nil
}
