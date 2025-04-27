package database

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GameType string

const (
	SingleGameType GameType = "single"
	PairGameType   GameType = "pair"
)

type MoveChooseType string

const (
	SequentialChooseType MoveChooseType = "sequential"
	RandomChooseType     MoveChooseType = "random"
)

type Game struct {
	ID         int            `json:"id"`
	Type       GameType       `json:"type"`
	MoveChoose MoveChooseType `json:"move_choose_type"`
	Player1ID  int            `json:"-"`
	Player1    User           `gorm:"foreignKey:Player1ID" json:"player1"`
	Player2ID  *int           `json:"-"`
	Player2    *User          `gorm:"foreignKey:Player2ID;null" json:"player2"`
	Player3ID  *int           `json:"-"`
	Player3    *User          `gorm:"foreignKey:Player3ID;null" json:"player3"`
	Player4ID  *int           `json:"-"`
	Player4    *User          `gorm:"foreignKey:Player4ID;null" json:"player4"`
	Name       string         `json:"name"`
	IP1        string         `json:"ip1"`
	IP2        string         `json:"ip2"`
	IP3        string         `json:"ip3"`
	IP4        string         `json:"ip4"`
	Outcome    string         `json:"outcome"`
	LastPlayer int            `json:"last_player"` // Last player entered in game
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
