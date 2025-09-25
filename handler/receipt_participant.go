package handler

import (
	"receipt-detector/entity"
	"receipt-detector/service"
	"strconv"

	"github.com/gin-gonic/gin"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	hHelper "github.com/michaelyusak/go-helper/helper"
)

type ReceiptParticipant struct {
	receiptParticipantService service.ReceiptParticipant
}

func NewReceiptParticipant(receiptParticipantService service.ReceiptParticipant) *ReceiptParticipant {
	return &ReceiptParticipant{
		receiptParticipantService: receiptParticipantService,
	}
}

func (h *ReceiptParticipant) AddParticipants(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	var req entity.AddParticipantsRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

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

	err = h.receiptParticipantService.AddParticipants(ctx.Request.Context(), int64(receiptId), req.Participants)
	if err != nil {
		ctx.Error(err)
		return
	}

	hHelper.ResponseOK(ctx, nil)
}

func (h *ReceiptParticipant) GetParticipants(ctx *gin.Context) {
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

	participants, err := h.receiptParticipantService.GetByReceiptId(ctx, int64(receiptId))
	if err != nil {
		ctx.Error(err)
		return
	}

	hHelper.ResponseOK(ctx, entity.GetParticipantsResponse{
		Participants: participants,
	})
}

func (h *ReceiptParticipant) GetAllowedContactTypes(ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	allowedContactTypes := h.receiptParticipantService.GetAllowedContactTypes()

	hHelper.ResponseOK(ctx, entity.GetAllowedContactTypesResponse{
		AllowedContactTypes: allowedContactTypes,
	})
}
