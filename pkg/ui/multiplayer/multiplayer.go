package multiplayer

import (
	"time"

	"github.com/boozec/rahanna/internal/logger"
	"github.com/boozec/rahanna/pkg/p2p"
	"go.uber.org/zap"
)

type MoveType string

const (
	AbandonGameMessage    MoveType = "abandon"
	RestoreGameMessage    MoveType = "restore"
	RestoreAckGameMessage MoveType = "restore-ack"
	MoveGameMessage       MoveType = "new-move"
)

type GameMove struct {
	Type    []byte `json:"type"`
	Payload []byte `json:"payload"`
}

type GameNetwork struct {
	server *p2p.TCPNetwork
	me     p2p.NetworkID
	peer   p2p.NetworkID
}

// Wrapper to a `TCPNetwork`
func NewGameNetwork(localID string, address string, onHandshake p2p.NetworkHandshakeFunc, onFirstHandshake p2p.NetworkHandshakeFunc, logger *zap.Logger) *GameNetwork {
	opts := p2p.TCPNetworkOpts{
		ListenAddr:       address,
		HandshakeFn:      onHandshake,
		FirstHandshakeFn: onFirstHandshake,
		RetryDelay:       time.Second * 2,
		Logger:           logger,
	}
	server := p2p.NewTCPNetwork(p2p.NetworkID(localID), opts)
	return &GameNetwork{
		server: server,
		me:     p2p.NetworkID(localID),
	}
}

func (n *GameNetwork) Peer() p2p.NetworkID {
	return n.peer
}

func (n *GameNetwork) Me() p2p.NetworkID {
	return n.me
}

func (n *GameNetwork) Send(messageType []byte, payload []byte) error {
	return n.server.Send(n.peer, messageType, payload)
}

func (n *GameNetwork) AddPeer(remoteID p2p.NetworkID, addr string) {
	n.peer = remoteID
	n.server.AddPeer(remoteID, addr)
}

func (n *GameNetwork) AddReceiveFunction(f p2p.NetworkMessageReceiveFunc) {
	n.server.OnReceiveFn = f
}

func (n *GameNetwork) Close() error {
	err := n.server.Close()
	logger, _ := logger.GetLogger()

	if err != nil {
		logger.Sugar().Errorf("can't close connection for network '%+v': %s", n, err.Error())
	} else {
		logger.Sugar().Infof("connection closed for network '%+v'", n)
	}

	return err
}
