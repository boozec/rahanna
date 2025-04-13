package multiplayer

import (
	"github.com/boozec/rahanna/internal/network"
)

type GameNetwork struct {
	Server *network.TCPNetwork
	Peer   string
}

func NewGameNetwork(localID, localIP string, localPort int, callback func()) *GameNetwork {
	server := network.NewTCPNetwork(localID, localIP, localPort, callback)
	peer := ""
	return &GameNetwork{
		Server: server,
		Peer:   peer,
	}
}
