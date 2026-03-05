package service

import (
	"context"
	"time"

	domain "my_mdb/internal/domain"
)

type UsersRepo interface {
	Create(ctx context.Context) (int, error)
	Exists(ctx context.Context, userID int) (bool, error)
}

type MoviesRepo interface {
	GetMovieByID(ctx context.Context, movieID int) (*domain.Movie, error)
	GetMoviesByIDs(ctx context.Context, movieIDs []int) ([]domain.Movie, error)
	SearchMovies(ctx context.Context, q string, limit int) ([]domain.Movie, error)
	ListGenres(ctx context.Context) ([]string, error)
	ListMoviesByGenre(ctx context.Context, genre string, limit int) ([]domain.Movie, error)
	TopMovies(ctx context.Context, limit int) ([]domain.Movie, error)
	RandomFromTop(ctx context.Context, topN, take int) ([]domain.Movie, error)
	MoviesByYearRange(ctx context.Context, yearFrom, yearTo int16, limit int) ([]domain.Movie, error)
	RefreshWeightedScore(ctx context.Context, m float64) error
	ListTopMissingMeta(ctx context.Context, limit int) ([]domain.Movie, error)
	NewReleasesWithGoodRating(ctx context.Context, limit int, minScore float32, minVotes int) ([]domain.Movie, error)
}

type RatingsRepo interface {
	CountByUser(ctx context.Context, userID int) (int, error)
	GetUserRatingForMovie(ctx context.Context, userID, movieID int) (*float32, error)
	ListUserRatedMovieIDs(ctx context.Context, userID, limit int) ([]int, error)
	UpsertRating(ctx context.Context, userID, movieID int, value float32) error
	GetWeightedScoreForMovie(ctx context.Context, movieID int) (*float32, error)
}

type PostersRepo interface {
	GetPosterByMovieID(ctx context.Context, movieID int) (*domain.Poster, error)
	UpsertPoster(ctx context.Context, movieID int, posterURL string) error
	GetPostersByMovieIDs(ctx context.Context, ids []int) (map[int]string, error)
}

type MovieDetailsRepo interface {
	GetMovieDetails(ctx context.Context, movieID int) (*domain.MovieDetailsCache, error)
	UpsertMovieDetails(ctx context.Context, movieID int, payload []byte) error
}

type RecommendationsRepo interface {
	GetByUser(ctx context.Context, userID, limit int) ([]domain.RecommendationItem, error)
	GetLastUpdatedAt(ctx context.Context, userID int) (*time.Time, error)
	ReplaceForUser(ctx context.Context, userID int, items []domain.RecommendationItem) error
}

type SimilarityRepo interface {
	GetSimilar(ctx context.Context, movieID int, limit int) ([]domain.SimilarItem, error)
	InsertIfNotExists(ctx context.Context, movieID int, items []domain.SimilarItem) error
}

type TagsRepo interface {
	TopMoviesByTagQuery(ctx context.Context, tagQuery string, limit int) ([]int, error)
}

type OMDbClient interface {
	FetchMovie(ctx context.Context, imdbID string) (detailsJSON []byte, posterURL *string, err error)
}

type RecClient interface {
	Recommend(ctx context.Context, userID int, limit int, excludeMovieIDs []int) ([]domain.RecommendationItem, error)
	SimilarMovies(ctx context.Context, movieID int, limit int) ([]domain.SimilarItem, error)
}

type PosterProvider interface {
	PosterMapForMovies(ctx context.Context, mvs []domain.Movie, allowFetchMissing bool) (map[int]string, error)
}

type RecProvider interface {
	GetForYouMovies(ctx context.Context, userID int, limit int) ([]domain.Movie, error)
}
