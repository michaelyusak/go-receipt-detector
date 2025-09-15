package repository

import (
	"context"
	"fmt"
	"receipt-detector/entity"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/sirupsen/logrus"
)

type receiptDetectionResultsElasticRepo struct {
	client *elasticsearch.TypedClient

	receiptDetectionResultsIndex string
}

func NewReceiptDetectionResultsElasticRepo(client *elasticsearch.TypedClient, receiptDetectionResultsIndex string) *receiptDetectionResultsElasticRepo {
	return &receiptDetectionResultsElasticRepo{
		client:                       client,
		receiptDetectionResultsIndex: receiptDetectionResultsIndex,
	}
}

func (r *receiptDetectionResultsElasticRepo) InserOne(ctx context.Context, result []entity.OcrEngineItemDetail) (string, error) {
	logrus.WithFields(logrus.Fields{
		"result": fmt.Sprintf("%+v", entity.ReceiptDetectionDocument{Result: result}),
	}).Infof("debug")

	res, err := r.client.Index(r.receiptDetectionResultsIndex).
		Request(entity.ReceiptDetectionDocument{
			Result: result,
		}).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("[repository][receiptDetectionResultsElasticRepo][InserOne][client.Index]: %w", err)
	}

	return res.Id_, nil
}
