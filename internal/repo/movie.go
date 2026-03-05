package repo

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"my_mdb/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MoviesRepo struct {
	pool *pgxpool.Pool
}

func NewMoviesRepo(pool *pgxpool.Pool) *MoviesRepo {
	return &MoviesRepo{pool: pool}
}

func (r *MoviesRepo) GetMovieByID(ctx context.Context, movieID int) (*domain.Movie, error) {
	var m domain.Movie
	err := r.pool.QueryRow(ctx, `
		SELECT movie_id, title, year, genres, imdb_id, tmdb_id
		FROM movies
		WHERE movie_id = $1
	`, movieID).Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MoviesRepo) GetMoviesByIDs(ctx context.Context, movieIDs []int) ([]domain.Movie, error) {
	if len(movieIDs) == 0 {
		return []domain.Movie{}, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, title, year, genres, imdb_id, tmdb_id
		FROM movies
		WHERE movie_id = ANY($1)
	`, movieIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, len(movieIDs))
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MoviesRepo) SearchMovies(ctx context.Context, q string, limit int) ([]domain.Movie, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.Movie{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	prefix := q + "%"
	any := "%" + q + "%"

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, title, year, genres, imdb_id, tmdb_id
		FROM movies
		WHERE title ILIKE $2
		ORDER BY
			CASE WHEN title ILIKE $1 THEN 0 ELSE 1 END,
			POSITION(LOWER($4) IN LOWER(title)) ASC,
			title ASC
		LIMIT $3
	`, prefix, any, limit, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, limit)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MoviesRepo) ListGenres(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT g.genre
		FROM (
			SELECT unnest(genres) AS genre
			FROM movies
		) g
		WHERE g.genre <> '' AND g.genre IS NOT NULL
		ORDER BY g.genre
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			return nil, err
		}
		out = append(out, genre)
	}
	return out, rows.Err()
}

func (r *MoviesRepo) ListMoviesByGenre(ctx context.Context, genre string, limit int) ([]domain.Movie, error) {
	genre = strings.TrimSpace(genre)
	if genre == "" {
		return []domain.Movie{}, nil
	}
	if limit <= 0 {
		limit = 30
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.movie_id, m.title, m.year, m.genres, m.imdb_id, m.tmdb_id
		FROM movies m
		LEFT JOIN movie_rating_stats s ON s.movie_id = m.movie_id
		WHERE $1 = ANY(m.genres)
		ORDER BY 
			s.weighted_score DESC NULLS LAST,
			s.votes DESC NULLS LAST,
			m.title ASC
		LIMIT $2
	`, genre, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, limit)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MoviesRepo) TopMovies(ctx context.Context, limit int) ([]domain.Movie, error) {
	if limit <= 0 {
		limit = 200
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.movie_id, m.title, m.year, m.genres, m.imdb_id, m.tmdb_id
		FROM movie_rating_stats s
		JOIN movies m ON m.movie_id = s.movie_id
		WHERE s.weighted_score IS NOT NULL
		ORDER BY s.weighted_score DESC, s.votes DESC, m.title ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, limit)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MoviesRepo) RandomFromTop(ctx context.Context, topN, take int) ([]domain.Movie, error) {
	if topN <= 0 {
		topN = 200
	}
	if take <= 0 {
		take = 5
	}
	if take > topN {
		return nil, fmt.Errorf("random from top: take (%d) must be <= topN (%d)", take, topN)
	}

	top, err := r.TopMovies(ctx, topN)
	if err != nil {
		return nil, err
	}
	if len(top) == 0 {
		return []domain.Movie{}, nil
	}
	if take > len(top) {
		take = len(top)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(top), func(i, j int) { top[i], top[j] = top[j], top[i] })

	return top[:take], nil
}

func (r *MoviesRepo) MoviesByYearRange(ctx context.Context, yearFrom, yearTo int16, limit int) ([]domain.Movie, error) {
	if limit <= 0 {
		limit = 30
	}

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, title, year, genres, imdb_id, tmdb_id
		FROM movies
		WHERE year BETWEEN $1 AND $2
		ORDER BY year DESC, title ASC
		LIMIT $3
	`, yearFrom, yearTo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, limit)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MoviesRepo) RefreshWeightedScore(ctx context.Context, m float64) error {
	const lockID int64 = 9238471

	var locked bool
	if err := r.pool.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, lockID).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	defer func() {
		_, _ = r.pool.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockID)
	}()

	_, err := r.pool.Exec(ctx, `SELECT refresh_weighted_score($1)`, m)
	return err
}

func (r *MoviesRepo) NewReleasesWithGoodRating(ctx context.Context, limit int, minScore float32, minVotes int) ([]domain.Movie, error) {
	if limit <= 0 {
		limit = 30
	}
	if minScore <= 0 {
		minScore = 3.8
	}
	if minVotes < 0 {
		minVotes = 0
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			m.movie_id, m.title, m.year, m.genres, m.imdb_id, m.tmdb_id
		FROM movies m
		JOIN movie_rating_stats s ON s.movie_id = m.movie_id
		WHERE
			m.year IS NOT NULL
			AND s.weighted_score IS NOT NULL
			AND s.weighted_score >= $1
			AND ($2 = 0 OR COALESCE(s.votes, 0) >= $2)
		ORDER BY
			m.year DESC,
			s.weighted_score DESC,
			m.title ASC
		LIMIT $3
	`, minScore, minVotes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, limit)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *MoviesRepo) ListTopMissingMeta(ctx context.Context, limit int) ([]domain.Movie, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			m.movie_id,
			m.title,
			m.year,
			m.genres,
			m.imdb_id,
			m.tmdb_id
		FROM movies m
		JOIN movie_rating_stats ms
			ON ms.movie_id = m.movie_id
		LEFT JOIN movie_details d
			ON d.movie_id = m.movie_id
		WHERE
			d.movie_id IS NULL
			OR d.payload IS NULL
		ORDER BY ms.weighted_score DESC NULLS LAST, m.movie_id ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, limit)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
