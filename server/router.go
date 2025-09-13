package server

import (
	"receipt-detector/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	hHandler "github.com/michaelyusak/go-helper/handler"
	hMiddleware "github.com/michaelyusak/go-helper/middleware"
	"github.com/sirupsen/logrus"
)

type routerOpts struct {
	common *hHandler.CommonHandler
}

func newRouter(config *config.AppConfig) *gin.Engine {
	commonHandler := &hHandler.CommonHandler{}

	return createRouter(routerOpts{
		common: commonHandler,
	},
		config.Cors.AllowedOrigins)
}

func createRouter(opts routerOpts, allowedOrigins []string) *gin.Engine {
	router := gin.New()

	corsConfig := cors.DefaultConfig()

	router.ContextWithFallback = true

	router.Use(
		hMiddleware.Logger(logrus.New()),
		hMiddleware.RequestIdHandlerMiddleware,
		hMiddleware.ErrorHandlerMiddleware,
		gin.Recovery(),
	)

	corsRouting(router, corsConfig, allowedOrigins)
	commonRouting(router, opts.common)

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
	router.GET("/ping", handler.Ping)
	router.NoRoute(handler.NoRoute)
}
