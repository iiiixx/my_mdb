package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepo struct {
	pool *pgxpool.Pool
}

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{pool: pool}
}

func (r *UsersRepo) Create(ctx context.Context) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `
        INSERT INTO users DEFAULT VALUES
        RETURNING user_id
    `).Scan(&id)
	return id, err
}

func (r *UsersRepo) Exists(ctx context.Context, userID int) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
        SELECT EXISTS (SELECT 1 FROM users WHERE user_id = $1)
    `, userID).Scan(&ok)
	return ok, err
}
