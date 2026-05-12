package repository

import (
	"context"
	"strings"

	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) AuthLogin(ctx context.Context, identifier string) (*entity.User, error) {
	var query string
	var field string

	if strings.Contains(identifier, "@") {
		field = "email"
	} else {
		field = "username"
	}

	query = `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			status
		FROM users
		WHERE ` + field + ` = $1
		LIMIT 1
	`

	var user entity.User
	err := r.db.QueryRow(ctx, query, identifier).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			status,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var user entity.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT
			id,
			email,
			username,
			full_name,
			status,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var user entity.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Name,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			id,
			email,
			username,
			full_name,
			password_hash
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Username,
		user.Name,
		user.PasswordHash,
	)

	return err
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET last_login_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}
