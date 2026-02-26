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

	RecsRepo  RecommendationsRepo
	RecClient RecClient

	Cfg config.Config
	Log *logrus.Logger
}

type HomeService struct {
	movies   MoviesRepo
	ratings  RatingsRepo
	recsRepo RecommendationsRepo
	rec      RecClient
	cfg      config.Config
	log      *logrus.Logger
}

func NewHomeService(d HomeServiceDeps) *HomeService {
	return &HomeService{
		movies:   d.Movies,
		ratings:  d.Ratings,
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
	Changing   ChangingBlock
}

type ChangingBlock struct {
	Kind   string
	Title  string
	Movies []domain.MovieCard
}

func (s *HomeService) BuildHome(ctx context.Context, userID int) (*HomePage, error) {
	// 1) For you (5)
	forYou, err := s.GetForYou(ctx, userID, 5)
	if err != nil {
		return nil, err
	}

	// 2) random 5 from top-200
	topPickMovies, err := s.movies.RandomFromTop(ctx, 200, 5)
	if err != nil {
		return nil, err
	}
	topPickCards := make([]domain.MovieCard, 0, len(topPickMovies))
	for i := range topPickMovies {
		topPickCards = append(topPickCards, toCard(topPickMovies[i], nil, nil, nil))
	}

	// 3) genres
	genres, err := s.movies.ListGenres(ctx)
	if err != nil {
		return nil, err
	}

	// 4) changing block
	nineties, err := s.movies.MoviesByYearRange(ctx, 1990, 1999, 10)
	if err != nil {
		return nil, err
	}
	chCards := make([]domain.MovieCard, 0, len(nineties))
	for i := range nineties {
		chCards = append(chCards, toCard(nineties[i], nil, nil, nil))
	}

	return &HomePage{
		ForYou:     forYou,
		Top200Pick: topPickCards,
		Genres:     genres,
		Changing: ChangingBlock{
			Kind:   "nineties",
			Title:  "Фильмы 90-ых",
			Movies: chCards,
		},
	}, nil
}

func (s *HomeService) GetForYou(ctx context.Context, userID int, limit int) ([]domain.MovieCard, error) {
	if limit <= 0 {
		limit = 20
	}

	items, err := s.recsRepo.GetByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 && s.rec != nil {
		recCtx, cancel := context.WithTimeout(ctx, s.cfg.RecTimeout)
		defer cancel()

		items, err = s.rec.Recommend(recCtx, userID, limit)
		if err != nil {
			s.log.WithError(err).WithField("user_id", userID).Warn("rec service recommend failed")
			return []domain.MovieCard{}, nil
		}
		_ = s.recsRepo.ReplaceForUser(ctx, userID, items)
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
