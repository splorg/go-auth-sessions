package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/splorg/go-auth-sessions/internal/config"
)

type AuthHandler struct {
  *config.ApiConfig
}

func NewAuthHandler(config *config.ApiConfig) *AuthHandler {
  return &AuthHandler{ApiConfig: config}
}

// adding this to auth handler to avoid creating a new handler
func (h *AuthHandler) HealthCheck(c *fiber.Ctx) error {
  return c.JSON(fiber.Map{
    "message": "ok",
  })
}