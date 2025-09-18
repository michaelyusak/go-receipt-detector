package handler

import (
	"receipt-detector/entity"
	"receipt-detector/service"
	"strconv"

	"github.com/gin-gonic/gin"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	"github.com/michaelyusak/go-helper/helper"
)

type Receipt struct {
	receiptService service.Receipt
}

func NewReceipt(receiptService service.Receipt) *Receipt {
	return &Receipt{
		receiptService: receiptService,
	}
}

func (h *Receipt) Create(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var req entity.CreateReceiptRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	receiptId, err := h.receiptService.CreateOne(ctx.Request.Context(), req.Receipt, req.DetectionResult)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, entity.CreateReceiptResponse{
		ReceiptId: receiptId,
	})
}

func (h *Receipt) GetByReceiptId(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	receiptIdStr := ctx.Param("receipt_id")
	if receiptIdStr == "" {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "receipt_id must be provided",
		}))
		return
	}

	receiptId, err := strconv.Atoi(receiptIdStr)
	if err != nil {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "receipt_id must be a number",
		}))
		return
	}

	receipt, receiptItems, err := h.receiptService.GetByReceiptId(ctx.Request.Context(), int64(receiptId))
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, entity.GetByReceiptIdResponse{
		Receipt:      *receipt,
		ReceiptItems: receiptItems,
	})
}

func (h *Receipt) UpdateReceipt(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	receiptIdStr := ctx.Param("receipt_id")
	if receiptIdStr == "" {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "receipt_id must be provided",
		}))
		return
	}

	receiptId, err := strconv.Atoi(receiptIdStr)
	if err != nil {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "receipt_id must be a number",
		}))
		return
	}

	var req entity.UpdateReceiptRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	req.ReceiptId = int64(receiptId)

	err = h.receiptService.UpdateReceipt(ctx, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, nil)
}
