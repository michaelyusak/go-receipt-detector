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

type receiptItems struct {
	dbtx repository.DBTX
}

func NewReceiptItems(dbtx repository.DBTX) *receiptItems {
	return &receiptItems{
		dbtx: dbtx,
	}
}

func (r *receiptItems) NewTx(tx *sql.Tx) repository.ReceiptItems {
	return &receiptItems{
		dbtx: tx,
	}
}

func (r *receiptItems) InsertMany(ctx context.Context, receiptItems []entity.ReceiptItem) error {
	q := `
		INSERT
		INTO receipt_items (receipt_id, item_category, item_name, item_quantity, item_price_currency, item_price_numeric, created_at)
		VALUES
	`

	now := helper.NowUnixMilli()
	args := []any{}

	for i, receiptItem := range receiptItems {
		offset := i * 6

		receiptIdIdx := strconv.Itoa(offset + 1)
		itemCategoryIdx := strconv.Itoa(offset + 2)
		itemNameIdx := strconv.Itoa(offset + 3)
		itemQuantityIdx := strconv.Itoa(offset + 4)
		itemPriceCurrencyIdx := strconv.Itoa(offset + 5)
		itemPriceNumericIdx := strconv.Itoa(offset + 6)
		createdAt := strconv.Itoa(int(now))

		q += `($` + receiptIdIdx +
			`, $` + itemCategoryIdx +
			`, $` + itemNameIdx +
			`, $` + itemQuantityIdx +
			`, $` + itemPriceCurrencyIdx +
			`, $` + itemPriceNumericIdx +
			`, ` + createdAt + `)`

		if i < len(receiptItems)-1 {
			q += `, `
		}

		args = append(args, receiptItem.ReceiptId)
		args = append(args, receiptItem.ItemCategory)
		args = append(args, receiptItem.ItemName)
		args = append(args, receiptItem.ItemQuantity)
		args = append(args, receiptItem.ItemPriceCurrency)
		args = append(args, receiptItem.ItemPriceNumeric)
	}

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("repository][postgres][receiptItems][InsertMany][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *receiptItems) GetByReceiptId(ctx context.Context, receiptId int64) ([]entity.ReceiptItem, error) {
	q := `
		SELECT 
			receipt_item_id,
			receipt_id, 
			item_category, 
			item_name, 
			item_quantity, 
			item_price_currency, 
			item_price_numeric, 
			created_at, 
			updated_at
		FROM receipt_items
		WHERE receipt_id = $1
			AND deleted_at IS NULL
	`

	receiptItems := []entity.ReceiptItem{}

	rows, err := r.dbtx.QueryContext(ctx, q, receiptId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return receiptItems, nil
		}

		return nil, fmt.Errorf("repository][postgres][receiptItems][GetByReceiptId][dbtx.QueryContext] %w", err)
	}

	for rows.Next() {
		var receiptItem entity.ReceiptItem

		err = rows.Scan(
			&receiptItem.ReceiptItemId,
			&receiptItem.ReceiptId,
			&receiptItem.ItemCategory,
			&receiptItem.ItemName,
			&receiptItem.ItemQuantity,
			&receiptItem.ItemPriceCurrency,
			&receiptItem.ItemPriceNumeric,
			&receiptItem.CreatedAt,
			&receiptItem.UpdatedAt,
		)
		if err != nil {
			return receiptItems, fmt.Errorf("repository][postgres][receiptItems][GetByReceiptId][rows.Scan] %w", err)
		}

		receiptItems = append(receiptItems, receiptItem)
	}

	return receiptItems, nil
}
