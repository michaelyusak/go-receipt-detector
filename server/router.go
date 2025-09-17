package server

import (
	"fmt"
	"receipt-detector/adaptor"
	"receipt-detector/config"
	"receipt-detector/external/ocr"
	"receipt-detector/handler"
	"receipt-detector/repository"
	"receipt-detector/service"

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
}

func newRouter(config *config.AppConfig) *gin.Engine {
	db, err := adaptor.ConnectPostgres(config.Db)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to db: %v", err))
	}

	es, err := adaptor.ConnectElastic(config.Elasticsearch)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to es: %v", err))
	}

	receiptDetectionHistoriesRepo := repository.NewReceiptDetectionHistoriesPostgresRepo(db)
	receiptDetectionResultsRepo := repository.NewReceiptDetectionResultsElasticRepo(es, config.Elasticsearch.Indices.ReceiptDetectionResults)
	receiptImageRepo := repository.NewReceiptImageLocalStorage(config.Storage.Local.Directory, config.Storage.Local.ServerHost+config.Storage.Local.ServerStaticPath)

	ocrEngine := ocr.NewOcEngineRestClient(config.Ocr.OcrEngine.BaseUrl)

	receiptDetectionService := service.NewReceiptDetectionService(service.ReceiptDetectionResultsOpts{
		OcrEngine:                     ocrEngine,
		ReceiptDetectionHistoriesRepo: receiptDetectionHistoriesRepo,
		ReceiptDetectionResultsRepo:   receiptDetectionResultsRepo,
		ReceiptImageRepo:              receiptImageRepo,
		MaxFileSizeMb:                 config.Ocr.MaxFileSize,
		AllowedFileType:               config.Ocr.AllowedFileType,
	})

	commonHandler := hHandler.NewCommonHandler(&APP_HEALTHY)
	receiptHandler := handler.NewReceipHandler(receiptDetectionService)

	return createRouter(routerOpts{
		common:  commonHandler,
		receipt: receiptHandler,
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
