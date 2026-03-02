package jobs

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func StartWeightedScoreCron(
	ctx context.Context,
	log *logrus.Logger,
	moviesRepo interface {
		RefreshWeightedScore(context.Context, float64) error
	},
	m float64,
	every time.Duration,
) {
	if every <= 0 {
		every = 6 * time.Hour
	}

	go func() {
		runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		if err := moviesRepo.RefreshWeightedScore(runCtx, m); err != nil {
			log.WithError(err).Error("initial weighted_score refresh failed")
		} else {
			log.WithFields(logrus.Fields{"m": m}).Info("initial weighted_score refreshed")
		}
	}()

	go func() {
		t := time.NewTicker(every)
		defer t.Stop()

		log.WithFields(logrus.Fields{
			"every": every.String(),
			"m":     m,
		}).Info("weighted_score cron started")

		for {
			select {
			case <-ctx.Done():
				log.Info("weighted_score cron stopped")
				return
			case <-t.C:
				runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
				err := moviesRepo.RefreshWeightedScore(runCtx, m)
				cancel()

				if err != nil {
					log.WithError(err).Error("weighted_score refresh failed")
					continue
				}
				log.WithFields(logrus.Fields{"m": m}).Info("weighted_score refreshed")
			}
		}
	}()
}
