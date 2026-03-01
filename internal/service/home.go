package service

import (
	"context"

	"my_mdb/internal/config"
	"my_mdb/internal/domain"

	"github.com/sirupsen/logrus"
)

type HomeServiceDeps struct {
	Movies  MoviesRepo
	Ratings RatingsRepo
	Tags    TagsRepo

	RecsRepo  RecommendationsRepo
	RecClient RecClient

	Cfg config.Config
	Log *logrus.Logger
}

type HomeService struct {
	movies   MoviesRepo
	ratings  RatingsRepo
	tags     TagsRepo
	recsRepo RecommendationsRepo
	rec      RecClient
	cfg      config.Config
	log      *logrus.Logger
}

func NewHomeService(d HomeServiceDeps) *HomeService {
	return &HomeService{
		movies:   d.Movies,
		ratings:  d.Ratings,
		tags:     d.Tags,
		recsRepo: d.RecsRepo,
		rec:      d.RecClient,
		cfg:      d.Cfg,
		log:      d.Log,
	}
}

type HomePage struct {
	ForYou     []domain.MovieCard
	Top200Pick []domain.MovieCard
	Genres     []string
	Changing   domain.ChangingBlock
}

func (s *HomeService) BuildHome(ctx context.Context, userID int) (*HomePage, error) {
	forYou, err := s.GetForYou(ctx, userID, 5)
	if err != nil {
		return nil, err
	}

	topPickMovies, err := s.movies.RandomFromTop(ctx, 200, 5)
	if err != nil {
		return nil, err
	}
	topPickCards := toCards(topPickMovies)

	genres, err := s.movies.ListGenres(ctx)
	if err != nil {
		return nil, err
	}

	changing, err := s.pickChangingBlockDaily(ctx, userID, 10)
	if err != nil {
		return nil, err
	}

	return &HomePage{
		ForYou:     forYou,
		Top200Pick: topPickCards,
		Genres:     genres,
		Changing:   changing,
	}, nil
}

func (s *HomeService) GetForYou(ctx context.Context, userID int, limit int) ([]domain.MovieCard, error) {
	if limit <= 0 {
		limit = 20
	}

	items, err := s.getOrFetchRecs(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []domain.MovieCard{}, nil
	}

	ids := make([]int, 0, len(items))
	scoreByID := make(map[int]float32, len(items))
	for _, it := range items {
		ids = append(ids, it.MovieID)
		scoreByID[it.MovieID] = it.Score
	}

	mvs, err := s.movies.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		score := scoreByID[mvs[i].ID]
		out = append(out, toCard(mvs[i], nil, nil, &score))
	}

	return out, nil
}

const userExcludeLimit = 5000

func (s *HomeService) getOrFetchRecs(ctx context.Context, userID int, limit int) ([]domain.RecommendationItem, error) {
	if userID <= 0 {
		return []domain.RecommendationItem{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	items, err := s.recsRepo.GetByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}

	if s.rec == nil {
		return []domain.RecommendationItem{}, nil
	}

	var excludeIDs []int
	excludeIDs, err = s.ratings.ListUserRatedMovieIDs(ctx, userID, userExcludeLimit)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Warn("failed to load user exclude ids")
		excludeIDs = nil
	}

	recCtx, cancel := context.WithTimeout(ctx, s.cfg.RecTimeout)
	defer cancel()

	items, err = s.rec.Recommend(recCtx, userID, limit, excludeIDs)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Warn("rec service recommend failed")
		return []domain.RecommendationItem{}, nil
	}

	if len(items) > 0 {
		if err := s.recsRepo.ReplaceForUser(ctx, userID, items); err != nil {
			s.log.WithError(err).WithField("user_id", userID).Warn("failed to save recs to db")
		}
	}

	return items, nil
}

func toCard(m domain.Movie, posterURL *string, userRate *float32, recScore *float32) domain.MovieCard {
	return domain.MovieCard{
		ID:        m.ID,
		Title:     m.Title,
		Year:      m.Year,
		Genres:    m.Genres,
		PosterURL: posterURL,
		UserRate:  userRate,
		RecScore:  recScore,
	}
}

func toCards(movies []domain.Movie) []domain.MovieCard {
	out := make([]domain.MovieCard, 0, len(movies))
	for i := range movies {
		out = append(out, toCard(movies[i], nil, nil, nil))
	}
	return out
}
