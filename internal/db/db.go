// internal/db/database.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Database struct {
	Pool   *pgxpool.Pool
	Logger *logrus.Logger
}

func New(ctx context.Context, dbURL string, logger *logrus.Logger) (*Database, error) {
	log := logger.WithFields(logrus.Fields{
		"layer":  "db",
		"method": "New",
	})
	log.Info("initializing database connection pool")

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.WithError(err).Error("failed to parse DB_URL")
		return nil, fmt.Errorf("db: parse config: %w", err)
	}

	cfg.MinConns = 2
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.WithError(err).Error("unable to create connection pool")
		return nil, fmt.Errorf("db: create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		log.WithError(err).Error("database ping failed")
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	stats := pool.Stat()
	log.WithFields(logrus.Fields{
		"total_conns": stats.TotalConns(),
		"idle_conns":  stats.IdleConns(),
		"acquired":    stats.AcquiredConns(),
	}).Info("database connection pool ready")

	return &Database{
		Pool:   pool,
		Logger: logger,
	}, nil
}

func (d *Database) Close() {
	log := d.Logger.WithFields(logrus.Fields{
		"layer":  "db",
		"method": "Close",
	})
	log.Info("closing database connection pool")

	start := time.Now()
	statsBefore := d.Pool.Stat()

	d.Pool.Close()

	log.WithFields(logrus.Fields{
		"connections_closed": statsBefore.TotalConns(),
		"duration_ms":        time.Since(start).Milliseconds(),
	}).Info("database connection pool closed")
}
