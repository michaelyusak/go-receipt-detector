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

type receiptDetection struct {
	ocrEngine                     ocr.OcrEngine
	receiptDetectionHistoriesRepo repository.ReceiptDetectionHistories
	receiptDetectionResultsRepo   repository.ReceiptDetectionResults
	receiptImagesRepo             repository.ReceiptImages
	cacheRepo                     repository.Cache

	maxFileSizeMb   float64
	allowedFileType map[string]bool

	logHeading     string
	allowedTypeStr string
}

type ReceiptDetectionResultsOpts struct {
	OcrEngine                     ocr.OcrEngine
	ReceiptDetectionHistoriesRepo repository.ReceiptDetectionHistories
	ReceiptDetectionResultsRepo   repository.ReceiptDetectionResults
	ReceiptImagesRepo             repository.ReceiptImages
	CacheRepo                     repository.Cache
	MaxFileSizeMb                 float64
	AllowedFileType               map[string]bool
}

func NewReceiptDetectionService(opts ReceiptDetectionResultsOpts) *receiptDetection {
	var allowedFileTypes []string

	for k := range opts.AllowedFileType {
		allowedFileTypes = append(allowedFileTypes, k)
	}

	return &receiptDetection{
		ocrEngine:                     opts.OcrEngine,
		receiptDetectionHistoriesRepo: opts.ReceiptDetectionHistoriesRepo,
		receiptDetectionResultsRepo:   opts.ReceiptDetectionResultsRepo,
		receiptImagesRepo:             opts.ReceiptImagesRepo,
		cacheRepo:                     opts.CacheRepo,

		maxFileSizeMb:   opts.MaxFileSizeMb,
		allowedFileType: opts.AllowedFileType,
		allowedTypeStr:  strings.Join(allowedFileTypes, ", "),

		logHeading: "[service][receiptDetection]",
	}
}

func (s *receiptDetection) DetectAndStoreReceipt(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*entity.ReceiptDetectionResult, error) {
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

	var wg sync.WaitGroup
	var fileName, resultId string
	var itemDetails []entity.OcrEngineItemDetail

	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		name, err := s.receiptImagesRepo.StoreOne(ctx, contentType, fileHeader)
		if err != nil {
			errCh <- hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[receiptImagesRepo.StoreOne] Failed to store image: %v", logHeading, err),
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

	go func(fileName, resultId string, itemDetails []entity.OcrEngineItemDetail) {
		c, cancel := context.WithTimeout(context.Background(), time.Duration(time.Minute))
		defer cancel()

		err = s.receiptDetectionHistoriesRepo.InsertOne(c, entity.ReceiptDetectionHistory{
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

		imageUrl, err := s.receiptImagesRepo.GetImageUrl(c, fileName)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"result_id": resultId,
				"error":     err,
			}).Warnf("%s[receiptImagesRepo.GetImageUrl] Failed to get image url", logHeading)
		}

		result := entity.ReceiptDetectionResult{
			ResultId: resultId,
			ImageUrl: imageUrl,
			Result:   itemDetails,
		}

		err = s.cacheRepo.SetReceiptDetectionResult(c, result)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"result_id": resultId,
				"error":     err,
			}).Warnf("%s[cacheRepo.SetReceiptDetectionResult] Failed to cache result", logHeading)
		}
	}(fileName, resultId, itemDetails)

	return &entity.ReceiptDetectionResult{
		ResultId: resultId,
		Result:   itemDetails,
	}, nil
}

func (s *receiptDetection) GetResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error) {
	logHeading := s.logHeading + "[GetResult]"

	history, err := s.receiptDetectionHistoriesRepo.GetByResultId(ctx, resultId)
	if err != nil {
		return nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptDetectionHistoriesRepo.GetByResultId] Failed to get history: %v [result_id: %s]", logHeading, err, resultId),
		})
	}
	if history == nil {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s[NilHistory] History not found [result_id: %s]", logHeading, resultId),
		})
	}

	if history.ResultId == resultId && history.RevisionId != "" {
		logrus.Infof("%s[RevisionExists] [requested_result_id: %s][revision_result_id: %s]", logHeading, history.ResultId, history.RevisionId)
		resultId = history.RevisionId
	}

	cachedResult, err := s.cacheRepo.GetReceiptDetectionResult(ctx, resultId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"result_id": resultId,
			"error":     err,
		}).Errorf("%s[cacheRepo.GetReceiptDetectionResult] Failed to get data", logHeading)
	}
	if cachedResult != nil {
		logrus.WithFields(logrus.Fields{
			"result_id": resultId,
		}).Infof("%s[CacheFound]", logHeading)
		return cachedResult, nil
	}

	result, err := s.receiptDetectionResultsRepo.GetByResultId(ctx, resultId)
	if err != nil {
		return nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptDetectionResultsRepo.GetByResultId] Failed to get result: %v [result_id: %s]", logHeading, err, resultId),
		})
	}

	imageUrl, err := s.receiptImagesRepo.GetImageUrl(ctx, history.ImagePath)
	if err != nil {
		return nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptImagesRepo.GetImageUrl] Failed to get image url: %v [result_id: %s]", logHeading, err, resultId),
		})
	}

	detectionResult := entity.ReceiptDetectionResult{
		ResultId: resultId,
		ImageUrl: imageUrl,
		Result:   result,
	}

	go func() {
		c, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := s.cacheRepo.SetReceiptDetectionResult(c, detectionResult)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"result_id": detectionResult.ResultId,
				"error":     err,
			}).Errorf("%s[cacheRepo.SetReceiptDetectionResult] Failed to cache result", logHeading)
		}
	}()

	return &detectionResult, nil
}
