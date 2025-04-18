package database

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Game struct {
	ID         int       `json:"id"`
	Player1ID  int       `json:"-"`
	Player1    User      `gorm:"foreignKey:Player1ID" json:"player1"`
	Player2ID  *int      `json:"-"`
	Player2    *User     `gorm:"foreignKey:Player2ID;null" json:"player2"`
	Name       string    `json:"name"`
	IP1        string    `json:"ip1"`
	IP2        string    `json:"ip2"`
	Outcome    string    `json:"outcome"`
	LastPlayer int       `json:"last_player"` // Last player entered in game
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
