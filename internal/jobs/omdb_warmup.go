package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"my_mdb/internal/config"
	"my_mdb/internal/domain"
	"my_mdb/internal/service"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type MoviesMissingMetaRepo interface {
	ListTopMissingMeta(ctx context.Context, limit int) ([]domain.Movie, error)
}

type PostersRepo interface {
	UpsertPoster(ctx context.Context, movieID int, posterURL string) error
}

type MovieDetailsRepo interface {
	UpsertMovieDetails(ctx context.Context, movieID int, payload []byte) error
}

var (
	omdbMu        sync.Mutex
	omdbStopUntil time.Time
)

var errOMDbRateLimited = errors.New("omdb rate limited")

func StartOMDbWarmupCron(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.Config,
	movies MoviesMissingMetaRepo,
	posters PostersRepo,
	details MovieDetailsRepo,
	omdb service.OMDbClient,
	batchSize int,
	every time.Duration,
	workers int,
) {
	if omdb == nil {
		log.Warn("OMDb warmup disabled: client is nil")
		return
	}
	if batchSize <= 0 {
		batchSize = 50
	}
	if workers <= 0 {
		workers = 3
	}
	if every <= 0 {
		every = 12 * time.Hour
	}

	ticker := time.NewTicker(every)
	go func() {
		defer ticker.Stop()

		runOnce(ctx, log, cfg, movies, posters, details, omdb, batchSize, workers)

		for {
			select {
			case <-ctx.Done():
				log.Info("OMDb warmup stopped")
				return
			case <-ticker.C:
				runOnce(ctx, log, cfg, movies, posters, details, omdb, batchSize, workers)
			}
		}
	}()
}

func runOnce(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.Config,
	movies MoviesMissingMetaRepo,
	posters PostersRepo,
	details MovieDetailsRepo,
	omdb service.OMDbClient,
	batchSize int,
	workers int,
) {
	omdbMu.Lock()
	stopUntil := omdbStopUntil
	omdbMu.Unlock()

	if !stopUntil.IsZero() && time.Now().Before(stopUntil) {
		log.WithField("until", stopUntil).Warn("omdb warmup: skipped (rate limited)")
		return
	}

	start := time.Now()

	cands, err := movies.ListTopMissingMeta(ctx, batchSize)
	if err != nil {
		log.WithError(err).Warn("omdb warmup: failed to load candidates")
		return
	}
	if len(cands) == 0 {
		log.Info("omdb warmup: nothing to do")
		return
	}

	sem := make(chan struct{}, workers)
	g, gctx := errgroup.WithContext(ctx)

	for i := range cands {
		mv := cands[i]

		if mv.IMDbID == nil || *mv.IMDbID <= 0 {
			continue
		}

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			imdb := fmt.Sprintf("tt%07d", *mv.IMDbID)

			reqCtx, cancel := context.WithTimeout(gctx, cfg.OMDbTimeoutForJob)
			defer cancel()

			payload, poster, err := omdb.FetchMovie(reqCtx, imdb)

			if payloadSaysRateLimited(payload) || errLooksRateLimited(err) {
				return errOMDbRateLimited
			}

			if err != nil {
				log.WithError(err).WithField("imdb_id", imdb).Info("omdb warmup: fetch failed")
				return nil
			}

			if len(payload) > 0 && details != nil {
				_ = details.UpsertMovieDetails(gctx, mv.ID, payload)
			}
			if poster != nil && *poster != "" && *poster != "N/A" && posters != nil {
				_ = posters.UpsertPoster(gctx, mv.ID, *poster)
			}

			return nil
		})
	}

	err = g.Wait()
	if errors.Is(err, errOMDbRateLimited) {
		until := untilTomorrowMidnight(time.Local).Add(1 * time.Minute)

		omdbMu.Lock()
		omdbStopUntil = until
		omdbMu.Unlock()

		log.WithFields(logrus.Fields{
			"until":     until,
			"batchSize": batchSize,
			"workers":   workers,
		}).Warn("omdb warmup: rate limit reached, pausing until tomorrow")
		return
	}

	log.WithFields(logrus.Fields{
		"took":      time.Since(start).String(),
		"cands":     len(cands),
		"batchSize": batchSize,
		"workers":   workers,
	}).Info("omdb warmup: cycle done")
}

func payloadSaysRateLimited(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}

	var v struct {
		Response string `json:"Response"`
		Error    string `json:"Error"`
	}
	if err := json.Unmarshal(payload, &v); err == nil {
		msg := strings.ToLower(v.Error)
		if strings.Contains(msg, "limit") && strings.Contains(msg, "reach") {
			return true
		}
		if strings.Contains(msg, "daily") && strings.Contains(msg, "limit") {
			return true
		}
	}

	s := strings.ToLower(string(payload))
	return strings.Contains(s, "limit") && strings.Contains(s, "reach")
}

func errLooksRateLimited(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "limit") && strings.Contains(s, "reach")
}

func untilTomorrowMidnight(loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	tomorrow := now.Add(24 * time.Hour)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
}
