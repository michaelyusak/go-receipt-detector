package repository

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

	logHeading string
}

func NewCacheRedisRepo(client *redis.Client, receiptDetectionResultCacheDuration time.Duration) *cacheRedisRepo {
	return &cacheRedisRepo{
		client: client,

		receiptDetectionResultCacheDuration: receiptDetectionResultCacheDuration,

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
