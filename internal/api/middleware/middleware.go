package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/boozec/rahanna/internal/api/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		payloadMap := map[string]string{"error": "unauthorized"}
		payload, _ := json.Marshal(payloadMap)

		if tokenString == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			w.Write([]byte(payload))
			return
		}

		_, err := auth.ValidateJWT(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)

			payload, _ := json.Marshal(payloadMap)

			w.Write([]byte(payload))
			return
		}
		next.ServeHTTP(w, r)
	})
}
