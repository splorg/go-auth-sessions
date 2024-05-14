package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/splorg/go-auth-sessions/internal/auth"
	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/validator"
)

func main() {
	validator.Setup()
	apiConfig, err := config.NewApiConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{}))

	authHandler := auth.NewAuthHandler(apiConfig)

	app.Get("/healthcheck", authHandler.HealthCheck)
	app.Post("/register", authHandler.Register)
	app.Post("/login", authHandler.Login)
	app.Post("/logout", authHandler.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not defined in .env")
	}

	app.Listen(port)
}
