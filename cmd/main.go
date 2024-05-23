package main

import (
	"log"
	"net/http"
	"os"

	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/handler"
	"github.com/splorg/go-auth-sessions/internal/middleware"
	"github.com/splorg/go-auth-sessions/internal/router"
	"github.com/splorg/go-auth-sessions/internal/util"
	"github.com/splorg/go-auth-sessions/internal/validator"
)

func main() {
	validator.Setup()
	apiConfig, err := config.NewApiConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not defined in .env")
	}

	mux := http.NewServeMux()

	authHandler := handler.NewAuthHandler(apiConfig)

	public := router.NewRouteGroup(mux, apiConfig, middleware.Logging)
	protected := router.NewRouteGroup(mux, apiConfig, middleware.Logging, middleware.Authenticate)

	public.HandleFunc("POST /register", authHandler.Register)
	public.HandleFunc("POST /login", authHandler.Login)
	public.HandleFunc("POST /logout", authHandler.Logout)

	protected.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) *handler.APIError {
		util.WriteJson(w, http.StatusOK, map[string]string{"message": "OK"})
		return nil
	})

	server := http.Server{
		Addr:    port,
		Handler: mux,
	}

	server.ListenAndServe()
}
