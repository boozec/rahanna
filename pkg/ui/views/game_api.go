package views

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boozec/rahanna/internal/api/database"
	tea "github.com/charmbracelet/bubbletea"
)

func (m GameModel) handleDatabaseGameMsg(msg database.Game) GameModel {
	m.game = &msg
	if m.peer == "peer-2" {
		m.network.Peer = msg.IP2
	} else {
		m.network.Peer = msg.IP1
	}
	return m
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

		// Establish peer connection
		if m.peer == "peer-2" {
			if game.IP2 != "" {
				remote := game.IP2
				go m.network.Server.AddPeer("peer-2", remote)
			}
		} else {
			if game.IP1 != "" {
				remote := game.IP1
				go m.network.Server.AddPeer("peer-1", remote)
			}
		}

		return game
	}
}
