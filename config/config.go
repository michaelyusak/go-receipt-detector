package config

import (
	"os"

	"github.com/michaelyusak/go-helper/config"
	hEntity "github.com/michaelyusak/go-helper/entity"
)

type OcrEngineConfig struct {
	BaseUrl string `json:"base_url"`
}

type CorsConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
}

type AppConfig struct {
	Port           string           `json:"port"`
	LogLevel       string           `json:"log_level"`
	GracefulPeriod hEntity.Duration `json:"graceful_period"`
	Cors           CorsConfig       `json:"cors"`
	OcrEngine      OcrEngineConfig  `json:"ocr_engine"`
}

func Init() (AppConfig, error) {
	configFilePath := os.Getenv("GO_RECEIPT_DETECTOR_CONFIG")

	return config.InitFromJson[AppConfig](configFilePath)
}
