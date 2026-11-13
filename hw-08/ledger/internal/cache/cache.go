package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func Open(ctx context.Context) (*Cache, error) {
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
	return &Cache{client: client}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) GetReportSummary(ctx context.Context, from, to string) (map[string]float64, bool, error) {
	key := "report:summary:" + from + ":" + to
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var report map[string]float64
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, false, err
	}
	return report, true, nil
}

func (c *Cache) SetReportSummary(ctx context.Context, from, to string, report map[string]float64) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	key := "report:summary:" + from + ":" + to
	return c.client.Set(ctx, key, data, 30*time.Second).Err()
}
