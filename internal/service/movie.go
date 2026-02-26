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

	OMDb OMDbClient

	Cfg config.Config
	Log *logrus.Logger
}

type MoviesService struct {
	movies  MoviesRepo
	ratings RatingsRepo
	posters PostersRepo
	details MovieDetailsRepo
	similar SimilarityRepo
	tags    TagsRepo
	omdb    OMDbClient
	cfg     config.Config
	log     *logrus.Logger
}

func NewMoviesService(d MoviesServiceDeps) *MoviesService {
	return &MoviesService{
		movies:  d.Movies,
		ratings: d.Ratings,
		posters: d.Posters,
		details: d.Details,
		similar: d.Similar,
		tags:    d.Tags,
		omdb:    d.OMDb,
		cfg:     d.Cfg,
		log:     d.Log,
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
	ids, err := s.ratings.ListUserRatings(ctx, userID, limit)
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

type MovieDetailsResponse struct {
	Movie    domain.Movie
	Poster   *string
	Details  json.RawMessage
	UserRate *float32
	Similar  []domain.MovieCard
}

func (s *MoviesService) GetMovieDetails(ctx context.Context, userID int, movieID int) (*MovieDetailsResponse, error) {
	m, err := s.movies.GetMovieByID(ctx, movieID)
	if err != nil {
		return nil, err
	}

	userRate, err := s.ratings.GetUserRatingForMovie(ctx, userID, movieID)
	if err != nil {
		return nil, err
	}

	var posterURL *string
	if s.posters != nil {
		p, err := s.posters.GetPoster(ctx, movieID)
		if err == nil && p != nil && p.PosterURL != "" {
			posterURL = &p.PosterURL
		}
	}

	var details json.RawMessage
	if s.details != nil {
		d, err := s.details.GetMovieDetails(ctx, movieID)
		if err == nil && d != nil && len(d.Payload) > 0 {
			details = json.RawMessage(d.Payload)
		} else if err != nil && err != pgx.ErrNoRows {
			s.log.WithError(err).WithField("movie_id", movieID).Warn("get movie_details failed")
		}
	}

	if len(details) == 0 && s.omdb != nil && m.IMDbID != nil && *m.IMDbID > 0 {
		imdb := fmt.Sprintf("tt%07d", *m.IMDbID)

		omdbCtx, cancel := context.WithTimeout(ctx, s.cfg.OMDbTimeout)
		defer cancel()

		payload, poster, err := s.omdb.FetchMovie(omdbCtx, imdb)
		if err != nil {
			s.log.WithError(err).WithField("imdb_id", imdb).Warn("omdb fetch failed")
		} else {
			if len(payload) > 0 {
				details = json.RawMessage(payload)
				if s.details != nil {
					_ = s.details.UpsertMovieDetails(ctx, movieID, payload)
				}
			}
			if poster != nil && *poster != "" {
				posterURL = poster
				if s.posters != nil {
					_ = s.posters.UpsertPoster(ctx, movieID, *poster)
				}
			}
		}
	}

	simItems, err := s.similar.GetSimilar(ctx, movieID, 5)
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

	return &MovieDetailsResponse{
		Movie:    *m,
		Poster:   posterURL,
		Details:  details,
		UserRate: userRate,
		Similar:  simCards,
	}, nil
}
