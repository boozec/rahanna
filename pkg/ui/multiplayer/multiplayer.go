package multiplayer

import (
	"time"

	"github.com/boozec/rahanna/internal/network"
	"go.uber.org/zap"
)

type GameNetwork struct {
	Server *network.TCPNetwork
	Peer   string
}

func NewGameNetwork(localID string, address string, onHandshake network.NetworkHandshakeFunc, logger *zap.Logger) *GameNetwork {
	opts := network.TCPNetworkOpts{
		ListenAddr:  address,
		HandshakeFn: onHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      logger,
	}
	server := network.NewTCPNetwork(network.NetworkID(localID), opts)
	peer := ""
	return &GameNetwork{
		Server: server,
		Peer:   peer,
	}
}
