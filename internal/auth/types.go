package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expire"`
}

type SessionData struct {
	Token  string    `json:"token"`
	UserId uuid.UUID `json:"userId"`
}
