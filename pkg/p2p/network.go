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

// TCPNetwork represents a TCP peer capable to send and receive messages
type TCPNetwork struct {
	sync.Mutex
	TCPNetworkOpts

	id            NetworkID
	listener      net.Listener
	connections   map[NetworkID]net.Conn
	peerAddresses map[NetworkID]string
	isClosed      bool
}

// Initiliaze a new TCP network
func NewTCPNetwork(localID NetworkID, opts TCPNetworkOpts) *TCPNetwork {
	n := &TCPNetwork{
		TCPNetworkOpts: opts,
		id:             localID,
		connections:    make(map[NetworkID]net.Conn),
		peerAddresses:  make(map[NetworkID]string),
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
	n.Lock()
	n.peerAddresses[remoteID] = addr
	n.Unlock()
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
		addr, ok := n.peerAddresses[remoteID]
		n.Unlock()
		if ok {
			go n.retryConnect(remoteID, addr)
		} else {
			n.Logger.Sugar().Warnf("no address found for peer %s to reconnect", remoteID)
		}
		return fmt.Errorf("failed to send message: %v", err)
	} else {
		n.Logger.Sugar().Infof("sent message to '%s': %s", remoteID, message.Payload)
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
		remoteID := NetworkID(remoteAddr)
		n.Lock()
		n.connections[remoteID] = conn
		n.peerAddresses[remoteID] = remoteAddr
		if err := n.HandshakeFn(); err != nil {
			n.Logger.Sugar().Errorf("error on handsharemoteIDking with %s: %v\n", remoteAddr, err)
			conn.Close()
			delete(n.connections, remoteID)
			delete(n.peerAddresses, remoteID)
			n.Unlock()
			return err
		}
		n.Unlock()

		n.Logger.Sugar().Infof("connected to remote peer %s\n", remoteAddr)

		// Read loop
		go n.listenForMessages(conn)
	}
}

// listenForMessages listens for incoming messages.
func (n *TCPNetwork) listenForMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	var remoteID NetworkID

	n.Lock()
	for id, c := range n.connections {
		if c == conn {
			remoteID = id
			break
		}
	}
	n.Unlock()

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			n.Logger.Debug("connection lost. Reconnecting...")
			n.Lock()
			delete(n.connections, remoteID)
			addr, ok := n.peerAddresses[remoteID]
			n.Unlock()
			if ok {
				go n.retryConnect(remoteID, addr)
			} else {
				n.Logger.Sugar().Warnf("no address found for peer %s to reconnect", remoteID)
			}
			return
		}

		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			n.Logger.Sugar().Errorf("failed to unmarshal message: %v\n", err)
			continue
		}

		n.Logger.Sugar().Infof("received message from '%s': %s", remoteID, string(message.Payload))

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

		if addr == "" {
			n.Logger.Sugar().Warnf("no address to retry connection for peer %s", remoteID)
			n.Lock()
			delete(n.peerAddresses, remoteID)
			n.Unlock()
			return
		}

		conn, err := net.Dial("tcp", addr)

		if err != nil {
			n.Logger.Sugar().Errorf("failed to connect to %s (%s): %v. Retrying in %v...", remoteID, addr, err, n.RetryDelay)
			time.Sleep(n.RetryDelay)
			if !n.isClosed && n.RetryDelay < 2*time.Minute {
				n.RetryDelay *= 2
			} else if !n.isClosed {
				n.Lock()
				delete(n.connections, remoteID)
				delete(n.peerAddresses, remoteID)
				n.Unlock()
				n.Logger.Sugar().Infof("stopped retrying and removed peer %s", remoteID)
				return
			} else {
				return // Exit if the network is closed
			}
			continue
		}

		n.Lock()
		n.connections[remoteID] = conn
		n.Unlock()

		n.Logger.Sugar().Infof("successfully connected to peer %s (%s)!", remoteID, addr)

		go n.listenForMessages(conn)
		return
	}
}
