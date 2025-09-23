package server

import (
	"receipt-detector/adaptor"
	"receipt-detector/config"
	"receipt-detector/external/ocr"
	"receipt-detector/handler"
	"receipt-detector/repository/elasticsearch"
	"receipt-detector/repository/localstorage"
	"receipt-detector/repository/postgres"
	"receipt-detector/repository/redis"
	"receipt-detector/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	hHandler "github.com/michaelyusak/go-helper/handler"
	hHelper "github.com/michaelyusak/go-helper/helper"
	hMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

var (
	APP_HEALTHY = false
)

type routerOpts struct {
	common           *hHandler.CommonHandler
	receiptDetection *handler.ReceiptDetection
	receipt          *handler.Receipt

	hash hHelper.HashHelper
}

func newRouter(config *config.AppConfig) *gin.Engine {
	db, err := adaptor.ConnectPostgres(config.Db)
	if err != nil {
		logrus.Panicf("Failed to connect to db: %v", err)
	}
	logrus.Info("Connected to postgres")

	es, err := adaptor.ConnectElastic(config.Elasticsearch)
	if err != nil {
		logrus.Panicf("Failed to connect to es: %v", err)
	}
	logrus.Info("Connected to elasticsearch")

	rds := adaptor.ConnectRedis(config.Redis)
	logrus.Info("Connected to redis")

	hashHelper := hHelper.NewHashHelper(config.Hash)

	receiptDetectionHistoriesRepo := postgres.NewReceiptDetectionHistories(db)
	receiptDetectionResultsRepo := elasticsearch.NewReceiptDetectionResults(es, config.Elasticsearch.Indices.ReceiptDetectionResults)
	receiptImagesRepo := localstorage.NewReceiptImages(config.Storage.Local.Directory, config.Storage.Local.ServerHost+config.Storage.Local.ServerStaticPath)
	cacheRepo := redis.NewCacheRedisRepo(redis.CacheRedisRepoOpt{
		Client:                              rds,
		ReceiptDetectionResultCacheDuration: time.Duration(config.Cache.Duration.ReceiptDetectionResult),
		ReceiptCacheDuration:                time.Duration(config.Cache.Duration.Receipt),
		ReceiptItemsCacheDuration:           time.Duration(config.Cache.Duration.ReceiptItems),
	})
	receiptsRepo := postgres.NewReceipts(db)
	receiptItemsRepo := postgres.NewReceiptItems(db)

	ocrEngine := ocr.NewOcEngineRestClient(config.Ocr.OcrEngine.BaseUrl)

	receiptDetectionService := service.NewReceiptDetectionService(service.ReceiptDetectionResultsOpts{
		OcrEngine:                     ocrEngine,
		ReceiptDetectionHistoriesRepo: receiptDetectionHistoriesRepo,
		ReceiptDetectionResultsRepo:   receiptDetectionResultsRepo,
		ReceiptImagesRepo:             receiptImagesRepo,
		MaxFileSizeMb:                 config.Ocr.MaxFileSize,
		AllowedFileType:               config.Ocr.AllowedFileType,
		CacheRepo:                     cacheRepo,
	})
	receiptService := service.NewBillService(service.ReceiptOpts{
		ReceiptsRepo:                  receiptsRepo,
		ReceiptItemsRepo:              receiptItemsRepo,
		ReceiptDetectionHistoriesRepo: receiptDetectionHistoriesRepo,
		ReceiptImagesRepo:             receiptImagesRepo,
		CacheRepo:                     cacheRepo,
	})

	commonHandler := hHandler.NewCommonHandler(&APP_HEALTHY)
	receiptDetectionHandler := handler.NewReceiptDetection(receiptDetectionService)
	receiptHandler := handler.NewReceipt(receiptService)

	return createRouter(routerOpts{
		common:           commonHandler,
		receiptDetection: receiptDetectionHandler,
		receipt:          receiptHandler,

		hash: hashHelper,
	},
		config.Cors.AllowedOrigins,
		config.Storage.Local)
}

func createRouter(opts routerOpts, allowedOrigins []string, localStorageConfig config.LocalStorageConfig) *gin.Engine {
	router := gin.New()

	corsConfig := cors.DefaultConfig()

	router.ContextWithFallback = true

	authMiddleware := hMiddleware.NewAuth(hMiddleware.AuthOpt{
		IsCheckDeviceId: true,
		Hash:            opts.hash,
	})

	router.Use(
		authMiddleware.Auth(),
		hMiddleware.Logger(logrus.New()),
		hMiddleware.RequestIdHandlerMiddleware,
		hMiddleware.ErrorHandlerMiddleware,
		gin.Recovery(),
	)

	if localStorageConfig.EnableStaticServer {
		staticRouting(router, localStorageConfig.ServerStaticPath, localStorageConfig.Directory)
	}

	corsRouting(router, corsConfig, allowedOrigins)
	commonRouting(router, opts.common)
	receiptDetectionRouting(router, opts.receiptDetection)
	receiptRouting(router, opts.receipt)

	return router
}

func corsRouting(router *gin.Engine, corsConfig cors.Config, allowedOrigins []string) {
	corsConfig.AllowOrigins = allowedOrigins
	corsConfig.AllowMethods = []string{"POST", "GET", "PUT", "PATCH", "DELETE"}
	corsConfig.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Accept", "User-Agent", "Cache-Control"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))
}

func commonRouting(router *gin.Engine, handler *hHandler.CommonHandler) {
	router.GET("/health", handler.Health)
	router.NoRoute(handler.NoRoute)
}

func staticRouting(router *gin.Engine, localStorageStaticPath, localStorageDirectory string) {
	router.Static(localStorageStaticPath, localStorageDirectory)
}

func receiptDetectionRouting(router *gin.Engine, handler *handler.ReceiptDetection) {
	receiptDetectionRouter := router.Group("/receipt/detect")

	receiptDetectionRouter.POST("", handler.DetectReceipt)
	receiptDetectionRouter.GET("/:result_id", handler.GetByResultId)
}

func receiptRouting(router *gin.Engine, handler *handler.Receipt) {
	receiptRouter := router.Group("/receipt")

	receiptRouter.POST("", handler.Create)
	receiptRouter.GET("/:receipt_id", handler.GetByReceiptId)
	receiptRouter.PATCH("/:receipt_id", handler.UpdateReceipt)
}
