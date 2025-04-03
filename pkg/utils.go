package utils

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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
