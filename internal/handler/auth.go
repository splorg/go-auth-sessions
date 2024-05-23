package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/splorg/go-auth-sessions/internal/config"
	"github.com/splorg/go-auth-sessions/internal/database"
	"github.com/splorg/go-auth-sessions/internal/util"
	"github.com/splorg/go-auth-sessions/internal/validator"
)

type authHandler struct {
	*config.ApiConfig
}

type registerDTO struct {
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type contextKey string

const userIDKey contextKey = "userId"

func NewAuthHandler(cfg *config.ApiConfig) *authHandler {
	return &authHandler{ApiConfig: cfg}
}

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) *APIError {
	var req registerDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewAPIError(http.StatusUnprocessableEntity, err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		return NewAPIError(http.StatusBadRequest, err)
	}

	password, err := util.HashPassword([]byte(req.Password))
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, errors.New("failed to encrypt password"))
	}

	newUser, err := h.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      req.Name,
		Username:  req.Username,
		Password:  string(password),
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, err)
	}

	util.WriteJson(w, http.StatusCreated, newUser)

	return nil
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) *APIError {
	var req loginDTO

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewAPIError(http.StatusUnprocessableEntity, err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		return NewAPIError(http.StatusBadRequest, err)
	}

	foundUser, err := h.DB.FindUserByEmail(r.Context(), req.Email)
	if err != nil {
		return NewAPIError(http.StatusNotFound, errors.New("no user found"))
	}

	if err := util.ComparePassword([]byte(foundUser.Password), []byte(req.Password)); err != nil {
		return NewAPIError(http.StatusUnauthorized, errors.New("invalid credentials"))
	}

	expirationTime := time.Now().Add(10 * time.Minute)
	sessionID := uuid.New().String()
	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr

	sessionData := config.SessionData{
		UserId:    foundUser.ID,
		Username:  foundUser.Username,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Email:     foundUser.Email,
		ExpiresAt: expirationTime,
		CreatedAt: time.Now(),
	}

	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, errors.New("failed to encode session data into JSON"))
	}

	err = h.Redis.Set(r.Context(), sessionID, sessionDataJSON, time.Until(expirationTime)).Err()
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, errors.New("failed to save session data to Redis store"))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expirationTime,
		HttpOnly: true,
	})

	util.WriteJson(w, http.StatusOK, foundUser)
	return nil
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) *APIError {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return NewAPIError(http.StatusUnauthorized, errors.New("unauthorized"))
	}

	err = h.Redis.Del(r.Context(), cookie.Value).Err()
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, errors.New("failed to delete session"))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	})

	util.WriteJson(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
	return nil
}

func (h *authHandler) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			util.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		sessionDataJSON, err := h.Redis.Get(r.Context(), cookie.Value).Result()
		if err != nil {
			util.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired session"})
			return
		}

		var sessionData config.SessionData
		err = json.Unmarshal([]byte(sessionDataJSON), &sessionData)
		if err != nil {
			util.WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to decode session data"})
			return
		}

		if sessionData.ExpiresAt.Before(time.Now()) {
			util.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, sessionData.UserId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
