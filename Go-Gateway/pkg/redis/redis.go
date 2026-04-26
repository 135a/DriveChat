package redis

import (
	"context"
	"log"

	"github.com/nym/go-gateway/internal/config"
	redisclient "github.com/redis/go-redis/v9"
)

var Client *redisclient.Client
var Ctx = context.Background()

func InitRedis() {
	Client = redisclient.NewClient(&redisclient.Options{
		Addr:     config.AppConfig.RedisAddr,
		Password: config.AppConfig.RedisPass,
		DB:       0,
	})

	if err := Client.Ping(Ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connection established.")
}
