package views

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/notnil/chess"
)

func (m GameModel) handleDatabaseGameMsg(msg database.Game) (GameModel, tea.Cmd) {
	m.game = &msg

	var cmd tea.Cmd

	peers := map[int]string{
		1: m.game.IP1,
		2: m.game.IP2,
		3: m.game.IP3,
		4: m.game.IP4,
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

func (m *GameModel) endGame(outcome string) tea.Cmd {
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

		return game
	}
}
