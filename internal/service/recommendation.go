package service

import (
	"context"
	"my_mdb/internal/domain"
)

func (s *MoviesService) GetForYouMovies(ctx context.Context, userID int, limit int) ([]domain.Movie, error) {
	if limit <= 0 {
		limit = 6
	}

	fallback := func() ([]domain.Movie, error) {
		return s.movies.NewReleasesWithGoodRating(ctx, limit, 3.8, 0)
	}

	cnt, err := s.ratings.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if cnt == 0 || s.recClient == nil {
		return fallback()
	}

	forYouMovies, err := s.buildForYou(ctx, userID, limit)
	if err != nil {
		s.log.WithError(err).Warn("recs failed, fallback")
		return fallback()
	}

	if len(forYouMovies) == 0 {
		s.log.WithFields(map[string]any{
			"user_id": userID,
			"cnt":     cnt,
		}).Info("recs empty (cold-start until next model refresh), fallback")
		return fallback()
	}

	return forYouMovies, nil
}

func (s *MoviesService) buildForYou(ctx context.Context, userID int, limit int) ([]domain.Movie, error) {
	if limit <= 0 {
		limit = 20
	}

	items, err := s.getOrFetchRecs(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []domain.Movie{}, nil
	}

	ids := make([]int, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.MovieID)
	}

	mvs, err := s.movies.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return mvs, nil
}

const userExcludeLimit = 5000

func (s *MoviesService) getOrFetchRecs(ctx context.Context, userID int, limit int) ([]domain.RecommendationItem, error) {
	if userID <= 0 {
		return []domain.RecommendationItem{}, nil
	}
	if limit <= 0 {
		limit = 20
	}
	/*
		items, err := s.recsRepo.GetByUser(ctx, userID, limit)
		if err != nil {
			return nil, err
		}
		if len(items) > 0 && len(items) >= limit {
			return items, nil
		}
	*/
	if s.recClient == nil {
		return []domain.RecommendationItem{}, nil
	}

	var excludeIDs []int
	excludeIDs, err := s.ratings.ListUserRatedMovieIDs(ctx, userID, userExcludeLimit)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Warn("failed to load user exclude ids")
		excludeIDs = nil
	}

	recCtx, cancel := context.WithTimeout(ctx, s.cfg.RecTimeout)
	defer cancel()

	items, err := s.recClient.Recommend(recCtx, userID, limit, excludeIDs)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Warn("rec service recommend failed")
		return []domain.RecommendationItem{}, nil
	}

	/*
		if len(items) > 0 {
			if err := s.recsRepo.ReplaceForUser(ctx, userID, items); err != nil {
				s.log.WithError(err).WithField("user_id", userID).Warn("failed to save recs to db")
			}
		}
	*/

	return items, nil
}
