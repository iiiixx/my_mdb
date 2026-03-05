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

func (r *PostersRepo) GetPosterByMovieID(ctx context.Context, movieID int) (*domain.Poster, error) {
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

func (r *PostersRepo) GetPostersByMovieIDs(ctx context.Context, ids []int) (map[int]string, error) {
	out := make(map[int]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT movie_id, poster_url
		FROM posters
		WHERE movie_id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			return nil, err
		}
		if url != "" && url != "N/A" {
			out[id] = url
		}
	}
	return out, rows.Err()
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
