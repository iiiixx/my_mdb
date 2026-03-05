package service

import (
	"context"
	"my_mdb/internal/domain"
)

func (s *MoviesService) getOrFetchSimilar(ctx context.Context, movieID int, limit int) ([]domain.SimilarItem, error) {
	if movieID <= 0 {
		s.log.WithField("movie_id", movieID).Warn("invalid movie_id")
		return []domain.SimilarItem{}, nil
	}
	if limit <= 0 {
		limit = defaultSimilarLimit
	}

	items, err := s.similar.GetSimilar(ctx, movieID, limit)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}

	if s.recClient == nil {
		s.log.WithField("movie_id", movieID).Warn("recClient is nil; skip rec-service call")
		return []domain.SimilarItem{}, nil
	}

	recCtx, cancel := context.WithTimeout(ctx, s.cfg.RecTimeout)
	defer cancel()

	items, err = s.recClient.SimilarMovies(recCtx, movieID, limit)
	if err != nil {
		s.log.WithError(err).WithField("movie_id", movieID).Warn("rec service similar failed")
		return []domain.SimilarItem{}, nil
	}

	if err := s.similar.InsertIfNotExists(ctx, movieID, items); err != nil {
		s.log.WithError(err).WithField("movie_id", movieID).Warn("failed to save similar to db")
	}

	return items, nil
}
