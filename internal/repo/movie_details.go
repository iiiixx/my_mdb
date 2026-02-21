package repo

import (
	"context"
	"my_mdb/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieDetailsRepo struct {
	pool *pgxpool.Pool
}

func NewMovieDetailsRepo(pool *pgxpool.Pool) *MovieDetailsRepo {
	return &MovieDetailsRepo{pool: pool}
}

func (r *MovieDetailsRepo) GetMovieDetails(ctx context.Context, movieID int) (*domain.MovieDetailsCache, error) {
	var d domain.MovieDetailsCache
	err := r.pool.QueryRow(ctx, `
		SELECT movie_id, payload, updated_at
		FROM movie_details
		WHERE movie_id = $1
	`, movieID).Scan(&d.MovieID, &d.Payload, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *MovieDetailsRepo) UpsertMovieDetails(ctx context.Context, movieID int, payload []byte) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO movie_details(movie_id, payload, updated_at)
		VALUES ($1, $2::jsonb, now())
		ON CONFLICT (movie_id)
		DO UPDATE SET payload = EXCLUDED.payload, updated_at = now()
	`, movieID, payload)
	return err
}
