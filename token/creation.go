package token

import (
	"time"

	"atwell.dev/reelingit/logger"
	"atwell.dev/reelingit/models"
	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(user models.User, logger logger.Logger) string {
	jwtSecret := GetJWTSecret(logger)

	// Create a JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	})

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Error("Failed to sign JWT", err)
		return ""
	}

	return tokenString
}