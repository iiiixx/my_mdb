package repo

import (
	"context"
	"my_mdb/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecommendationsRepo struct {
	pool *pgxpool.Pool
}

func NewRecommendationsRepo(pool *pgxpool.Pool) *RecommendationsRepo {
	return &RecommendationsRepo{pool: pool}
}

func (r *RecommendationsRepo) GetByUser(ctx context.Context, userID, limit int) ([]domain.RecommendationItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, score
		FROM recommendations
		WHERE user_id = $1
		ORDER BY score DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.RecommendationItem, 0, limit)
	for rows.Next() {
		var it domain.RecommendationItem
		if err := rows.Scan(&it.MovieID, &it.Score); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *RecommendationsRepo) GetLastUpdatedAt(ctx context.Context, userID int) (*time.Time, error) {
	var ts *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(updated_at)
		FROM recommendations
		WHERE user_id = $1
	`, userID).Scan(&ts)
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func (r *RecommendationsRepo) ReplaceForUser(ctx context.Context, userID int, items []domain.RecommendationItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM recommendations WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	for _, it := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO recommendations(user_id, movie_id, score, updated_at)
			VALUES ($1, $2, $3, now())
		`, userID, it.MovieID, it.Score)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
