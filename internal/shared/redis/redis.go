package redis

import (
	"context"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewRedis(
	addr,
	pw string,
) *goredis.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := goredis.NewClient(
		&goredis.Options{
			Addr:         addr,
			Password:     pw,
			DB:           0,
			PoolSize:     20,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
	)

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis client connected")
	return rdb
}
