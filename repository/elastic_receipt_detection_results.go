package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"receipt-detector/entity"

	"github.com/elastic/go-elasticsearch/v9"
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

func (r *receiptDetectionResultsElasticRepo) InsertOne(ctx context.Context, result []entity.OcrEngineItemDetail) (string, error) {
	res, err := r.client.Index(r.receiptDetectionResultsIndex).
		Request(entity.ReceiptDetectionDocument{
			Result: result,
		}).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("[repository][receiptDetectionResultsElasticRepo][InserOne][client.Index]: %w", err)
	}

	return res.Id_, nil
}

func (r *receiptDetectionResultsElasticRepo) GetByResultId(ctx context.Context, resultId string) ([]entity.OcrEngineItemDetail, error) {
	esRes, err := r.client.Get(r.receiptDetectionResultsIndex, resultId).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("[repository][receiptDetectionResultsElasticRepo][GetByResultId][client.Get]: %w [result_id: %s]", err, resultId)
	}

	if !esRes.Found {
		return nil, nil
	}

	var res entity.ReceiptDetectionDocument

	err = json.Unmarshal([]byte(esRes.Source_), &res)
	if err != nil {
		return nil, fmt.Errorf("[repository][receiptDetectionResultsElasticRepo][GetByResultId][json.Unmarshal]: %w [result_id: %s]", err, resultId)
	}

	return res.Result, nil
}
