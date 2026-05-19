package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTransaction runs a function within a database transaction.
func WithTransaction(ctx context.Context, db *pgxpool.Pool, fn func(dbQ PgxExt) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
