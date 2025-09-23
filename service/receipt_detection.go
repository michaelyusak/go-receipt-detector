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

	logTag     string
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

		logTag: "[service][receiptDetection]",
	}
}

func (s *receiptDetection) DetectAndStoreReceipt(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (*entity.ReceiptDetectionResult, error) {
	logTag := s.logTag + "[DetectAndStoreReceipt]"

	if fileHeader.Size > int64(s.maxFileSizeMb)*1024*1024 {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusRequestEntityTooLarge,
			Message:         fmt.Sprintf("%s File size too large", logTag),
			ResponseMessage: "File size too large",
		})
	}

	fileTypeOk, contentType, err := hHelper.FileTypeAllowed(fileHeader, s.allowedFileType)
	if err != nil {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusUnprocessableEntity,
			Message:         fmt.Sprintf("%s[hHelper.FileTypeAllowed] Failed to detect file type: %v", logTag, err),
			ResponseMessage: "Corrupted or invalid file",
		})
	}
	if !fileTypeOk {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusBadRequest,
			Message:         fmt.Sprintf("%s File type not allowed: %s", logTag, contentType),
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
				Message: fmt.Sprintf("%s[receiptImagesRepo.StoreOne] Failed to store image: %v", logTag, err),
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
				Message: fmt.Sprintf("%s[ocrEngine.DetectReceipt] Failed detect receipt: %v", logTag, err),
			})
			return
		}

		id, err := s.receiptDetectionResultsRepo.InsertOne(ctx, details)
		if err != nil {
			errCh <- hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[receiptDetectionResultsRepo.InserOne] Failed to record ocr result: %v", logTag, err),
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
			}).Errorf("%s[receiptDetectionHistoriesRepo.InsertOne] Failed to insert reciept detection history", logTag)
		}

		imageUrl, err := s.receiptImagesRepo.GetImageUrl(c, fileName)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"result_id": resultId,
				"error":     err,
			}).Warnf("%s[receiptImagesRepo.GetImageUrl] Failed to get image url", logTag)
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
			}).Warnf("%s[cacheRepo.SetReceiptDetectionResult] Failed to cache result", logTag)
		}
	}(fileName, resultId, itemDetails)

	return &entity.ReceiptDetectionResult{
		ResultId: resultId,
		Result:   itemDetails,
	}, nil
}

func (s *receiptDetection) GetResult(ctx context.Context, resultId string) (*entity.ReceiptDetectionResult, error) {
	logTag := s.logTag + "[GetResult]"

	history, err := s.receiptDetectionHistoriesRepo.GetByResultId(ctx, resultId)
	if err != nil {
		return nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptDetectionHistoriesRepo.GetByResultId] Failed to get history: %v [result_id: %s]", logTag, err, resultId),
		})
	}
	if history == nil {
		return nil, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s[NilHistory] History not found [result_id: %s]", logTag, resultId),
		})
	}

	if history.ResultId == resultId && history.RevisionId != "" {
		logrus.Infof("%s[RevisionExists] [requested_result_id: %s][revision_result_id: %s]", logTag, history.ResultId, history.RevisionId)
		resultId = history.RevisionId
	}

	cachedResult, err := s.cacheRepo.GetReceiptDetectionResult(ctx, resultId)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"result_id": resultId,
			"error":     err,
		}).Errorf("%s[cacheRepo.GetReceiptDetectionResult] Failed to get data", logTag)
	}
	if cachedResult != nil {
		logrus.WithFields(logrus.Fields{
			"result_id": resultId,
		}).Infof("%s[CacheFound]", logTag)
		return cachedResult, nil
	}

	result, err := s.receiptDetectionResultsRepo.GetByResultId(ctx, resultId)
	if err != nil {
		return nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptDetectionResultsRepo.GetByResultId] Failed to get result: %v [result_id: %s]", logTag, err, resultId),
		})
	}

	imageUrl, err := s.receiptImagesRepo.GetImageUrl(ctx, history.ImagePath)
	if err != nil {
		return nil, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptImagesRepo.GetImageUrl] Failed to get image url: %v [result_id: %s]", logTag, err, resultId),
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
			}).Errorf("%s[cacheRepo.SetReceiptDetectionResult] Failed to cache result", logTag)
		}
	}()

	return &detectionResult, nil
}
