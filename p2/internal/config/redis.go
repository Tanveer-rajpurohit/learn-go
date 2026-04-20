package config

import (
    "os"
    "github.com/redis/go-redis/v9"
)

func ConnectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
}