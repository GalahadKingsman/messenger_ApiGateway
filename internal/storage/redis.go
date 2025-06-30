package storage

import (
	"context"
	"os"

	redis "github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
	Rdb *redis.Client
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}
