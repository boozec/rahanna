package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/boozec/rahanna/api/auth"
	"github.com/boozec/rahanna/api/database"
	utils "github.com/boozec/rahanna/pkg"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("POST /register")
	var user database.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.JsonError(&w, err.Error())
		return
	}

	if len(user.Password) < 4 {
		utils.JsonError(&w, "password too short")
		return
	}

	var storedUser database.User
	db, _ := database.GetDb()
	result := db.Where("username = ?", user.Username).First(&storedUser)

	if result.Error == nil {
		utils.JsonError(&w, "user with this username already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.JsonError(&w, err.Error())
		return
	}
	user.Password = string(hashedPassword)

	result = db.Create(&user)
	if result.Error != nil {
		utils.JsonError(&w, result.Error.Error())
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		utils.JsonError(&w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("POST /login")
	var inputUser database.User
	err := json.NewDecoder(r.Body).Decode(&inputUser)
	if err != nil {
		utils.JsonError(&w, err.Error())
		return
	}

	var storedUser database.User

	db, _ := database.GetDb()
	result := db.Where("username = ?", inputUser.Username).First(&storedUser)
	if result.Error != nil {
		utils.JsonError(&w, "invalid credentials")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(inputUser.Password))
	if err != nil {
		utils.JsonError(&w, "invalid credentials")
		return
	}

	token, err := auth.GenerateJWT(storedUser.ID)
	if err != nil {
		utils.JsonError(&w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
