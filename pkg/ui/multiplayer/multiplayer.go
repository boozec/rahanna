package multiplayer

import (
	"time"

	"github.com/boozec/rahanna/internal/network"
	"go.uber.org/zap"
)

type GameNetwork struct {
	server *network.TCPNetwork
	me     network.NetworkID
	peer   network.NetworkID
}

// Wrapper to a `TCPNetwork`
func NewGameNetwork(localID string, address string, onHandshake network.NetworkHandshakeFunc, logger *zap.Logger) *GameNetwork {
	opts := network.TCPNetworkOpts{
		ListenAddr:  address,
		HandshakeFn: onHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      logger,
	}
	server := network.NewTCPNetwork(network.NetworkID(localID), opts)
	return &GameNetwork{
		server: server,
		me:     network.NetworkID(localID),
	}
}

func (n *GameNetwork) Peer() network.NetworkID {
	return n.peer
}

func (n *GameNetwork) Me() network.NetworkID {
	return n.me
}

func (n *GameNetwork) Send(payload []byte) error {
	return n.server.Send(n.peer, payload)
}

func (n *GameNetwork) AddPeer(remoteID network.NetworkID, addr string) {
	n.peer = remoteID
	n.server.AddPeer(remoteID, addr)
}

func (n *GameNetwork) AddReceiveFunction(f network.NetworkMessageReceiveFunc) {
	n.server.OnReceiveFn = f
}
