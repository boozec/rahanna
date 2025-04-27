package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/boozec/rahanna/internal/api/auth"
	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/logger"
	"github.com/boozec/rahanna/pkg/p2p"
	"github.com/gorilla/mux"
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
	if result := db.Where("username = ?", user.Username).First(&storedUser); result.Error == nil {
		JsonError(&w, "user with this username already exists")
		return
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		JsonError(&w, err.Error())
		return
	}
	user.Password = string(hashedPassword)

	if result := db.Create(&user); result.Error != nil {
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
	if result := db.Where("username = ?", inputUser.Username).First(&storedUser); result.Error != nil {
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

	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok {
		JsonError(&w, "claims not found")
		return
	}

	var payload struct {
		IP         string                  `json:"ip"`
		Type       database.GameType       `json:"type"`
		MoveChoose database.MoveChooseType `json:"move_choose_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		JsonError(&w, err.Error())
		return
	}

	db, _ := database.GetDb()

	var name string
	for {
		name = p2p.NewSession()
		if err := db.Where("name = ? AND outcome = '*'", name).First(nil).Error; err != nil {
			break
		}
	}

	play := database.Game{
		Type:       payload.Type,
		MoveChoose: payload.MoveChoose,
		Player1ID:  claims.UserID,
		Name:       name,
		IP1:        payload.IP,
		Outcome:    "*",
		LastPlayer: 1,
	}

	if result := db.Create(&play); result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": play.ID, "type": play.Type, "moove_choose_type": play.MoveChoose, "name": name,
	})
}

func EnterGame(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	log.Info("POST /enter-game")

	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok {
		JsonError(&w, "claims not found")
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

	db, _ := database.GetDb()

	var game database.Game

	if result := db.Where("name = ?", payload.Name).First(&game); result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	switch game.Type {
	case database.SingleGameType:
		if game.Player2ID == nil {
			game.Player2ID = &claims.UserID
			game.IP2 = payload.IP
			game.LastPlayer = 2
		} else {
			switch claims.UserID {
			case game.Player1ID:
				game.IP1 = payload.IP
				game.LastPlayer = 1
			case *game.Player2ID:
				game.IP2 = payload.IP
				game.LastPlayer = 2
			}
		}

	case database.PairGameType:
		if game.Player2ID == nil {
			game.Player2ID = &claims.UserID
			game.IP2 = payload.IP
			game.LastPlayer = 2
		} else if game.Player3ID == nil {
			game.Player3ID = &claims.UserID
			game.IP3 = payload.IP
			game.LastPlayer = 3
		} else if game.Player4ID == nil {
			game.Player4ID = &claims.UserID
			game.IP4 = payload.IP
			game.LastPlayer = 4
		} else {
			switch claims.UserID {
			case game.Player1ID:
				game.IP1 = payload.IP
				game.LastPlayer = 1
			case *game.Player2ID:
				game.IP2 = payload.IP
				game.LastPlayer = 2
			case *game.Player3ID:
				game.IP3 = payload.IP
				game.LastPlayer = 3
			case *game.Player4ID:
				game.IP4 = payload.IP
				game.LastPlayer = 4
			}
		}

	default:
		log.Fatal("Game type not recognized")
	}

	game.UpdatedAt = time.Now()

	if err := db.Save(&game).Error; err != nil {
		JsonError(&w, err.Error())
		return
	}

	result := db.Where("id = ?", game.ID).
		Preload("Player1", auth.OmitPassword).
		Preload("Player2", auth.OmitPassword).
		Preload("Player3", auth.OmitPassword).
		Preload("Player4", auth.OmitPassword).
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

	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok {
		JsonError(&w, "claims not found")
		return
	}

	db, _ := database.GetDb()
	var games []database.Game

	if result := db.Where("player1_id = ? OR player2_id = ? OR player3_id = ? OR player4_id = ?",
		claims.UserID, claims.UserID, claims.UserID, claims.UserID,
	).
		Preload("Player1", auth.OmitPassword).
		Preload("Player2", auth.OmitPassword).
		Preload("Player3", auth.OmitPassword).
		Preload("Player4", auth.OmitPassword).
		Order("updated_at DESC").
		Find(&games); result.Error != nil {
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

	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok {
		JsonError(&w, "claims not found")
		return
	}

	db, _ := database.GetDb()
	var game database.Game

	if result := db.Where("id = ? AND (player1_id = ? OR player2_id = ? OR player3_id = ? OR player4_id = ?)",
		id, claims.UserID, claims.UserID, claims.UserID, claims.UserID).
		Preload("Player1", auth.OmitPassword).
		Preload("Player2", auth.OmitPassword).
		Preload("Player3", auth.OmitPassword).
		Preload("Player4", auth.OmitPassword).
		First(&game); result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(game)
}

func EndGame(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	vars := mux.Vars(r)
	id := vars["id"]
	log.Info(fmt.Sprintf("POST /play/%s/end", id))

	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok {
		JsonError(&w, "claims not found")
		return
	}

	var payload struct {
		Outcome string `json:"outcome"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		JsonError(&w, err.Error())
		return
	}

	db, _ := database.GetDb()

	var game database.Game

	// FIXME: this is not secure
	if result := db.Where(
		"id = ? AND (player1_id = ? OR player2_id = ? OR player3_id = ? OR player4_id = ?)",
		id, claims.UserID, claims.UserID, claims.UserID, claims.UserID,
	).First(&game); result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	game.Outcome = payload.Outcome

	if err := db.Save(&game).Error; err != nil {
		JsonError(&w, err.Error())
		return
	}

	result := db.Where("id = ?", game.ID).
		Preload("Player1", auth.OmitPassword).
		Preload("Player2", auth.OmitPassword).
		Preload("Player3", auth.OmitPassword).
		Preload("Player4", auth.OmitPassword).
		First(&game)

	if result.Error != nil {
		JsonError(&w, result.Error.Error())
		return
	}

	json.NewEncoder(w).Encode(game)
}
