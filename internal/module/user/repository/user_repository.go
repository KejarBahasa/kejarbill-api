package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(
	db *pgxpool.Pool,
) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*entity.User, error) {
	query := `
		SELECT
			id,
			full_name,
			username,
			email,
			status,
			token_version,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var user entity.User

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.Email,
		&user.Status,
		&user.TokenVersion,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
