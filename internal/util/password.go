package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(password []byte) ([]byte, error) {
  return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func ComparePassword(hashedPassword []byte, password []byte) error {
  return bcrypt.CompareHashAndPassword(hashedPassword, password)
}