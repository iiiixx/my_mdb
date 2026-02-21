package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RatingsRepo struct {
	pool *pgxpool.Pool
}

func NewRatingsRepo(pool *pgxpool.Pool) *RatingsRepo {
	return &RatingsRepo{pool: pool}
}

func (r *RatingsRepo) CountByUser(ctx context.Context, userID int) (int, error) {
	var cnt int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ratings WHERE user_id = $1
	`, userID).Scan(&cnt)
	return cnt, err
}

func (r *RatingsRepo) GetUserRatingForMovie(ctx context.Context, userID, movieID int) (*float32, error) {
	var rating float32
	err := r.pool.QueryRow(ctx, `
		SELECT rating
		FROM ratings
		WHERE user_id = $1 AND movie_id = $2
	`, userID, movieID).Scan(&rating)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &rating, nil
}

func (r *RatingsRepo) ListUserRatings(ctx context.Context, userID, limit int) ([]int, error) {
	if limit <= 0 {
		limit = 30
	}

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id
		FROM ratings
		WHERE user_id = $1
		ORDER BY ts DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var movieID int
		if err := rows.Scan(&movieID); err != nil {
			return nil, err
		}
		ids = append(ids, movieID)
	}

	return ids, rows.Err()
}

func (r *RatingsRepo) UpsertRating(ctx context.Context, userID, movieID int, value float32) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ratings (user_id, movie_id, rating, ts)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, movie_id)
		DO UPDATE SET rating = EXCLUDED.rating, ts = now()
	`, userID, movieID, value)

	return err
}
