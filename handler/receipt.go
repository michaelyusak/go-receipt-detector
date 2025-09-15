package handler

import (
	"fmt"
	"net/http"
	"receipt-detector/service"

	"github.com/gin-gonic/gin"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

type ReceiptHandler struct {
	receiptDetection service.ReceiptDetection
}

func NewReceipHandler(receiptDetection service.ReceiptDetection) *ReceiptHandler {
	return &ReceiptHandler{
		receiptDetection: receiptDetection,
	}
}

func (h *ReceiptHandler) DetectReceipt(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.Error(hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:    http.StatusUnprocessableEntity,
			Message: fmt.Sprintf("Failed to read file from request: %v", err),
		}))
		return
	}

	data, err := h.receiptDetection.DetectAndStoreReceipt(ctx.Request.Context(), file, fileHeader)
	if err != nil {
		ctx.Error(err)
		return
	}

	hHelper.ResponseOK(ctx, data)
}
