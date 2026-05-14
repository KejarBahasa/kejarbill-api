package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func WithTransaction(
	ctx context.Context,

	db *pgxpool.Pool,

	fn func(tx pgx.Tx) error,
) error {

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
