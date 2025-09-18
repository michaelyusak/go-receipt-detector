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

type bill struct {
	billRepo                    repository.BillRepository
	billItemRepo                repository.BillItemRepository
	receiptDetectionHistoryRepo repository.ReceiptDetectionHistoriesRepository
	receiptImageRepo            repository.ReceiptImageRepository
	cacheRepo                   repository.CacheRepository

	logHeading string
}

type BillOpts struct {
	BillRepo                    repository.BillRepository
	BillItemRepo                repository.BillItemRepository
	ReceiptDetectionHistoryRepo repository.ReceiptDetectionHistoriesRepository
	ReceiptImageRepo            repository.ReceiptImageRepository
	CacheRepo                   repository.CacheRepository
}

func NewBillService(opt BillOpts) *bill {
	return &bill{
		billRepo:                    opt.BillRepo,
		billItemRepo:                opt.BillItemRepo,
		receiptDetectionHistoryRepo: opt.ReceiptDetectionHistoryRepo,
		receiptImageRepo:            opt.ReceiptImageRepo,
		cacheRepo:                   opt.CacheRepo,

		logHeading: "[service][bill]",
	}
}

func (s *bill) convertDetectionResultToBillItems(detectionResult entity.DetectionResult, billId int64) []entity.BillItem {
	billItems := []entity.BillItem{}

	for _, res := range detectionResult {
		var billItem entity.BillItem

		billItem.BillId = billId
		billItem.ItemCategory = res.Category
		billItem.ItemName = res.Info.Item
		billItem.ItemQuantity = res.Info.Qty
		billItem.ItemPriceCurrency = res.Info.Price.Currency
		billItem.ItemPriceNumeric = res.Info.Price.Numeric

		billItems = append(billItems, billItem)
	}

	return billItems
}

func (s *bill) CreateOne(ctx context.Context, bill entity.Bill, detectionResult entity.DetectionResult) (int64, error) {
	logHeading := s.logHeading + "[CreateOne]"

	if bill.BillDate == 0 {
		bill.BillDate = helper.NowUnixMilli()
	}

	billId, err := s.billRepo.InsertOne(ctx, bill)
	if err != nil {
		return 0, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[billRepo.InsertOne] Failed to insert bill to postgres: %v [result_id: %s]", logHeading, err, bill.ResultId),
		})
	}

	billItems := s.convertDetectionResultToBillItems(detectionResult, billId)

	err = s.billItemRepo.InsertMany(ctx, billItems)
	if err != nil {
		return 0, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[billItemRepo.InsertMany] Failed to insert bill items to postgres %v [bill_id: %v]", logHeading, err, billId),
		})
	}

	return billId, nil
}

func (s *bill) GetByBillId(ctx context.Context, billId int64) (*entity.Bill, []entity.BillItem, error) {
	logHeading := s.logHeading + "[GetByBillId]"

	bill, err := s.billRepo.GetByBillId(ctx, billId)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[billRepo.GetByBillId] Failed to get bill: %v [bill_id: %v]", logHeading, err, billId),
		})
	}
	if bill == nil {
		return nil, nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusNotFound,
			ResponseMessage: "Bill not found",
		})
	}

	billItems, err := s.billItemRepo.GetByBillId(ctx, billId)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[billItemRepo.GetByBillId] Failed to get bill items: %v [bill_id: %v]", logHeading, err, billId),
		})
	}

	history, err := s.receiptDetectionHistoryRepo.GetByResultId(ctx, bill.ResultId)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptDetectionHistoryRepo.GetByResultId] Failed to get detection history: %v [bill_id: %v]", logHeading, err, billId),
		})
	}

	imageUrl, err := s.receiptImageRepo.GetImageUrl(ctx, history.ImagePath)
	if err != nil {
		return nil, nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptImageRepo.GetImageUrl] Failed to get image url: %v [bill_id: %v]", logHeading, err, billId),
		})
	}

	bill.BillImageUrl = imageUrl

	return bill, billItems, nil
}

func (s *bill) UpdateBill(ctx context.Context, newBill entity.UpdateBillRequest) error {
	return s.billRepo.UpdateBill(ctx, newBill)
}
