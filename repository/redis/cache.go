package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"receipt-detector/entity"
	"time"

	"github.com/redis/go-redis/v9"
)

type cache struct {
	client *redis.Client

	receiptDetectionResultTTL time.Duration
	receiptTTL                time.Duration
	receiptItemsTTL           time.Duration

	logTag string
}

type CacheOpt struct {
	Client                    *redis.Client
	ReceiptDetectionResultTTL time.Duration
	ReceiptTTL                time.Duration
	ReceiptItemsTTL           time.Duration
}

func NewCache(opt CacheOpt) *cache {
	return &cache{
		client: opt.Client,

		receiptDetectionResultTTL: opt.ReceiptDetectionResultTTL,
		receiptTTL:                opt.ReceiptTTL,
		receiptItemsTTL:           opt.ReceiptItemsTTL,

		logTag: "[repository][redis][cache]",
	}
}

func (r *cache) Set(ctx context.Context, key string, data []byte, duration time.Duration) error {
	logTag := r.logTag + "[Set]"

	err := r.client.Set(ctx, key, data, duration).Err()
	if err != nil {
		return fmt.Errorf("%s[client.Set] Failed to set data: %w [key: %s][data: %s]", logTag, err, key, string(data))
	}

	return nil
}

func (r *cache) Get(ctx context.Context, key string) ([]byte, error) {
	logTag := r.logTag + "[Get]"

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s[client.Get] Failed to set data: %w [key: %s]", logTag, err, key)
	}

	return []byte(val), nil
}

func (r *cache) receiptDetectionResultKey(resultId string) string {
	return fmt.Sprintf("receipt_detection_result:%s", resultId)
}

func (r *cache) SetReceiptDetectionResult(ctx context.Context, detectionResult entity.ReceiptDetectionResult) error {
	logTag := r.logTag + "[SetReceiptDetectionResult]"

	data, err := json.Marshal(detectionResult)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal detection result: %w [result: %+v]", logTag, err, detectionResult)
	}

	return r.Set(
		ctx,
		r.receiptDetectionResultKey(detectionResult.ResultId),
		data,
		r.receiptDetectionResultTTL,
	)
}

func (r *cache) GetReceiptDetectionResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error) {
	logTag := r.logTag + "[GetReceiptDetectionResult]"

	data, err := r.Get(ctx, r.receiptDetectionResultKey(resultId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.Get] Failed to get cache: %w [result_id: %s]", logTag, err, resultId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var detectionResult entity.ReceiptDetectionResult

	err = json.Unmarshal(data, &detectionResult)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logTag, err, data)
	}

	return &detectionResult, nil
}

func (r *cache) receiptKey(receiptId int64) string {
	return fmt.Sprintf("receipt:%v", receiptId)
}

func (r *cache) SetReceipt(ctx context.Context, receipt entity.Receipt) error {
	logTag := r.logTag + "[SetReceipt]"

	data, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal receipt: %w [receipt_id: %v]", logTag, err, receipt.ReceiptId)
	}

	return r.Set(
		ctx,
		r.receiptKey(receipt.ReceiptId),
		data,
		r.receiptTTL,
	)
}

func (r *cache) GetReceipt(ctx context.Context, receiptId int64) (*entity.Receipt, error) {
	logTag := r.logTag + "[GetReceipt]"

	data, err := r.Get(ctx, r.receiptKey(receiptId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.Get] Failed to get cache: %w [receipt_id: %v]", logTag, err, receiptId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var receipt entity.Receipt

	err = json.Unmarshal(data, &receipt)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logTag, err, data)
	}

	return &receipt, nil
}

func (r *cache) receiptItemsKey(receiptId int64) string {
	return fmt.Sprintf("receipt_items:%v", receiptId)
}

func (r *cache) SetReceiptItems(ctx context.Context, receiptId int64, receiptItems []entity.ReceiptItem) error {
	logTag := r.logTag + "[SetReceiptItems]"

	data, err := json.Marshal(receiptItems)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal receipt: %w [receipt_id: %v]", logTag, err, receiptId)
	}

	return r.Set(
		ctx,
		r.receiptItemsKey(receiptId),
		data,
		r.receiptItemsTTL,
	)
}

func (r *cache) GetReceiptItems(ctx context.Context, receiptId int64) ([]entity.ReceiptItem, error) {
	logTag := r.logTag + "[GetReceiptItems]"

	data, err := r.Get(ctx, r.receiptItemsKey(receiptId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.Get] Failed to get cache: %w [receipt_id: %v]", logTag, err, receiptId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var receiptItem []entity.ReceiptItem

	err = json.Unmarshal(data, &receiptItem)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logTag, err, data)
	}

	return receiptItem, nil
}
