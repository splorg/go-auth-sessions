package router

import (
	"net/http"

	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/handler"
	"github.com/splorg/go-auth-sessions/internal/middleware"
	"github.com/splorg/go-auth-sessions/internal/util"
)

type routeGroup struct {
	mux         *http.ServeMux
	middlewares []middleware.Middleware
	config      *config.ApiConfig
}

func NewRouteGroup(mux *http.ServeMux, cfg *config.ApiConfig, middlewares ...middleware.Middleware) *routeGroup {
	return &routeGroup{mux: mux, middlewares: middlewares, config: cfg}
}

func (rg *routeGroup) HandleFunc(pattern string, handlerFunc handler.ApiHandler) {
	finalHandler := middlewareChain(httpHandlerAdapter(handlerFunc), rg.config, rg.middlewares...)
	rg.mux.Handle(pattern, finalHandler)
}

func httpHandlerAdapter(h handler.ApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			util.WriteJson(w, err.StatusCode, map[string]string{"error": err.Error()})
		}
	}
}

func middlewareChain(h http.Handler, cfg *config.ApiConfig, middlewares ...middleware.Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h, cfg)
	}
	return h
}
