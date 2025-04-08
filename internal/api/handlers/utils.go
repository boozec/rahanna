package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Set a JSON response with status code 400
func JsonError(w *http.ResponseWriter, error string) {
	payloadMap := map[string]string{"error": error}

	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusBadRequest)

	payload, err := json.Marshal(payloadMap)

	if err != nil {
		(*w).WriteHeader(http.StatusBadGateway)
		(*w).Write([]byte(err.Error()))
	} else {
		(*w).Write(payload)
	}
}
