package service

import (
	"context"
	"encoding/json"
	"fmt"

	"my_mdb/internal/config"
	"my_mdb/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type MoviesServiceDeps struct {
	Movies  MoviesRepo
	Ratings RatingsRepo
	Posters PostersRepo
	Details MovieDetailsRepo
	Similar SimilarityRepo
	Tags    TagsRepo

	OMDb      OMDbClient
	RecClient RecClient

	Cfg config.Config
	Log *logrus.Logger
}

type MoviesService struct {
	movies    MoviesRepo
	ratings   RatingsRepo
	posters   PostersRepo
	details   MovieDetailsRepo
	similar   SimilarityRepo
	tags      TagsRepo
	omdb      OMDbClient
	recClient RecClient
	cfg       config.Config
	log       *logrus.Logger
}

func NewMoviesService(d MoviesServiceDeps) *MoviesService {
	return &MoviesService{
		movies:    d.Movies,
		ratings:   d.Ratings,
		posters:   d.Posters,
		details:   d.Details,
		similar:   d.Similar,
		recClient: d.RecClient,
		tags:      d.Tags,
		omdb:      d.OMDb,
		cfg:       d.Cfg,
		log:       d.Log,
	}
}

func (s *MoviesService) Search(ctx context.Context, q string, limit int) ([]domain.MovieCard, error) {
	mvs, err := s.movies.SearchMovies(ctx, q, limit)
	if err != nil {
		return nil, err
	}

	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		out = append(out, toCard(mvs[i], nil, nil, nil))
	}
	return out, nil
}

func (s *MoviesService) Top200(ctx context.Context) ([]domain.MovieCard, error) {
	mvs, err := s.movies.TopMovies(ctx, 200)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		out = append(out, toCard(mvs[i], nil, nil, nil))
	}
	return out, nil
}

func (s *MoviesService) Genres(ctx context.Context) ([]string, error) {
	return s.movies.ListGenres(ctx)
}

func (s *MoviesService) ByGenre(ctx context.Context, genre string, limit int) ([]domain.MovieCard, error) {
	mvs, err := s.movies.ListMoviesByGenre(ctx, genre, limit)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		out = append(out, toCard(mvs[i], nil, nil, nil))
	}
	return out, nil
}

func (s *MoviesService) WatchedByUser(ctx context.Context, userID int, limit int) ([]domain.MovieCard, error) {
	ids, err := s.ratings.ListUserRatedMovieIDs(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	mvs, err := s.movies.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		r, err := s.ratings.GetUserRatingForMovie(ctx, userID, mvs[i].ID)
		if err != nil {
			return nil, err
		}
		out = append(out, toCard(mvs[i], nil, r, nil))
	}
	return out, nil
}

func (s *MoviesService) Rate(ctx context.Context, userID, movieID int, v float32) error {
	return s.ratings.UpsertRating(ctx, userID, movieID, v)
}

func (s *MoviesService) ByTagQuery(ctx context.Context, tagQuery string, limit int) ([]domain.MovieCard, error) {
	ids, err := s.tags.TopMoviesByTagQuery(ctx, tagQuery, limit)
	if err != nil {
		return nil, err
	}
	mvs, err := s.movies.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MovieCard, 0, len(mvs))
	for i := range mvs {
		out = append(out, toCard(mvs[i], nil, nil, nil))
	}
	return out, nil
}

func (s *MoviesService) GetMovieDetails(ctx context.Context, userID int, movieID int) (*domain.MovieDetailsResponse, error) {
	m, err := s.movies.GetMovieByID(ctx, movieID)
	if err != nil {
		return nil, err
	}

	userRate, err := s.ratings.GetUserRatingForMovie(ctx, userID, movieID)
	if err != nil {
		return nil, err
	}

	meta, _ := s.getMovieMeta(ctx, movieID, m.IMDbID, true)

	simItems, err := s.getOrFetchSimilar(ctx, movieID, defaultSimilarLimit)
	if err != nil {
		return nil, err
	}

	simIDs := make([]int, 0, len(simItems))
	for _, it := range simItems {
		simIDs = append(simIDs, it.SimilarMovieID)
	}

	simMovies, err := s.movies.GetMoviesByIDs(ctx, simIDs)
	if err != nil {
		return nil, err
	}

	simCards := make([]domain.MovieCard, 0, len(simMovies))
	for i := range simMovies {
		simCards = append(simCards, toCard(simMovies[i], nil, nil, nil))
	}

	return &domain.MovieDetailsResponse{
		Movie:    *m,
		Poster:   meta.Poster,
		Details:  meta.Details,
		UserRate: userRate,
		Similar:  simCards,
	}, nil
}

const defaultSimilarLimit = 5

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
		s.log.WithFields(logrus.Fields{"movie_id": movieID, "n": len(items)}).Debug("similar cache hit (db)")
		return items, nil
	}

	s.log.WithField("movie_id", movieID).Debug("similar cache miss (db)")

	if s.recClient == nil {
		s.log.WithField("movie_id", movieID).Warn("recClient is nil; skip rec-service call")
		return []domain.SimilarItem{}, nil
	}

	recCtx, cancel := context.WithTimeout(ctx, s.cfg.RecTimeout)
	defer cancel()

	s.log.WithFields(logrus.Fields{"movie_id": movieID, "limit": limit}).Debug("calling rec-service SimilarMovies")

	items, err = s.recClient.SimilarMovies(recCtx, movieID, limit)
	if err != nil {
		s.log.WithError(err).WithField("movie_id", movieID).Warn("rec service similar failed")
		return []domain.SimilarItem{}, nil
	}

	s.log.WithFields(logrus.Fields{"movie_id": movieID, "n": len(items)}).Debug("rec-service returned similar")

	if err := s.similar.ReplaceForMovie(ctx, movieID, items); err != nil {
		s.log.WithError(err).WithField("movie_id", movieID).Warn("failed to save similar to db")
	}

	return items, nil
}

func (s *MoviesService) getMovieMeta(ctx context.Context, movieID int, imdbID *int, allowFetch bool) (domain.MovieMeta, error) {
	var meta domain.MovieMeta

	if s.posters != nil {
		p, err := s.posters.GetPoster(ctx, movieID)
		if err == nil && p != nil && p.PosterURL != "" {
			meta.Poster = &p.PosterURL
		} else if err != nil && err != pgx.ErrNoRows {
			s.log.WithError(err).WithField("movie_id", movieID).Warn("get poster failed")
		}
	}

	if s.details != nil {
		d, err := s.details.GetMovieDetails(ctx, movieID)
		if err == nil && d != nil && len(d.Payload) > 0 {
			meta.Details = json.RawMessage(d.Payload)
		} else if err != nil && err != pgx.ErrNoRows {
			s.log.WithError(err).WithField("movie_id", movieID).Warn("get movie_details failed")
		}
	}

	needDetails := len(meta.Details) == 0
	needPoster := meta.Poster == nil || *meta.Poster == ""

	if !allowFetch || s.omdb == nil || imdbID == nil || *imdbID <= 0 || (!needDetails && !needPoster) {
		return meta, nil
	}

	imdb := fmt.Sprintf("tt%07d", *imdbID)
	omdbCtx, cancel := context.WithTimeout(ctx, s.cfg.OMDbTimeout)
	defer cancel()

	payload, poster, err := s.omdb.FetchMovie(omdbCtx, imdb)
	if err != nil {
		s.log.WithError(err).WithField("imdb_id", imdb).Warn("omdb fetch failed")
		return meta, nil
	}

	if needDetails && len(payload) > 0 {
		meta.Details = json.RawMessage(payload)
		if s.details != nil {
			_ = s.details.UpsertMovieDetails(ctx, movieID, payload)
		}
	}

	if needPoster && poster != nil && *poster != "" {
		meta.Poster = poster
		if s.posters != nil {
			_ = s.posters.UpsertPoster(ctx, movieID, *poster)
		}
	}

	return meta, nil
}
