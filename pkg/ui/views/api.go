package views

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
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

// sendAPIRequest sends an HTTP request to the API with the given parameters
func sendAPIRequest(method, url string, payload []byte, authorization string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorization))

	client := &http.Client{}
	return client.Do(req)
}
