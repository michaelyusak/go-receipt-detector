package config

import (
	"fmt"
	"os"

	hConfig "github.com/michaelyusak/go-helper/config"
	hEntity "github.com/michaelyusak/go-helper/entity"
)

type OcrEngineConfig struct {
	BaseUrl string `json:"base_url"`
}

type OcrConfig struct {
	OcrEngine       OcrEngineConfig `json:"ocr_engine"`
	MaxFileSize     float64         `json:"max_file_size_mb"`
	AllowedFileType map[string]bool `json:"allowed_file_type"`
}

type CorsConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
}

type ElasticSearchIndicesConfig struct {
	ReceiptDetectionResults string `json:"receipt_detection_results"`
}

type ElasticSearchConfig struct {
	Addresses  []string                   `json:"addresses"`
	Username   string                     `json:"username"`
	Password   string                     `json:"password"`
	Indices    ElasticSearchIndicesConfig `json:"indices"`
	CaCertPath string                     `json:"ca_cert_path"`
	CaCert     []byte                     `json:"-"`
}

type CacheDurationConfig struct {
	ReceiptDetectionResult hEntity.Duration `json:"receipt_detection_result"`
	Receipt                hEntity.Duration `json:"receipt"`
	ReceiptItems           hEntity.Duration `json:"receipt_items"`
}

type CacheConfig struct {
	Duration CacheDurationConfig `json:"duration"`
}

type LocalStorageConfig struct {
	Directory          string `json:"directory"`
	EnableStaticServer bool   `json:"enable_static_server"`
	ServerHost         string `json:"server_host"`
	ServerStaticPath   string `json:"server_static_path"`
}

type StorageConfig struct {
	Local LocalStorageConfig `json:"local"`
}

type AppConfig struct {
	Port           string              `json:"port"`
	LogLevel       string              `json:"log_level"`
	GracefulPeriod hEntity.Duration    `json:"graceful_period"`
	Cors           CorsConfig          `json:"cors"`
	Db             hEntity.DBConfig    `json:"db"`
	Elasticsearch  ElasticSearchConfig `json:"elasticsearch"`
	Redis          hEntity.RedisConfig `json:"redis"`
	Cache          CacheConfig         `json:"cache"`
	Storage        StorageConfig       `json:"storage"`
	Ocr            OcrConfig           `json:"ocr"`
}

func Init() (AppConfig, error) {
	configFilePath := os.Getenv("GO_RECEIPT_DETECTOR_CONFIG")

	var conf AppConfig

	conf, err := hConfig.InitFromJson[AppConfig](configFilePath)
	if err != nil {
		return conf, fmt.Errorf("[config][Init][hConfig.InitFromJson] Failed to init config from json: %w", err)
	}

	esCert, err := os.ReadFile(conf.Elasticsearch.CaCertPath)
	if err != nil {
		return conf, fmt.Errorf("[config][Init][os.ReadFile] Failed to read es cert: %w", err)
	}

	conf.Elasticsearch.CaCert = esCert

	return conf, nil
}
