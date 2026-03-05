package service

import (
	"context"
	"encoding/json"
	"fmt"

	"my_mdb/internal/config"
	"my_mdb/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type MoviesServiceDeps struct {
	Movies   MoviesRepo
	Ratings  RatingsRepo
	Posters  PostersRepo
	Details  MovieDetailsRepo
	Similar  SimilarityRepo
	RecsRepo RecommendationsRepo
	Tags     TagsRepo

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
	recsRepo  RecommendationsRepo
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
		recsRepo:  d.RecsRepo,
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

	pm, err := s.PosterMapForMovies(ctx, mvs, false)
	if err != nil {
		return nil, err
	}

	return mapMoviesToCards(mvs, pm, CardMapOpt{}), nil
}

func (s *MoviesService) Genres(ctx context.Context) ([]string, error) {
	return s.movies.ListGenres(ctx)
}

func (s *MoviesService) ByGenre(ctx context.Context, genre string, limit int) ([]domain.MovieCard, error) {
	mvs, err := s.movies.ListMoviesByGenre(ctx, genre, limit)
	if err != nil {
		return nil, err
	}

	pm, err := s.PosterMapForMovies(ctx, mvs, false)
	if err != nil {
		return nil, err
	}

	return mapMoviesToCards(mvs, pm, CardMapOpt{}), nil
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

	pm, err := s.PosterMapForMovies(ctx, mvs, true)
	if err != nil {
		return nil, err
	}

	userRates := make(map[int]*float32, len(mvs))
	for i := range mvs {
		r, err := s.ratings.GetUserRatingForMovie(ctx, userID, mvs[i].ID)
		if err != nil {
			return nil, err
		}
		userRates[mvs[i].ID] = r
	}

	return mapMoviesToCards(mvs, pm, CardMapOpt{UserRates: userRates}), nil
}

func (s *MoviesService) Rate(ctx context.Context, userID, movieID int, v float32) error {
	if v < 0 || v > 5 {
		return fmt.Errorf("rating must be between 0 and 5")
	}

	_, err := s.movies.GetMovieByID(ctx, movieID)
	if err != nil {
		return err
	}

	if err := s.ratings.UpsertRating(ctx, userID, movieID, v); err != nil {
		return err
	}

	s.log.WithFields(logrus.Fields{"user_id": userID, "movie_id": movieID, "value": v}).Info("rating upserted")

	return nil
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

func (s *MoviesService) Recommendation(ctx context.Context, userID int) ([]domain.MovieCard, error) {
	mvs, err := s.GetForYouMovies(ctx, userID, 60)
	if err != nil {
		return nil, err
	}
	if len(mvs) == 0 {
		return []domain.MovieCard{}, nil
	}

	var pm map[int]string
	if s.posters != nil {
		pm, err = s.PosterMapForMovies(ctx, mvs, true)
		if err != nil {
			return nil, err
		}
	}

	return mapMoviesToCards(mvs, pm, CardMapOpt{}), nil
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

	platformRate, err := s.ratings.GetWeightedScoreForMovie(ctx, movieID)
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

	pm, err := s.PosterMapForMovies(ctx, simMovies, true)
	if err != nil {
		return nil, err
	}

	simCards := mapMoviesToCards(simMovies, pm, CardMapOpt{})

	return &domain.MovieDetailsResponse{
		Movie:        *m,
		Poster:       meta.Poster,
		Details:      meta.Details,
		UserRate:     userRate,
		PlatformRate: platformRate,
		Similar:      simCards,
	}, nil
}

const defaultSimilarLimit = 6

func (s *MoviesService) getMovieMeta(ctx context.Context, movieID int, imdbID *int, allowFetch bool) (domain.MovieMeta, error) {
	var meta domain.MovieMeta

	if s.posters != nil {
		p, err := s.posters.GetPosterByMovieID(ctx, movieID)
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

func (s *MoviesService) PosterMapForMovies(ctx context.Context, mvs []domain.Movie, allowWarmup bool) (map[int]string, error) {
	out := make(map[int]string, len(mvs))
	if len(mvs) == 0 {
		return out, nil
	}

	ids := make([]int, 0, len(mvs))
	for i := range mvs {
		ids = append(ids, mvs[i].ID)
	}

	if s.posters != nil {
		pm, err := s.posters.GetPostersByMovieIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		out = pm
	}

	if allowWarmup && len(mvs) <= 100 {
		_ = s.WarmupMissingPosters(ctx, mvs, out)

		if s.posters != nil {
			pm, err := s.posters.GetPostersByMovieIDs(ctx, ids)
			if err == nil {
				out = pm
			}
		}
	}

	return out, nil
}

func (s *MoviesService) WarmupMissingPosters(ctx context.Context, movies []domain.Movie, existing map[int]string) error {
	if s.omdb == nil || s.posters == nil {
		return nil
	}

	sem := make(chan struct{}, 3)
	g, gctx := errgroup.WithContext(ctx)

	for i := range movies {
		mv := movies[i]

		if url, ok := existing[mv.ID]; ok && url != "" && url != "N/A" {
			continue
		}
		if mv.IMDbID == nil || *mv.IMDbID <= 0 {
			continue
		}

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			imdb := fmt.Sprintf("tt%07d", *mv.IMDbID)

			omdbCtx, cancel := context.WithTimeout(gctx, s.cfg.OMDbTimeout)
			defer cancel()

			_, poster, err := s.omdb.FetchMovie(omdbCtx, imdb)
			if err != nil || poster == nil || *poster == "" || *poster == "N/A" {
				return nil
			}

			if err := s.posters.UpsertPoster(gctx, mv.ID, *poster); err != nil && s.log != nil {
				s.log.WithError(err).WithField("movie_id", mv.ID).Warn("upsert poster failed")
			}
			return nil
		})
	}

	return g.Wait()
}
