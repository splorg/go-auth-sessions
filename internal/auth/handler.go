package auth

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/util"
	"github.com/splorg/go-auth-sessions/internal/validator"
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

func (h *AuthHandler) SignIn(c *fiber.Ctx) error {
  var req LoginDTO

  if err := c.BodyParser(&req); err != nil {
    return err
  }

  if err := validator.ValidateStruct(req); err != nil {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
  }

  foundUser, err := h.DB.FindUserByEmail(c.Context(), req.Email)
  if err != nil {
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no user found"})
  }

  if err := util.ComparePassword([]byte(foundUser.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

  expirationTime := time.Now().Add(10 * time.Minute)
  
  claims := &Claims{
    Email: foundUser.Email,
    StandardClaims: jwt.StandardClaims{
      Subject: foundUser.ID.String(),
      ExpiresAt: expirationTime.Unix(),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

  tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
  }

  sessionID := uuid.New().String()

  sessionData := map[string]interface{}{
    "token": tokenString,
    "userId": foundUser.ID,
  }

  sessionDataJSON, err := json.Marshal(sessionData)
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to encode session data into JSON"})
  }

  err = h.RedisClient.Set(c.Context(), sessionID, sessionDataJSON, time.Until(expirationTime)).Err()
  if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save session data to Redis store"})
  }

  cookie := &fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expirationTime,
		HTTPOnly: true,
	}

  c.Cookie(cookie)

  return c.Status(fiber.StatusOK).JSON(foundUser)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
  sessionID := c.Cookies("session_id")
  if sessionID == "" {
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
  }

  err := h.RedisClient.Del(c.Context(), sessionID).Err()
  if err != nil {
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "failed to delete session"})
  }

  c.ClearCookie("session_id")

  return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "logged out successfully"})
}