package middleware

import (
	"net/http"

	"github.com/splorg/go-auth-sessions/internal/config"
)

type Middleware func(http.Handler, *config.ApiConfig) http.Handler

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}
