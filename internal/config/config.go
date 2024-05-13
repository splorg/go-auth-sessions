package config

import (
	"github.com/redis/go-redis/v9"
	"github.com/splorg/go-auth-sessions/internal/database"
)

type ApiConfig struct {
  DB *database.Queries
  RedisClient *redis.Client
}