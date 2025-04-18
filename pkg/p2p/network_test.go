package p2p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestPeerToPeerCommunication tests if two peers can communicate.
func TestPeerToPeerCommunication(t *testing.T) {
	// Create a mock of the first peer (peer-1)
	peer1Opts := TCPNetworkOpts{
		ListenAddr:  ":9001",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      zap.L(),
	}
	peer1 := NewTCPNetwork("peer-1", peer1Opts)

	peer1.RegisterHandler(func(msg Message) {
		assert.Equal(t, "Hey from peer-2!", string(msg.Payload))
	})

	// Create a mock of the second peer (peer-2)
	peer2Opts := TCPNetworkOpts{
		ListenAddr:  ":9002",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      zap.L(),
		OnReceiveFn: func(msg Message) {
			assert.Equal(t, "Hey from peer-1!", string(msg.Payload))
		},
	}
	peer2 := NewTCPNetwork("peer-2", peer2Opts)

	// Start the first peer and add the second peer
	go peer1.AddPeer("peer-2", peer2.ListenAddr)
	go peer2.AddPeer("peer-1", peer1.ListenAddr)

	// Wait for the connections to be established
	// You might need a little more time based on network delay and retry logic
	time.Sleep(5 * time.Second)

	// Send a message from peer-1 to peer-2
	err := peer1.Send("peer-2", []byte("Hey from peer-1!"))
	assert.NoError(t, err)

	err = peer2.Send("peer-1", []byte("Hey from peer-2!"))
	assert.NoError(t, err)

	// Allow some time for the message to be received and handled
	time.Sleep(2 * time.Second)
}

// TestSendFailure tests if sending a message fails when no connection exists.
func TestSendFailure(t *testing.T) {
	peer1Opts := TCPNetworkOpts{
		ListenAddr:  ":9001",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      zap.L(),
	}
	peer1 := NewTCPNetwork("peer-1", peer1Opts)

	// Create a mock of the second peer (peer-2)
	peer2Opts := TCPNetworkOpts{
		ListenAddr:  ":9002",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      zap.L(),
	}
	_ = NewTCPNetwork("peer-2", peer2Opts)

	// Attempt to send a message without establishing a connection first
	err := peer1.Send("peer-2", []byte("Message without connection"))
	assert.Error(t, err, "Expected error when sending to a non-connected peer")
}
