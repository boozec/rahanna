package views

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/pkg/p2p"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/notnil/chess"
)

func (m GameModel) handleDatabaseGameMsg(msg database.Game) (GameModel, tea.Cmd) {
	m.game = &msg

	var cmd tea.Cmd

	// Establish peer connection
	if m.network.Peer() == p2p.EmptyNetworkID {
		if m.network.Me() == m.playerPeer(1) {
			if m.game.IP2 != "" {
				remote := m.game.IP2
				go m.network.AddPeer(m.playerPeer(2), remote)
			}
		} else {
			if m.game.IP1 != "" {
				remote := m.game.IP1
				go m.network.AddPeer(m.playerPeer(1), remote)
			}
		}
	}

	if m.game.Outcome != chess.NoOutcome.String() {
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

		return game
	}
}

type EndGameMsg struct {
	abandoned bool
}

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
