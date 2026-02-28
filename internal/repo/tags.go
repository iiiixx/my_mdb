package repo

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TagsRepo struct {
	pool *pgxpool.Pool
}

func NewTagsRepo(pool *pgxpool.Pool) *TagsRepo {
	return &TagsRepo{pool: pool}
}

func (r *TagsRepo) TopMoviesByTagQuery(
	ctx context.Context,
	tagQuery string,
	limit int,
) ([]int, error) {

	tagQuery = strings.TrimSpace(tagQuery)
	if tagQuery == "" {
		return []int{}, nil
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx, `
		SELECT gs.movie_id
		FROM genome_tags gt
		JOIN genome_scores gs ON gs.tag_id = gt.tag_id
		WHERE gt.tag ILIKE $1
		GROUP BY gs.movie_id
		ORDER BY MAX(gs.relevance) DESC
		LIMIT $2
	`, "%"+tagQuery+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []int
	for rows.Next() {
		var movieID int
		if err := rows.Scan(&movieID); err != nil {
			return nil, err
		}
		out = append(out, movieID)
	}
	return out, rows.Err()
}
