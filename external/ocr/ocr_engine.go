package ocr

import (
	"context"
	"fmt"
	"mime/multipart"
	"receipt-detector/entity"

	"github.com/go-resty/resty/v2"
)

type ocrEngineRestClient struct {
	client  *resty.Client
	baseUrl string

	detectReceiptPath string
	fileParam         string

	logHeading string
}

func NewOcEngineRestClient(baseUrl string) *ocrEngineRestClient {
	return &ocrEngineRestClient{
		client:  resty.New(),
		baseUrl: baseUrl,

		detectReceiptPath: "/detect",
		fileParam:         "file",

		logHeading: "[external][ocr][ocrEngineRestClient]",
	}
}

func (r *ocrEngineRestClient) DetectReceipt(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) ([]entity.OcrEngineItemDetail, error) {
	logHeading := r.logHeading + "[DetectReceipt]"

	ocrResponse := &entity.OcrEngineResponse[[]entity.OcrEngineItemDetail]{}

	resp, err := r.client.R().
		SetFileReader(r.fileParam, fileHeader.Filename, file).
		SetResult(ocrResponse).
		SetError(ocrResponse).
		Post(r.baseUrl + r.detectReceiptPath)
	if err != nil {
		return nil, fmt.Errorf("%s[client.R()] %w", logHeading, err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("%s[resp.IsError] Error Response [status_code: %v][resp: %s]", logHeading, resp.StatusCode(), string(resp.Body()))
	}

	return ocrResponse.Data, nil
}
