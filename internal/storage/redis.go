package storage

import (
	redis "github.com/redis/go-redis/v9"
)

var (
	Rdb *redis.Client
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}
