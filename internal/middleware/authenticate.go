package middleware

import (
	"net/http"

	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/handler"
)

func Authenticate(next http.Handler, cfg *config.ApiConfig) http.Handler {
	authHandler := handler.NewAuthHandler(cfg)
	return authHandler.AuthenticationMiddleware(next)
}
