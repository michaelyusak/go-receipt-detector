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
	billCacheDuration                   time.Duration
	billItemsCacheDuration              time.Duration

	logHeading string
}

type CacheRedisRepoOpt struct {
	Client                              *redis.Client
	ReceiptDetectionResultCacheDuration time.Duration
	BillCacheDuration                   time.Duration
	BillItemsCacheDuration              time.Duration
}

func NewCacheRedisRepo(opt CacheRedisRepoOpt) *cacheRedisRepo {
	return &cacheRedisRepo{
		client: opt.Client,

		receiptDetectionResultCacheDuration: opt.ReceiptDetectionResultCacheDuration,
		billCacheDuration:                   opt.BillCacheDuration,
		billItemsCacheDuration:              opt.BillItemsCacheDuration,

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

func (r *cacheRedisRepo) billCacheKey(billId int64) string {
	return fmt.Sprintf("bill:%v", billId)
}

func (r *cacheRedisRepo) SetBillCache(ctx context.Context, bill entity.Bill) error {
	logHeading := r.logHeading + "[SetBillCache]"

	data, err := json.Marshal(bill)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal bill: %w [bill_id: %v]", logHeading, err, bill.BillId)
	}

	return r.SetCache(
		ctx,
		r.billCacheKey(bill.BillId),
		data,
		r.billCacheDuration,
	)
}

func (r *cacheRedisRepo) GetBillCache(ctx context.Context, billId int64) (*entity.Bill, error) {
	logHeading := r.logHeading + "[GetBillCache]"

	data, err := r.GetCache(ctx, r.billCacheKey(billId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.GetCache] Failed to get cache: %w [bill_id: %v]", logHeading, err, billId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var bill entity.Bill

	err = json.Unmarshal(data, &bill)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logHeading, err, data)
	}

	return &bill, nil
}

func (r *cacheRedisRepo) billItemsCacheKey(billId int64) string {
	return fmt.Sprintf("bill_items:%v", billId)
}

func (r *cacheRedisRepo) SetBillItemsCache(ctx context.Context, billId int64, billItems []entity.BillItem) error {
	logHeading := r.logHeading + "[SetBillItemsCache]"

	data, err := json.Marshal(billItems)
	if err != nil {
		return fmt.Errorf("%s[json.Marshal] Failed to marshal bill: %w [bill_id: %v]", logHeading, err, billId)
	}

	return r.SetCache(
		ctx,
		r.billItemsCacheKey(billId),
		data,
		r.billItemsCacheDuration,
	)
}

func (r *cacheRedisRepo) GetBillItemsCache(ctx context.Context, billId int64) ([]entity.BillItem, error) {
	logHeading := r.logHeading + "[GetBillItemsCache]"

	data, err := r.GetCache(ctx, r.billItemsCacheKey(billId))
	if err != nil {
		return nil, fmt.Errorf("%s[r.GetCache] Failed to get cache: %w [bill_id: %v]", logHeading, err, billId)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var billItem []entity.BillItem

	err = json.Unmarshal(data, &billItem)
	if err != nil {
		return nil, fmt.Errorf("%s[json.Unmarshal] Failed to unmarshal data: %w [data: %v]", logHeading, err, data)
	}

	return billItem, nil
}
