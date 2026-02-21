package repo

import (
	"context"
	"my_mdb/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostersRepo struct {
	pool *pgxpool.Pool
}

func NewPostersRepo(pool *pgxpool.Pool) *PostersRepo {
	return &PostersRepo{pool: pool}
}

func (r *PostersRepo) GetPoster(ctx context.Context, movieID int) (*domain.Poster, error) {
	var p domain.Poster
	err := r.pool.QueryRow(ctx, `
		SELECT movie_id, poster_url, updated_at
		FROM posters
		WHERE movie_id = $1
	`, movieID).Scan(&p.MovieID, &p.PosterURL, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostersRepo) UpsertPoster(ctx context.Context, movieID int, posterURL string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO posters(movie_id, poster_url, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (movie_id)
		DO UPDATE SET poster_url = EXCLUDED.poster_url, updated_at = now()
	`, movieID, posterURL)
	return err
}
