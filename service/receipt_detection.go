package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"receipt-detector/entity"
	"receipt-detector/external/ocr"
	"receipt-detector/repository"
	"strings"
	"sync"
	"time"

	hApperror "github.com/michaelyusak/go-helper/apperror"
	hHelper "github.com/michaelyusak/go-helper/helper"
	"github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

type receiptDetection struct {
	ocrEngine                     ocr.OcrEngine
	receiptDetectionHistoriesRepo repository.ReceiptDetectionHistoriesRepository
	receiptDetectionResultsRepo   repository.ReceiptDetectionResultsRepository
	receiptImageRepo              repository.ReceiptImageRepository

	maxFileSizeMb   float64
	allowedFileType map[string]bool

	logHeading     string
	allowedTypeStr string
}

type ReceiptDetectionResultsOpts struct {
	OcrEngine                     ocr.OcrEngine
	ReceiptDetectionHistoriesRepo repository.ReceiptDetectionHistoriesRepository
	ReceiptDetectionResultsRepo   repository.ReceiptDetectionResultsRepository
	ReceiptImageRepo              repository.ReceiptImageRepository
	MaxFileSizeMb                 float64
	AllowedFileType               map[string]bool
}

func NewReceiptDetectionService(opts ReceiptDetectionResultsOpts) *receiptDetection {
	var allowedFileTypes []string

	for k, _ := range opts.AllowedFileType {
		allowedFileTypes = append(allowedFileTypes, k)
	}

	return &receiptDetection{
		ocrEngine:                     opts.OcrEngine,
		receiptDetectionHistoriesRepo: opts.ReceiptDetectionHistoriesRepo,
		receiptDetectionResultsRepo:   opts.ReceiptDetectionResultsRepo,
		receiptImageRepo:              opts.ReceiptImageRepo,

		maxFileSizeMb:   opts.MaxFileSizeMb,
		allowedFileType: opts.AllowedFileType,
		allowedTypeStr:  strings.Join(allowedFileTypes, ", "),

		logHeading: "[service][receiptDetection]",
	}
}

func (s *receiptDetection) DetectAndStoreReceipt(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) ([]entity.OcrEngineItemDetail, error) {
	logHeading := s.logHeading + "[DetectAndStoreReceipt]"

	if fileHeader.Size > int64(s.maxFileSizeMb)*1024*1024 {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusRequestEntityTooLarge,
			Message:         fmt.Sprintf("%s File size too large", logHeading),
			ResponseMessage: "File size too large",
		})
	}

	fileTypeOk, contentType, err := hHelper.FileTypeAllowed(fileHeader, s.allowedFileType)
	if err != nil {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusUnprocessableEntity,
			Message:         fmt.Sprintf("%s[hHelper.FileTypeAllowed] Failed to detect file type: %v", logHeading, err),
			ResponseMessage: "Corrupted or invalid file",
		})
	}
	if !fileTypeOk {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusBadRequest,
			Message:         fmt.Sprintf("%s File type not allowed: %s", logHeading, contentType),
			ResponseMessage: fmt.Sprintf("File type %s not allowed. List of allowed file types: %s", contentType, s.allowedTypeStr),
		})
	}

	var fileName, resultId string
	var itemDetails []entity.OcrEngineItemDetail

	errCh := make(chan error, 2)

	wg.Add(1)

	go func() {
		defer wg.Done()

		name, err := s.receiptImageRepo.StoreOne(ctx, contentType, fileHeader)
		if err != nil {
			errCh <- hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[receiptImageRepo.StoreOne] Failed to store image: %v", logHeading, err),
			})
			return
		}

		fileName = name
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		details, err := s.ocrEngine.DetectReceipt(ctx, file, fileHeader)
		if err != nil {
			errCh <- hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[ocrEngine.DetectReceipt] Failed detect receipt: %v", logHeading, err),
			})
			return
		}

		id, err := s.receiptDetectionResultsRepo.InsertOne(ctx, details)
		if err != nil {
			errCh <- hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[receiptDetectionResultsRepo.InserOne] Failed to record ocr result: %v", logHeading, err),
			})
			return
		}

		resultId = id
		itemDetails = details
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	go func(fileName, resultId string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
		defer cancel()

		err = s.receiptDetectionHistoriesRepo.InsertOne(ctx, entity.ReceiptDetectionHistory{
			ImagePath: fileName,
			ResultId:  resultId,
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"filename":  fileName,
				"result_id": resultId,
				"error":     err,
			}).Errorf("%s[receiptDetectionHistoriesRepo.InsertOne] Failed to insert reciept detection history", logHeading)
		}
	}(fileName, resultId)

	return itemDetails, nil
}
