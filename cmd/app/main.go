package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"my_mdb/internal/config"
	"my_mdb/internal/db"
	"my_mdb/internal/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	log := logger.New()
	log.Info("starting application")

	database, err := db.New(ctx, cfg.DBURL, log)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize database")
	}
	defer database.Close()

	waitForShutdown(log)
}

func waitForShutdown(log *logrus.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Info("application is running")

	<-stop

	log.Info("shutdown signal received")

	// Даём немного времени на завершение операций
	time.Sleep(2 * time.Second)

	log.Info("application stopped")
}
