package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg *config.Config) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Pass,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Failed to parse DB config: %v", err)
	}

	poolConfig.MaxConns = cfg.DB.PoolMaxConns
	poolConfig.MinConns = cfg.DB.PoolMinConns
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to create DB pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	log.Println("PostgreSQL connection pool established")
	return pool
}
