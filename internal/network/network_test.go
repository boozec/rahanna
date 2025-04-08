package network

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPeerToPeerCommunication tests if two peers can communicate.
func TestPeerToPeerCommunication(t *testing.T) {
	// Create a mock of the first peer (peer-1)
	peer1IP := "127.0.0.1"
	peer1Port := 9001
	peer1 := NewTCPNetwork("peer-1", peer1IP, peer1Port, func() {})

	// Create a mock of the second peer (peer-2)
	peer2IP := "127.0.0.1"
	peer2Port := 9002
	peer2 := NewTCPNetwork("peer-2", peer2IP, peer2Port, func() {})

	// Register a message handler on peer-2 to receive the message from peer-1
	peer2.RegisterHandler("chat", func(msg Message) {
		assert.Equal(t, "peer-1", msg.Source.ID)
		assert.Equal(t, "Hey from peer-1!", string(msg.Payload))
	})

	// Start the first peer and add the second peer
	go peer1.AddPeer("peer-2", peer2IP, peer2Port)
	go peer2.AddPeer("peer-1", peer1IP, peer1Port)

	// Wait for the connections to be established
	// You might need a little more time based on network delay and retry logic
	time.Sleep(5 * time.Second)

	// Send a message from peer-1 to peer-2
	err := peer1.Send("peer-2", "chat", []byte("Hey from peer-1!"))
	assert.NoError(t, err)

	// Allow some time for the message to be received and handled
	time.Sleep(2 * time.Second)
}

// TestSendFailure tests if sending a message fails when no connection exists.
func TestSendFailure(t *testing.T) {
	peer1 := NewTCPNetwork("peer-1", "127.0.0.1", 9001, func() {})
	_ = NewTCPNetwork("peer-2", "127.0.0.1", 9002, func() {})

	// Attempt to send a message without establishing a connection first
	err := peer1.Send("peer-2", "chat", []byte("Message without connection"))
	assert.Error(t, err, "Expected error when sending to a non-connected peer")
}
