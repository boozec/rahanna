package main

import (
	"os"

	"github.com/boozec/rahanna/api/database"
	"github.com/boozec/rahanna/api/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	database.InitDb(os.Getenv("DATABASE_URL"))

	r := mux.NewRouter()
	r.HandleFunc("/register", handlers.RegisterUser).Methods(http.MethodPost)
	r.HandleFunc("/login", handlers.LoginUser).Methods(http.MethodPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
