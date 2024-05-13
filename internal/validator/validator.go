package validator

import (
	"log"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func Setup() {
  validate = validator.New()

  if err := validate.RegisterValidation("password_strength", passwordStrength); err != nil {
    log.Fatal("Failed to register validation.")
  }
}

func passwordStrength(fl validator.FieldLevel) bool {
  password := fl.Field().String()

  if len(password) < 8 {
    return false
  }

  hasDigit := false
  for _, char := range password {
    if unicode.IsDigit(char) {
      hasDigit = true
      break
    }
  }
  if !hasDigit {
    return false
  }

  hasUpper := false
  for _, char := range password {
    if unicode.IsUpper(char) {
      hasUpper = true
      break
    }
  }

  return hasUpper
}

func ValidateStruct(s interface{}) error {
  return validate.Struct(s)
}