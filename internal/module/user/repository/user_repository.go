package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/user/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"

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

func (r *UserRepository) ExistsByID(ctx context.Context, db database.PgxExt, userID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, userID).Scan(&exists)

	return exists, err
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*entity.User, error) {
	query := `
		SELECT
			id,
			full_name,
			username,
			email,
			avatar_url,
			status,
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
		&user.AvatarUrl,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByIDs(ctx context.Context, db database.PgxExt, userIDs []string) ([]entity.User, error) {
	query := `
		SELECT
			id,
			full_name,
			email
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := db.Query(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entity.User, 0)
	for rows.Next() {
		var user entity.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}
