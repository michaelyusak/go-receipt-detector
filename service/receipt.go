package service

import (
	"context"
	"fmt"
	"net/http"
	"receipt-detector/entity"
	"receipt-detector/helper"
	"receipt-detector/repository"

	hApperror "github.com/michaelyusak/go-helper/apperror"
)

type receipt struct {
	receiptsRepo                  repository.Receipts
	receiptItemsRepo              repository.ReceiptItems
	receiptDetectionHistoriesRepo repository.ReceiptDetectionHistories
	receiptImagesRepo             repository.ReceiptImages
	cacheRepo                     repository.CacheRepository

	logHeading string
}

type ReceiptOpts struct {
	ReceiptsRepo                  repository.Receipts
	ReceiptItemsRepo              repository.ReceiptItems
	ReceiptDetectionHistoriesRepo repository.ReceiptDetectionHistories
	ReceiptImagesRepo             repository.ReceiptImages
	CacheRepo                     repository.CacheRepository
}

func NewBillService(opt ReceiptOpts) *receipt {
	return &receipt{
		receiptsRepo:                  opt.ReceiptsRepo,
		receiptItemsRepo:              opt.ReceiptItemsRepo,
		receiptDetectionHistoriesRepo: opt.ReceiptDetectionHistoriesRepo,
		receiptImagesRepo:             opt.ReceiptImagesRepo,
		cacheRepo:                     opt.CacheRepo,

		logHeading: "[service][receipt]",
	}
}

func (s *receipt) convertDetectionResultToReceiptItems(detectionResult entity.DetectionResult, receiptId int64) []entity.ReceiptItem {
	receiptItems := []entity.ReceiptItem{}

	for _, res := range detectionResult {
		var receiptItem entity.ReceiptItem

		receiptItem.ReceiptId = receiptId
		receiptItem.ItemCategory = res.Category
		receiptItem.ItemName = res.Info.Item
		receiptItem.ItemQuantity = res.Info.Qty
		receiptItem.ItemPriceCurrency = res.Info.Price.Currency
		receiptItem.ItemPriceNumeric = res.Info.Price.Numeric

		receiptItems = append(receiptItems, receiptItem)
	}

	return receiptItems
}

func (s *receipt) CreateOne(ctx context.Context, receipt entity.Receipt, detectionResult entity.DetectionResult) (int64, error) {
	logHeading := s.logHeading + "[CreateOne]"

	if receipt.ReceiptDate == 0 {
		receipt.ReceiptDate = helper.NowUnixMilli()
	}

	receiptId, err := s.receiptsRepo.InsertOne(ctx, receipt)
	if err != nil {
		return 0, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptsRepo.InsertOne] Failed to insert receipt to postgres: %v [result_id: %s]", logHeading, err, receipt.ResultId),
		})
	}

	receiptItems := s.convertDetectionResultToReceiptItems(detectionResult, receiptId)

	err = s.receiptItemsRepo.InsertMany(ctx, receiptItems)
	if err != nil {
		return 0, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptItemsRepo.InsertMany] Failed to insert receipt items to postgres %v [receipt_id: %v]", logHeading, err, receiptId),
		})
	}

	return receiptId, nil
}

func (s *receipt) GetByReceiptId(ctx context.Context, receiptId int64) (*entity.Receipt, []entity.ReceiptItem, error) {
	logHeading := s.logHeading + "[GetByReceiptId]"

	receipt, err := s.receiptsRepo.GetByReceiptId(ctx, receiptId)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptsRepo.GetByReceiptId] Failed to get receipt: %v [receipt_id: %v]", logHeading, err, receiptId),
		})
	}
	if receipt == nil {
		return nil, nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusNotFound,
			ResponseMessage: "Receipt not found",
		})
	}

	receiptItems, err := s.receiptItemsRepo.GetByReceiptId(ctx, receiptId)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[billItemRepo.GetByBillId] Failed to get receipt items: %v [receipt_id: %v]", logHeading, err, receiptId),
		})
	}

	history, err := s.receiptDetectionHistoriesRepo.GetByResultId(ctx, receipt.ResultId)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptDetectionHistoriesRepo.GetByResultId] Failed to get detection history: %v [receipt_id: %v]", logHeading, err, receiptId),
		})
	}

	imageUrl, err := s.receiptImagesRepo.GetImageUrl(ctx, history.ImagePath)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptImagesRepo.GetImageUrl] Failed to get image url: %v [receipt_id: %v]", logHeading, err, receiptId),
		})
	}

	receipt.ReceiptImageUrl = imageUrl

	return receipt, receiptItems, nil
}

func (s *receipt) UpdateReceipt(ctx context.Context, newReceipt entity.UpdateReceiptRequest) error {
	return s.receiptsRepo.UpdateReceipt(ctx, newReceipt)
}
