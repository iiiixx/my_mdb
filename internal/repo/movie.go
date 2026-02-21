package repo

import (
	"context"
	"fmt"
	"strings"

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

	like := "%" + q + "%"

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, title, year, genres, imdb_id, tmdb_id
		FROM movies
		WHERE title ILIKE $1
		ORDER BY year DESC NULLS LAST, title ASC
		LIMIT $2
	`, like, limit)
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
		SELECT movie_id, title, year, genres, imdb_id, tmdb_id
		FROM movies
		WHERE $1 = ANY(genres)
		ORDER BY year DESC NULLS LAST, title ASC
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
		FROM top_movies t
		JOIN movies m ON m.movie_id = t.movie_id
		ORDER BY t.score DESC
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
		return nil, fmt.Errorf("take (%d) must be <= topN (%d)", take, topN)
	}

	rows, err := r.pool.Query(ctx, `
		WITH top AS (
			SELECT movie_id
			FROM top_movies
			ORDER BY score DESC
			LIMIT $1
		)
		SELECT m.movie_id, m.title, m.year, m.genres, m.imdb_id, m.tmdb_id
		FROM top
		JOIN movies m ON m.movie_id = top.movie_id
		ORDER BY random()
		LIMIT $2
	`, topN, take)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Movie, 0, take)
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Genres, &m.IMDbID, &m.TMDBID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
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
