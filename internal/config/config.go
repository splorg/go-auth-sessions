package config

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/splorg/go-auth-sessions/internal/database"
)

type SessionData struct {
	UserId    uuid.UUID `json:"userId"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
}

type ApiConfig struct {
	DB    *database.Queries
	Redis *redis.Client
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
		DB:    database.New(conn),
		Redis: redisClient,
	}

	return apiConfig, nil
}
