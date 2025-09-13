package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"receipt-detector/config"
	"receipt-detector/log"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func Init() {
	config, err := config.Init()
	if err != nil {
		panic(err)
	}

	log.Init(config.LogLevel)

	router := newRouter(&config)

	srv := http.Server{
		Handler: router,
		Addr:    config.Port,
	}

	go func() {
		logrus.Infof("Sever running on port %s", config.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 10)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logrus.Infof("Server shutting down in %s ...", time.Duration(config.GracefulPeriod).String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GracefulPeriod))
	defer cancel()

	<-ctx.Done()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server shut down: %s", err.Error())
	}

	logrus.Info("Server shut down")
}
