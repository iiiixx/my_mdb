package repo

import (
	"context"
	"my_mdb/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SimilarityRepo struct {
	pool *pgxpool.Pool
}

func NewSimilarityRepo(pool *pgxpool.Pool) *SimilarityRepo {
	return &SimilarityRepo{pool: pool}
}

func (r *SimilarityRepo) GetSimilar(ctx context.Context, movieID int, limit int) ([]domain.SimilarItem, error) {
	if limit <= 0 {
		limit = 5
	}

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, similar_movie_id, score
		FROM movie_similarity
		WHERE movie_id = $1
		ORDER BY score DESC
		LIMIT $2
	`, movieID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.SimilarItem, 0, limit)
	for rows.Next() {
		var it domain.SimilarItem
		if err := rows.Scan(&it.MovieID, &it.SimilarMovieID, &it.Score); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
