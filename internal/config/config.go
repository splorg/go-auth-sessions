package config

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/splorg/go-auth-sessions/internal/database"
)

type ApiConfig struct {
  DB *database.Queries
  RedisClient *redis.Client
}

func NewApiConfig() (*ApiConfig, error) {
  redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
    return nil, errors.New("REDIS_ADDRESS is not defined in .env")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
    return nil, errors.New("REDIS_PASSWORD is not defined in .env")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
    return nil, errors.New("DB_URL is not defined in .env")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
    return nil, errors.New("cannot connect to database")
	}

	var testQuery int

	err = conn.QueryRow("SELECT 1").Scan(&testQuery)
	if err != nil {
    return nil, errors.New("database connection test failed")
	} else {
		log.Print("connection test query executed successfully")
	}

	apiConfig := &ApiConfig{
		DB:          database.New(conn),
		RedisClient: redisClient,
	}

  return apiConfig, nil
}