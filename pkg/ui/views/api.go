package views

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/boozec/rahanna/internal/api/auth"
)

// getAuthorizationToken reads the authentication token from the .rahannarc file
func getAuthorizationToken() (string, error) {
	f, err := os.Open(".rahannarc")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var authorization string
	for scanner.Scan() {
		authorization = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading auth token: %v", err)
	}

	return authorization, nil
}

// From a JWT token it returns the associated user ID
func getUserID() (int, error) {
	token, err := getAuthorizationToken()
	if err != nil {
		return -1, err
	}

	claims, err := auth.ValidateJWT(token)
	if err != nil {
		return -1, err
	}

	return claims.UserID, nil

}

// sendAPIRequest sends an HTTP request to the API with the given parameters
func sendAPIRequest(method, url string, payload []byte, authorization string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", authorization)

	client := &http.Client{}
	return client.Do(req)
}
