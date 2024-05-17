package auth

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/database"
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

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterDTO

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validator.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	password, err := util.HashPassword([]byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encrypt password"})
	}

	newUser, err := h.DB.CreateUser(c.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      req.Name,
		Username:  req.Username,
		Password:  string(password),
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(newUser)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginDTO

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
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

	sessionID := uuid.New().String()

	sessionData := SessionData{
		UserId:    foundUser.ID,
		Email:     foundUser.Email,
		ExpiresAt: expirationTime,
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

func (h *AuthHandler) AuthenticationMiddleware(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")

	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	sessionDataJSON, err := h.RedisClient.Get(c.Context(), sessionID).Result()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired session"})
	}

	var sessionData SessionData

	err = json.Unmarshal([]byte(sessionDataJSON), &sessionData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to decode session data"})
	}

	if sessionData.ExpiresAt.Unix() < time.Now().Unix() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	c.Locals("userId", sessionData.UserId)
	return c.Next()
}
