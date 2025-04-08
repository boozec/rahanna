package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/boozec/rahanna/internal/api/auth"
	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/logger"
	"github.com/boozec/rahanna/internal/network"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type NewGameRequest struct {
	IP string `json:"ip"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	log.Info("POST /auth/register")
	var user database.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	if len(user.Password) < 4 {
		JsonError(&w, "password too short")
		return
	}

	var storedUser database.User
	db, _ := database.GetDb()
	result := db.Where("username = ?", user.Username).First(&storedUser)

	if result.Error == nil {
		JsonError(&w, "user with this username already exists")
		return
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		JsonError(&w, err.Error())
		return
	}
	user.Password = string(hashedPassword)

	result = db.Create(&user)
	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	log.Info("POST /auth/login")
	var inputUser database.User
	err := json.NewDecoder(r.Body).Decode(&inputUser)
	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	var storedUser database.User

	db, _ := database.GetDb()
	result := db.Where("username = ?", inputUser.Username).First(&storedUser)
	if result.Error != nil {
		JsonError(&w, "invalid credentials")
		return
	}

	if err := CheckPasswordHash(storedUser.Password, inputUser.Password); err != nil {
		JsonError(&w, "invalid credentials")
		return
	}

	token, err := auth.GenerateJWT(storedUser.ID)
	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func NewPlay(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	log.Info("POST /play")
	claims, err := auth.ValidateJWT(r.Header.Get("Authorization"))

	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	var payload struct {
		IP string `json:"ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		JsonError(&w, err.Error())
		return
	}

	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	db, _ := database.GetDb()

	name := network.NewSession()
	play := database.Game{
		Player1ID: claims.UserID,
		Player2ID: nil,
		Name:      name,
		IP1:       payload.IP,
		IP2:       "",
	}

	result := db.Create(&play)
	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"name": name})
}

func EnterGame(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	log.Info("POST /enter-game")
	claims, err := auth.ValidateJWT(r.Header.Get("Authorization"))

	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	var payload struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		JsonError(&w, err.Error())
		return
	}

	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	db, _ := database.GetDb()

	var game database.Game

	result := db.Where("name = ? AND player2_id IS NULL", payload.Name).First(&game)
	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	game.Player2ID = &claims.UserID
	game.IP2 = payload.IP
	game.UpdatedAt = time.Now()

	if err := db.Save(&game).Error; err != nil {
		JsonError(&w, err.Error())
		return
	}

	result = db.Where("id = ?", game.ID).
		Preload("Player1", func(db *gorm.DB) *gorm.DB {
			return db.Omit("Password")
		}).
		Preload("Player2", func(db *gorm.DB) *gorm.DB {
			return db.Omit("Password")
		}).
		First(&game)

	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(game)
}

func AllPlay(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	log.Info("GET /play")

	claims, err := auth.ValidateJWT(r.Header.Get("Authorization"))

	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	db, _ := database.GetDb()
	var games []database.Game

	result := db.Where("player1_id = ? OR player2_id = ?", claims.UserID, claims.UserID).
		Preload("Player1", func(db *gorm.DB) *gorm.DB {
			return db.Omit("Password")
		}).
		Preload("Player2", func(db *gorm.DB) *gorm.DB {
			return db.Omit("Password")
		}).
		Order("updated_at DESC").
		Find(&games)

	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(games)
}

func GetGameId(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	vars := mux.Vars(r)
	id := vars["id"]
	log.Info(fmt.Sprintf("GET /play/%s", id))

	claims, err := auth.ValidateJWT(r.Header.Get("Authorization"))

	if err != nil {
		JsonError(&w, err.Error())
		return
	}

	db, _ := database.GetDb()
	var game database.Game

	result := db.Where("id = ? AND (player1_id = ? OR player2_id = ?)", id, claims.UserID, claims.UserID).
		Preload("Player1", func(db *gorm.DB) *gorm.DB {
			return db.Omit("Password")
		}).
		Preload("Player2", func(db *gorm.DB) *gorm.DB {
			return db.Omit("Password")
		}).
		First(&game)

	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(game)
}
