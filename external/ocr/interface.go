package ocr

import (
	"context"
	"mime/multipart"
	"receipt-detector/entity"
)

type OcrEngine interface {
	DetectReceipt(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) ([]entity.OcrEngineItemDetail, error)
}
