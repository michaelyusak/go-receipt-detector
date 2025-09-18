package handler

import (
	"fmt"
	"net/http"
	"receipt-detector/service"

	"github.com/gin-gonic/gin"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

type ReceiptDetection struct {
	receiptDetectionService service.ReceiptDetection
}

func NewReceiptDetection(receiptDetectionService service.ReceiptDetection) *ReceiptDetection {
	return &ReceiptDetection{
		receiptDetectionService: receiptDetectionService,
	}
}

func (h *ReceiptDetection) DetectReceipt(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:    http.StatusUnprocessableEntity,
			Message: fmt.Sprintf("Failed to read file from request: %v", err),
		}))
		return
	}

	data, err := h.receiptDetectionService.DetectAndStoreReceipt(ctx.Request.Context(), file, fileHeader)
	if err != nil {
		ctx.Error(err)
		return
	}

	hHelper.ResponseOK(ctx, data)
}

func (h *ReceiptDetection) GetByResultId(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	resultId := ctx.Param("result_id")
	if resultId == "" {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusBadRequest,
			ResponseMessage: "result_id must be provided",
		}))
		return
	}

	data, err := h.receiptDetectionService.GetResult(ctx.Request.Context(), resultId)
	if err != nil {
		ctx.Error(err)
		return
	}

	hHelper.ResponseOK(ctx, data)
}
