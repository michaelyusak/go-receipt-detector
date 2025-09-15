package adaptor

import (
	"fmt"
	"receipt-detector/config"

	"github.com/elastic/go-elasticsearch/v9"
)

func ConnectElastic(config config.ElasticSearchConfig) (*elasticsearch.TypedClient, error) {
	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
		CACert:    config.CaCert,
	})
	if err != nil {
		return nil, fmt.Errorf("[adaptor][ConnectElastic][elasticsearch.NewTypedClient] Failed to create new typed client: %w", err)
	}

	return es, nil
}
