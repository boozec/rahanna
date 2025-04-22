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
		RetryDelay:  2 * time.Second,
		Logger:      zap.L(),
	}
	peer1 := NewTCPNetwork("peer-1", peer1Opts)
	defer peer1.Close()
	time.Sleep(3 * time.Second)

	peer1.RegisterHandler(func(msg Message) {
		assert.Equal(t, "Hey from peer-2!", string(msg.Payload))
	})

	// Create a mock of the second peer (peer-2)
	peer2Opts := TCPNetworkOpts{
		ListenAddr:  ":9002",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  2 * time.Second,
		Logger:      zap.L(),
		OnReceiveFn: func(msg Message) {
			assert.Equal(t, "Hey from peer-1!", string(msg.Payload))
		},
	}
	peer2 := NewTCPNetwork("peer-2", peer2Opts)
	defer peer2.Close()
	time.Sleep(3 * time.Second)

	// Start the first peer and add the second peer
	peer1.AddPeer("peer-2", peer2.ListenAddr)
	peer2.AddPeer("peer-1", peer1.ListenAddr)

	// Wait for connections to be established with a timeout
	time.Sleep(5 * time.Second)

	// Send a message from peer-1 to peer-2
	err := peer1.Send("peer-2", []byte("simple-msg"), []byte("Hey from peer-1!"))
	assert.NoError(t, err)

	err = peer2.Send("peer-1", []byte("simple-msg"), []byte("Hey from peer-2!"))
	assert.NoError(t, err)

	// Allow some time for the message to be received and handled
	time.Sleep(2 * time.Second)
}

// TestSendFailure tests if sending a message fails when no connection exists.
func TestSendFailure(t *testing.T) {
	peer1Opts := TCPNetworkOpts{
		ListenAddr:  ":9003",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      zap.L(),
	}
	peer1 := NewTCPNetwork("peer-1", peer1Opts)
	defer peer1.Close()

	// Create a mock of the second peer (peer-2) - but don't add it to peer1
	peer2Opts := TCPNetworkOpts{
		ListenAddr:  ":9004",
		HandshakeFn: DefaultHandshake,
		RetryDelay:  time.Second * 2,
		Logger:      zap.L(),
	}
	peer2 := NewTCPNetwork("peer-2", peer2Opts)
	defer peer2.Close()

	// Attempt to send a message without establishing a connection first
	err := peer1.Send("peer-2", []byte("msg"), []byte("Message without connection"))
	assert.Error(t, err, "Expected error when sending to a non-connected peer")
}
