package handler

import (
	"receipt-detector/entity"
	"receipt-detector/service"
	"strconv"

	"github.com/gin-gonic/gin"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	"github.com/michaelyusak/go-helper/helper"
)

type BillHandler struct {
	billService service.Bill
}

func NewBillHandler(billService service.Bill) *BillHandler {
	return &BillHandler{
		billService: billService,
	}
}

func (h *BillHandler) Create(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var req entity.CreateBillRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	billId, err := h.billService.CreateOne(ctx.Request.Context(), req.Bill, req.DetectionResult)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, entity.CreateBillResponse{
		BillId: billId,
	})
}

func (h *BillHandler) GetById(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	billIdStr := ctx.Param("bill_id")
	if billIdStr == "" {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "bill_id must be provided",
		}))
		return
	}

	billId, err := strconv.Atoi(billIdStr)
	if err != nil {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "bill_id must be a number",
		}))
		return
	}

	bill, billItems, err := h.billService.GetByBillId(ctx.Request.Context(), int64(billId))
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, entity.GetByBillIdResponse{
		Bill:      *bill,
		BillItems: billItems,
	})
}

func (h *BillHandler) UpdateBill(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	billIdStr := ctx.Param("bill_id")
	if billIdStr == "" {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "bill_id must be provided",
		}))
		return
	}

	billId, err := strconv.Atoi(billIdStr)
	if err != nil {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "bill_id must be a number",
		}))
		return
	}

	var req entity.UpdateBillRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	req.BillId = int64(billId)

	err = h.billService.UpdateBill(ctx, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	helper.ResponseOK(ctx, nil)
}
