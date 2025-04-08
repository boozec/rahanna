package multiplayer

import (
	"github.com/boozec/rahanna/internal/network"
)

type PlayNetwork struct {
	server *network.TCPNetwork
	peer   string
}

func NewPlayNetwork(localID, localIP string, localPort int) *PlayNetwork {
	server := network.NewTCPNetwork(localID, localIP, localPort)
	peer := ""
	return &PlayNetwork{
		server: server,
		peer:   peer,
	}
}
