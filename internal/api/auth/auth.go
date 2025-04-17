package auth

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Key used for JWT encryption/decryption
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Kind of JWT token
var TokenType = "Bearer"

// Extends JWT Claims with the UserID field
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// Generate a JWT token from an userID.
func GenerateJWT(userID int) (string, error) {
	// Set expiration date for the token to 90 days
	expirationTime := time.Now().Add(90 * 24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return TokenType + " " + tokenString, nil
}

// Validate a JWT token for a kind of time
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	tokenParts := strings.Split(tokenString, " ")
	if len(tokenParts) != 2 {
		return nil, errors.New("not valid JWT")
	}

	if tokenParts[0] != TokenType {
		return nil, errors.New("not valid JWT type")
	}

	token, err := jwt.ParseWithClaims(tokenParts[1], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}

// Common omit password field for users
func OmitPassword(db *gorm.DB) *gorm.DB {
	return db.Omit("Password")
}
