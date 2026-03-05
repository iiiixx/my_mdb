package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"my_mdb/internal/config"
	"my_mdb/internal/db"
	"my_mdb/internal/jobs"
	"my_mdb/internal/logger"
	"my_mdb/internal/omdb"
	"my_mdb/internal/repo"
	"my_mdb/internal/service"
	grpcclient "my_mdb/internal/transport/grpc"
	httptransport "my_mdb/internal/transport/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	_ = godotenv.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := config.Load()

	log := logger.New()
	log.Info("starting application")

	database, err := db.New(ctx, cfg.DBURL, log)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize database")
	}
	runMigrations(cfg, log)
	defer database.Close()

	usersRepo := repo.NewUsersRepo(database.Pool)
	moviesRepo := repo.NewMoviesRepo(database.Pool)
	ratingsRepo := repo.NewRatingsRepo(database.Pool)
	postersRepo := repo.NewPostersRepo(database.Pool)
	detailsRepo := repo.NewMovieDetailsRepo(database.Pool)
	recsRepo := repo.NewRecommendationsRepo(database.Pool)
	similarRepo := repo.NewSimilarityRepo(database.Pool)
	tagsRepo := repo.NewTagsRepo(database.Pool)

	var omdbClient service.OMDbClient
	if cfg.OMDbAPIKey != "" {
		c, err := omdb.New(cfg.OMDbAPIKey)
		if err != nil {
			log.WithError(err).Fatal("failed to init omdb client")
		}
		omdbClient = c
		log.Info("OMDb client initialized")
	} else {
		log.Warn("OMDb API key not set, external movie details disabled")
	}

	var recClient service.RecClient

	if cfg.RecGRPCAddr != "" {
		c, err := grpcclient.Dial(cfg.RecGRPCAddr)
		if err != nil {
			log.WithError(err).Warn("failed to init rec grpc client")
		} else {
			recClient = c
			defer c.Close()
			log.Info("rec grpc client initialized")
		}
	}

	jobs.StartWeightedScoreCron(ctx, log, moviesRepo, 250, 6*time.Hour)
	//jobs.StartOMDbWarmupCron(ctx, log, cfg, moviesRepo, postersRepo, detailsRepo, omdbClient, 900, 24*time.Hour, 2)

	app := service.NewApp(service.Deps{
		Users:   usersRepo,
		Movies:  moviesRepo,
		Ratings: ratingsRepo,
		Posters: postersRepo,
		Details: detailsRepo,
		Recs:    recsRepo,
		Similar: similarRepo,
		Tags:    tagsRepo,

		OMDb: omdbClient,
		Rec:  recClient,

		Cfg: cfg,
		Log: log,
	})

	router := httptransport.NewRouter(app, cfg.HTTPAddr)

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.WithField("addr", cfg.HTTPAddr).Info("http server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("http server failed")
		}
	}()

	waitForShutdown(ctx, server, log)
}

func runMigrations(cfg config.Config, log *logrus.Logger) {
	m, err := migrate.New(
		"file://internal/migrations",
		cfg.DBURL,
	)
	if err != nil {
		log.WithError(err).Fatal("failed to init migrate")
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.WithError(err).Fatal("failed to run migrations")
	}

	log.Info("migrations applied")
}

func waitForShutdown(ctx context.Context, server *http.Server, log *logrus.Logger) {
	log.Info("application is running")

	<-ctx.Done()

	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("graceful shutdown failed")
	} else {
		log.Info("http server stopped gracefully")
	}

	log.Info("application stopped")
}
