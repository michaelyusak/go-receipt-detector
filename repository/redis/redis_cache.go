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

type cacheRedisRepo struct {
	client *redis.Client

	receiptDetectionResultCacheDuration time.Duration
	receiptCacheDuration                time.Duration
	receiptItemsCacheDuration           time.Duration

	logHeading string
}

type CacheRedisRepoOpt struct {
	Client                              *redis.Client
	ReceiptDetectionResultCacheDuration time.Duration
	ReceiptCacheDuration                time.Duration
	ReceiptItemsCacheDuration           time.Duration
}

func NewCacheRedisRepo(opt CacheRedisRepoOpt) *cacheRedisRepo {
	return &cacheRedisRepo{
		client: opt.Client,

		receiptDetectionResultCacheDuration: opt.ReceiptDetectionResultCacheDuration,
		receiptCacheDuration:                opt.ReceiptCacheDuration,
		receiptItemsCacheDuration:           opt.ReceiptItemsCacheDuration,

		logHeading: "[repository][cacheRedisRepo]",
	}
}

func (r *cacheRedisRepo) SetCache(ctx context.Context, key string, data []byte, duration time.Duration) error {
	logHeading := r.logHeading + "[SetCache]"

	err := r.client.Set(ctx, key, data, duration).Err()
	if err != nil {
		return fmt.Errorf("%s[client.Set] Failed to set data: %w [key: %s][data: %s]", logHeading, err, key, string(data))
	}

	return nil
}

func (r *cacheRedisRepo) GetCache(ctx context.Context, key string) ([]byte, error) {
	logHeading := r.logHeading + "[GetCache]"

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s[client.Get] Failed to set data: %w [key: %s]", logHeading, err, key)
	}

	return []byte(val), nil
}

func (r *cacheRedisRepo) receiptDetectionResultCacheKey(resultId string) string {
	return fmt.Sprintf("receipt_detection_result:%s", resultId)
}

func (r *cacheRedisRepo) SetReceiptDetectionResult(ctx context.Context, detectionResult entity.ReceiptDetectionResult) error {
	logHeading := r.logHeading + "[SetReceiptDetectionResult]"

	data, err := json.Marshal(detectionResult)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal detection result: %w [result: %+v]", logHeading, err, detectionResult)
	}

	return r.SetCache(
		ctx,
		r.receiptDetectionResultCacheKey(detectionResult.ResultId),
		data,
		r.receiptDetectionResultCacheDuration,
	)
}

func (r *cacheRedisRepo) GetReceiptDetectionResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error) {
	logHeading := r.logHeading + "[GetReceiptDetectionResult]"

	data, err := r.GetCache(ctx, r.receiptDetectionResultCacheKey(resultId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.GetCache] Failed to get cache: %w [result_id: %s]", logHeading, err, resultId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var detectionResult entity.ReceiptDetectionResult

	err = json.Unmarshal(data, &detectionResult)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logHeading, err, data)
	}

	return &detectionResult, nil
}

func (r *cacheRedisRepo) receiptCacheKey(receiptId int64) string {
	return fmt.Sprintf("receipt:%v", receiptId)
}

func (r *cacheRedisRepo) SetReceiptCache(ctx context.Context, receipt entity.Receipt) error {
	logHeading := r.logHeading + "[SetReceiptCache]"

	data, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal receipt: %w [receipt_id: %v]", logHeading, err, receipt.ReceiptId)
	}

	return r.SetCache(
		ctx,
		r.receiptCacheKey(receipt.ReceiptId),
		data,
		r.receiptCacheDuration,
	)
}

func (r *cacheRedisRepo) GetReceiptCache(ctx context.Context, receiptId int64) (*entity.Receipt, error) {
	logHeading := r.logHeading + "[GetReceiptCache]"

	data, err := r.GetCache(ctx, r.receiptCacheKey(receiptId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.GetCache] Failed to get cache: %w [receipt_id: %v]", logHeading, err, receiptId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var receipt entity.Receipt

	err = json.Unmarshal(data, &receipt)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logHeading, err, data)
	}

	return &receipt, nil
}

func (r *cacheRedisRepo) receiptItemsCacheKey(receiptId int64) string {
	return fmt.Sprintf("receipt_items:%v", receiptId)
}

func (r *cacheRedisRepo) SetReceiptItemsCache(ctx context.Context, receiptId int64, receiptItems []entity.ReceiptItem) error {
	logHeading := r.logHeading + "[SetReceiptItemsCache]"

	data, err := json.Marshal(receiptItems)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal receipt: %w [receipt_id: %v]", logHeading, err, receiptId)
	}

	return r.SetCache(
		ctx,
		r.receiptItemsCacheKey(receiptId),
		data,
		r.receiptItemsCacheDuration,
	)
}

func (r *cacheRedisRepo) GetReceiptItemsCache(ctx context.Context, receiptId int64) ([]entity.ReceiptItem, error) {
	logHeading := r.logHeading + "[GetReceiptItemsCache]"

	data, err := r.GetCache(ctx, r.receiptItemsCacheKey(receiptId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.GetCache] Failed to get cache: %w [receipt_id: %v]", logHeading, err, receiptId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var receiptItem []entity.ReceiptItem

	err = json.Unmarshal(data, &receiptItem)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logHeading, err, data)
	}

	return receiptItem, nil
}
