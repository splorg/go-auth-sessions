package auth

import (
	"time"

	"github.com/google/uuid"
)

type SessionData struct {
	UserId    uuid.UUID `json:"userId"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
}
