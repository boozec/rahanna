package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndValidateJWT(t *testing.T) {
	// Set up the JWT secret for the test.
	os.Setenv("JWT_SECRET", "testsecret")
	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	userID := 123
	tokenString, err := GenerateJWT(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	claims, err := ValidateJWT(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	_, err := ValidateJWT("invalid_token")
	assert.Error(t, err)
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	// Create a token that has already expired.
	claims := &Claims{
		UserID: 123,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	assert.NoError(t, err)

	_, err = ValidateJWT(tokenString)
	assert.Error(t, err)
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	userID := 123
	tokenString, err := GenerateJWT(userID)
	assert.NoError(t, err)

	// Set a different secret for validation.
	wrongKey := []byte("wrongsecret")

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return wrongKey, nil
	})

	assert.Error(t, err)
}
