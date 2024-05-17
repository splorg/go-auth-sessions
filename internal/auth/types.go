package auth

import (
	"time"

	"github.com/google/uuid"
)

type SessionData struct {
	UserId    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expiresAt"`
}
