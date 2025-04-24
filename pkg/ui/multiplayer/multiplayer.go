package multiplayer

import (
	"slices"
	"time"

	"github.com/boozec/rahanna/internal/logger"
	"github.com/boozec/rahanna/pkg/p2p"
	"go.uber.org/zap"
)

type MoveType string

const (
	AbandonGameMessage    MoveType = "abandon"
	MoveGameMessage       MoveType = "new-move"
	RestoreAckGameMessage MoveType = "restore-ack"
	RestoreGameMessage    MoveType = "restore"
)

type GameMove struct {
	Source  p2p.NetworkID `json:"source"`
	Type    []byte        `json:"type"`
	Payload []byte        `json:"payload"`
}

type GameNetwork struct {
	server *p2p.TCPNetwork
	me     p2p.NetworkID
	peers  []p2p.NetworkID
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

func (n *GameNetwork) Peers() []p2p.NetworkID {
	return n.peers
}

func (n *GameNetwork) Me() p2p.NetworkID {
	return n.me
}

// Send a message to all peers
func (n *GameNetwork) SendAll(messageType []byte, payload []byte) error {
	for _, peer := range n.peers {
		n.server.Send(peer, messageType, payload)
	}

	return nil
}

// Send a message to only one peer
func (n *GameNetwork) Send(peer p2p.NetworkID, messageType []byte, payload []byte) error {
	return n.server.Send(peer, messageType, payload)
}

func (n *GameNetwork) AddPeer(remoteID p2p.NetworkID, addr string) {
	if exists := slices.Contains(n.peers, remoteID); !exists {
		n.peers = append(n.peers, remoteID)
	}
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
