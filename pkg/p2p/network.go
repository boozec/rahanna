package p2p

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

// `Message` represents a structured message on this network.
type Message struct {
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"`
	Payload   []byte `json:"payload"`
}

// A network ID is represented by a string
type NetworkID string

// Default empty network' ID
const EmptyNetworkID NetworkID = NetworkID("")

// This type represents the function that is called every time a new message
// arrives to the server.
type NetworkMessageReceiveFunc func(msg Message)

// This type represent the callback function invokes every new handshake between
// two peers
type NetworkHandshakeFunc func() error

func DefaultHandshake() error {
	return nil
}

// Network options to define on new `TCPNetwork`
type TCPNetworkOpts struct {
	ListenAddr  string
	RetryDelay  time.Duration
	HandshakeFn NetworkHandshakeFunc
	OnReceiveFn NetworkMessageReceiveFunc
	Logger      *zap.Logger
}

// TCPNetwork represents a full-duplex TCP peer.
type TCPNetwork struct {
	sync.Mutex
	TCPNetworkOpts

	id          NetworkID
	listener    net.Listener
	connections map[NetworkID]net.Conn
	isClosed    bool
}

// Initiliaze a new TCP network
func NewTCPNetwork(localID NetworkID, opts TCPNetworkOpts) *TCPNetwork {
	n := &TCPNetwork{
		TCPNetworkOpts: opts,
		id:             localID,
		connections:    make(map[NetworkID]net.Conn),
	}

	go n.startServer()

	return n
}

// Close listener' connection
func (n *TCPNetwork) Close() error {
	return n.listener.Close()
}

// Add a new peer connection to the local peer
func (n *TCPNetwork) AddPeer(remoteID NetworkID, addr string) {
	go n.retryConnect(remoteID, addr)
}

// Send methods is used to send a message to a specified remote peer
func (n *TCPNetwork) Send(remoteID NetworkID, payload []byte) error {
	n.Lock()
	conn, exists := n.connections[remoteID]
	n.Unlock()

	if !exists {
		return fmt.Errorf("not connected to peer %s", remoteID)
	}

	message := Message{
		Payload:   payload,
		Source:    n.listener.Addr().String(),
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	_, err = conn.Write(append(data, '\n'))
	if err != nil {
		n.Logger.Sugar().Errorf("failed to send message to %s: %v. Reconnecting...", remoteID, err)
		n.Lock()
		delete(n.connections, remoteID)
		n.Unlock()

		go n.retryConnect(remoteID, "")

		return fmt.Errorf("failed to send message: %v", err)
	} else {
		n.Logger.Sugar().Infof("Sent message to '%s': %s", conn.LocalAddr(), message.Payload)
	}

	return nil
}

// RegisterHandler registers a callback for a message type.
func (n *TCPNetwork) RegisterHandler(callback NetworkMessageReceiveFunc) {
	n.OnReceiveFn = callback
}

// startServer starts a TCP server to accept connections.
func (n *TCPNetwork) startServer() error {
	var err error

	n.listener, err = net.Listen("tcp", n.ListenAddr)
	if err != nil {
		n.Logger.Sugar().Errorf("failed to start server: %v", err)
		return err
	}

	n.isClosed = false

	go n.listenLoop()

	n.Logger.Sugar().Infof("server started on %s\n", n.ListenAddr)

	return nil

}

func (n *TCPNetwork) listenLoop() error {
	for {
		conn, err := n.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			n.isClosed = true
			n.Logger.Sugar().Errorf("connection is closed in such a way: %v\n", err)
			return err
		}

		if err != nil {
			n.Logger.Sugar().Errorf("failed to accept connection: %v\n", err)
			continue
		}

		remoteAddr := conn.RemoteAddr().String()
		n.Lock()
		n.connections[NetworkID(remoteAddr)] = conn
		if err := n.HandshakeFn(); err != nil {
			n.Logger.Sugar().Errorf("error on handshaking: %v\n", err)
			return err
		}
		n.Unlock()
		n.RetryDelay = 2 * time.Second

		n.Logger.Sugar().Infof("connected to remote peer %s\n", remoteAddr)

		// Read loop
		go n.listenForMessages(conn)
	}
}

// listenForMessages listens for incoming messages.
func (n *TCPNetwork) listenForMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			n.Logger.Debug("connection lost. Reconnecting...")
			n.Lock()

			// FIXME: a better way to re-establish the connection between peer
			for id, c := range n.connections {
				if c == conn {
					delete(n.connections, id)
					go n.retryConnect(id, "")
					break
				}
			}
			n.Unlock()
			return
		}

		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			n.Logger.Sugar().Errorf("failed to unmarshal message: %v\n", err)
			continue
		}

		n.Logger.Sugar().Infof("Received message from '%s': %s", message.Source, string(message.Payload))

		n.OnReceiveFn(message)
	}
}

// retryConnect attempts to connect to a remote peer.
func (n *TCPNetwork) retryConnect(remoteID NetworkID, addr string) {
	for {
		n.Lock()
		_, exists := n.connections[remoteID]
		n.Unlock()

		if exists {
			time.Sleep(5 * time.Second)
			continue
		}

		conn, err := net.Dial("tcp", addr)

		if err != nil {
			n.Logger.Sugar().Errorf("failed to connect to %s: %v. Retrying in %v...", remoteID, err, n.RetryDelay)
			time.Sleep(n.RetryDelay)
			if !n.isClosed && n.RetryDelay < 30*time.Second {
				n.RetryDelay *= 2
			} else {
				n.Lock()
				delete(n.connections, remoteID)
				n.Unlock()
				n.Logger.Sugar().Infof("removed %s connection", remoteID)
				return
			}
			continue
		}

		n.Lock()
		n.connections[remoteID] = conn
		n.Unlock()
		n.Logger.Sugar().Infof("successfully connected to peer %s!", remoteID)

		go n.listenForMessages(conn)
	}
}
