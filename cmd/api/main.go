package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/boozec/rahanna/api/database"
	"github.com/boozec/rahanna/api/handlers"
	"github.com/boozec/rahanna/api/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	database.InitDb(os.Getenv("DATABASE_URL"))

	r := mux.NewRouter()
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", handlers.LoginUser).Methods(http.MethodPost)
	r.Handle("/play", middleware.AuthMiddleware(http.HandlerFunc(handlers.NewPlay))).Methods(http.MethodPost)
	r.Handle("/enter-play", middleware.AuthMiddleware(http.HandlerFunc(handlers.EnterPlay))).Methods(http.MethodPost)

	slog.Info("Serving on :8080")
	handler := cors.AllowAll().Handler(r)
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
