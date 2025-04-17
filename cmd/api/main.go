package main

import (
	"net/http"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/api/handlers"
	"github.com/boozec/rahanna/internal/api/middleware"
	"github.com/boozec/rahanna/internal/logger"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	database.InitDb(os.Getenv("DATABASE_URL"))
	log := logger.InitLogger("rahanna.log", false)
	

	r := mux.NewRouter()
	r.HandleFunc("/auth/register", handlers.RegisterUser).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", handlers.LoginUser).Methods(http.MethodPost)
	r.Handle("/play", middleware.AuthMiddleware(http.HandlerFunc(handlers.NewPlay))).Methods(http.MethodPost)
	r.Handle("/play", middleware.AuthMiddleware(http.HandlerFunc(handlers.AllPlay))).Methods(http.MethodGet)
	r.Handle("/play/{id}", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetGameId))).Methods(http.MethodGet)
	r.Handle("/play/{id}/end", middleware.AuthMiddleware(http.HandlerFunc(handlers.EndGame))).Methods(http.MethodPost)
	r.Handle("/enter-game", middleware.AuthMiddleware(http.HandlerFunc(handlers.EnterGame))).Methods(http.MethodPost)

	log.Info("Serving on :8080")
	handler := cors.AllowAll().Handler(r)
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
