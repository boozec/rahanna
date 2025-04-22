package views

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/boozec/rahanna/internal/api/database"
	"github.com/boozec/rahanna/internal/logger"
	"github.com/boozec/rahanna/pkg/p2p"
	"github.com/boozec/rahanna/pkg/ui/multiplayer"
	tea "github.com/charmbracelet/bubbletea"
)

type responseOk struct {
	Name   string `json:"name"`
	GameID int    `json:"id"`
	IP     string `json:"ip"`
	Port   int    `json:"int"`
}

// API response types
type playResponse struct {
	Ok    responseOk
	Error string `json:"error"`
}

type StartGameMsg struct{}

func (m *PlayModel) handlePlayResponse(msg playResponse) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.err = nil

	if msg.Error != "" {
		m.err = fmt.Errorf("%s", msg.Error)
		if msg.Error == "unauthorized" {
			return m, logout(m.width, m.height+1)
		}
	} else {
		m.playName = msg.Ok.Name
		m.currentGameId = msg.Ok.GameID
		logger, _ := logger.GetLogger()

		callbackCompleted := make(chan bool)
		m.network = multiplayer.NewGameNetwork(fmt.Sprintf("%s-1", m.playName), fmt.Sprintf("%s:%d", msg.Ok.IP, msg.Ok.Port), p2p.DefaultHandshake, func(net.Conn) error {
			close(callbackCompleted)
			return nil
		}, logger)

		return m, func() tea.Msg {
			<-callbackCompleted
			return StartGameMsg{}
		}
	}

	return m, nil
}

func (m *PlayModel) handleGameResponse(msg database.Game) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.game = &msg
	m.err = nil

	var ip []string
	var localID string

	if m.game.LastPlayer == 2 {
		ip = strings.Split(m.game.IP2, ":")
		localID = fmt.Sprintf("%s-2", m.game.Name)
	} else {
		ip = strings.Split(m.game.IP1, ":")
		localID = fmt.Sprintf("%s-1", m.game.Name)
	}

	if len(ip) == 2 {
		localIP := ip[0]
		localPort, _ := strconv.ParseInt(ip[1], 10, 32)

		logger, _ := logger.GetLogger()
		network := multiplayer.NewGameNetwork(localID, fmt.Sprintf("%s:%d", localIP, localPort), p2p.DefaultHandshake, p2p.DefaultHandshake, logger)

		return m, SwitchModelCmd(NewGameModel(m.width, m.height+1, m.game.ID, network, m.gameToRestore != nil))
	}
	return m, nil
}

func (m *PlayModel) handleGamesResponse(msg []database.Game) (tea.Model, tea.Cmd) {
	m.isLoading = false
	m.games = msg
	m.err = nil
	m.paginator.SetTotalPages(len(m.games))
	return m, nil
}

func (m *PlayModel) newGameCallback() tea.Cmd {
	return func() tea.Msg {
		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Set up network connection
		port, err := p2p.GetRandomAvailablePort()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		ip := p2p.GetOutboundIP().String()

		// Prepare request payload
		payload, err := json.Marshal(map[string]string{
			"ip": fmt.Sprintf("%s:%d", ip, port),
		})
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Send API request
		url := os.Getenv("API_BASE") + "/play"
		resp, err := sendAPIRequest("POST", url, payload, authorization)
		if err != nil {
			return playResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		// Handle response
		if resp.StatusCode != http.StatusOK {
			var response playResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return playResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return playResponse{Error: response.Error}
		}

		// Decode successful response
		var response struct {
			Name  string `json:"name"`
			ID    int    `json:"id"`
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return playResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return playResponse{Ok: responseOk{Name: response.Name, GameID: response.ID, IP: ip, Port: port}}
	}
}

func (m PlayModel) enterGame() tea.Cmd {
	return func() tea.Msg {
		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Set up network connection
		port, err := p2p.GetRandomAvailablePort()
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		ip := p2p.GetOutboundIP().String()

		// Prepare request payload
		payload, err := json.Marshal(map[string]string{
			"ip":   fmt.Sprintf("%s:%d", ip, port),
			"name": m.namePrompt.Value(),
		})
		if err != nil {
			return playResponse{Error: err.Error()}
		}

		// Send API request
		url := os.Getenv("API_BASE") + "/enter-game"
		resp, err := sendAPIRequest("POST", url, payload, authorization)
		if err != nil {
			return playResponse{Error: err.Error()}
		}
		defer resp.Body.Close()

		// Handle response
		if resp.StatusCode != http.StatusOK {
			var response playResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return playResponse{Error: fmt.Sprintf("HTTP error: %d, unable to decode body", resp.StatusCode)}
			}
			return playResponse{Error: response.Error}
		}

		// Decode successful response
		var response database.Game
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return playResponse{Error: fmt.Sprintf("Error decoding JSON: %v", err)}
		}

		return response
	}
}

func (m *PlayModel) fetchGames() tea.Cmd {
	return func() tea.Msg {
		var games []database.Game
		// Get authorization token
		authorization, err := getAuthorizationToken()
		if err != nil {
			return games
		}

		// Send API request
		url := os.Getenv("API_BASE") + "/play"
		resp, err := sendAPIRequest("GET", url, nil, authorization)
		if err != nil {
			return games
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
			return []database.Game{}
		}

		return games
	}
}
