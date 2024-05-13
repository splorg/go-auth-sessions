package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/splorg/go-auth-sessions/internal/auth"
	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/database"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDRESS is not defined in .env")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		log.Fatal("REDIS_PASSWORD is not defined in .env")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not defined in .env")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err.Error())
	}

	var testQuery int

	err = conn.QueryRow("SELECT 1").Scan(&testQuery)
	if err != nil {
		log.Fatalf("database connection test failed: %v", err.Error())
	} else {
		log.Print("connection test query executed successfully")
	}

	apiConfig := &config.ApiConfig{
		DB:          database.New(conn),
		RedisClient: redisClient,
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{}))

	authHandler := auth.NewAuthHandler(apiConfig)

	app.Get("/healthcheck", authHandler.HealthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not defined in .env")
	}

	app.Listen(port)
}
