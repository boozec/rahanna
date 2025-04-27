package views

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/pkg/p2p"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/notnil/chess"
)

func (m GameModel) handleDatabaseGameMsg(msg database.Game) (GameModel, tea.Cmd) {
	m.game = &msg

	var cmd tea.Cmd

	peers := map[int]string{
		1: m.game.IP1,
		2: m.game.IP2,
	}

	if m.game.Type == database.PairGameType {
		peers[3] = m.game.IP3
		peers[4] = m.game.IP4
	}

	myPlayerNum := -1
	switch m.network.Me() {
	case m.playerPeer(1):
		myPlayerNum = 1
	case m.playerPeer(2):
		myPlayerNum = 2
	case m.playerPeer(3):
		myPlayerNum = 3
	case m.playerPeer(4):
		myPlayerNum = 4
	}

	// Add all peers to every other peer
	for playerNum, ip := range peers {
		if playerNum != myPlayerNum && ip != "" {
			m.network.AddPeer(m.playerPeer(playerNum), ip)
		}
	}

	if myPlayerNum == 1 && m.turn == p2p.EmptyNetworkID {
		// FIXME: use another way instead of sleep
		time.Sleep(2 * time.Second)
		if m.game.MoveChoose == database.RandomChooseType {
			players := []int{1, 3}
			m.turn = m.playerPeer(players[rand.Intn(len(players))])
		} else {
			m.turn = m.playerPeer(1)
		}
		m.network.SendAll([]byte("define-turn"), []byte(string(m.turn)))

	}

	if m.restore {
		cmd = func() tea.Msg {
			return RestoreGameMsg{}
		}
	} else if m.game.Outcome != chess.NoOutcome.String() {
		cmd = func() tea.Msg {
			return EndGameMsg{}
		}
	}

	return m, cmd
}

func (m *GameModel) getGame() tea.Cmd {
	return func() tea.Msg {
		var game database.Game

		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return nil
		}

		// Send API request
		url := fmt.Sprintf("%s/play/%d", os.Getenv("API_BASE"), m.currentGameID)
		resp, err := sendAPIRequest("GET", url, nil, authorization)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
			return nil
		}

		m.game = &game

		return game
	}
}

type EndGameMsg struct {
	abandoned bool
}

type RestoreGameMsg struct{}

func (m *GameModel) endGame(outcome string, abandon bool) tea.Cmd {
	return func() tea.Msg {
		var game database.Game

		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return err
		}

		// Prepare request payload
		payload, err := json.Marshal(map[string]string{
			"outcome": outcome,
		})

		// Send API request
		url := fmt.Sprintf("%s/play/%d/end", os.Getenv("API_BASE"), m.currentGameID)
		resp, err := sendAPIRequest("POST", url, payload, authorization)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&game); err != nil {
			return err
		}

		if abandon {
			m.network.SendAll([]byte("abandon"), []byte("🏳️"))
		}

		return game
	}
}
