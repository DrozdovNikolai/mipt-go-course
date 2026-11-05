package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func Open(ctx context.Context) (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	dbNumber := 0
	if value := os.Getenv("REDIS_DB"); value != "" {
		var err error
		dbNumber, err = strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
		}
	}
	client := redis.NewClient(&redis.Options{
		Addr: addr, DB: dbNumber, Password: os.Getenv("REDIS_PASSWORD"),
	})
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}
	return client, nil
}
