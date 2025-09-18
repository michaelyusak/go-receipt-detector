package server

import (
	"receipt-detector/adaptor"
	"receipt-detector/config"
	"receipt-detector/external/ocr"
	"receipt-detector/handler"
	"receipt-detector/repository"
	"receipt-detector/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	hHandler "github.com/michaelyusak/go-helper/handler"
	hMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

var (
	APP_HEALTHY = false
)

type routerOpts struct {
	common  *hHandler.CommonHandler
	receipt *handler.ReceiptHandler
	bill    *handler.BillHandler
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

	redis := adaptor.ConnectRedis(config.Redis)
	logrus.Info("Connected to redis")

	receiptDetectionHistoriesRepo := repository.NewReceiptDetectionHistoriesPostgres(db)
	receiptDetectionResultsRepo := repository.NewReceiptDetectionResultsElastic(es, config.Elasticsearch.Indices.ReceiptDetectionResults)
	receiptImagesRepo := repository.NewReceiptImageLocalStorage(config.Storage.Local.Directory, config.Storage.Local.ServerHost+config.Storage.Local.ServerStaticPath)
	cacheRepo := repository.NewCacheRedisRepo(repository.CacheRedisRepoOpt{
		Client:                              redis,
		ReceiptDetectionResultCacheDuration: time.Duration(config.Cache.Duration.ReceiptDetectionResult),
		BillCacheDuration:                   time.Duration(config.Cache.Duration.Bill),
		BillItemsCacheDuration:              time.Duration(config.Cache.Duration.BillItems),
	})
	billRepo := repository.NewBillPostgresRepo(db)
	billItemRepo := repository.NewBillItemsPostgresRepo(db)

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
	billService := service.NewBillService(service.BillOpts{
		BillRepo:                      billRepo,
		BillItemRepo:                  billItemRepo,
		ReceiptDetectionHistoriesRepo: receiptDetectionHistoriesRepo,
		ReceiptImagesRepo:             receiptImagesRepo,
		CacheRepo:                     cacheRepo,
	})

	commonHandler := hHandler.NewCommonHandler(&APP_HEALTHY)
	receiptHandler := handler.NewReceipHandler(receiptDetectionService)
	billHandler := handler.NewBillHandler(billService)

	return createRouter(routerOpts{
		common:  commonHandler,
		receipt: receiptHandler,
		bill:    billHandler,
	},
		config.Cors.AllowedOrigins,
		config.Storage.Local)
}

func createRouter(opts routerOpts, allowedOrigins []string, localStorageConfig config.LocalStorageConfig) *gin.Engine {
	router := gin.New()

	corsConfig := cors.DefaultConfig()

	router.ContextWithFallback = true

	router.Use(
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
	receiptRouting(router, opts.receipt)
	billRouting(router, opts.bill)

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

func receiptRouting(router *gin.Engine, handler *handler.ReceiptHandler) {
	receiptRouter := router.Group("/receipt")

	receiptRouter.POST("/detect", handler.DetectReceipt)
	receiptRouter.GET("/:result_id", handler.GetByResultId)
}

func billRouting(router *gin.Engine, handler *handler.BillHandler) {
	billRouter := router.Group("/bill")

	billRouter.POST("", handler.Create)
	billRouter.GET("/:bill_id", handler.GetById)
	billRouter.PATCH("/:bill_id", handler.UpdateBill)
}
