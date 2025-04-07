package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

// PeerInfo represents a peer's ID and IP.
type PeerInfo struct {
	ID   string `json:"id"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// Message represents a structured message.
type Message struct {
	Type      string   `json:"type"`
	Payload   []byte   `json:"payload"`
	Source    PeerInfo `json:"source"`
	Target    PeerInfo `json:"target"`
	Timestamp int64    `json:"timestamp"`
}

type NetworkCallback func(msg Message)

// TCPNetwork represents a full-duplex TCP peer.
type TCPNetwork struct {
	localPeer   PeerInfo
	connections map[string]net.Conn
	listener    net.Listener
	callbacks   map[string]NetworkCallback
	callbacksMu sync.RWMutex
	isConnected bool
	retryDelay  time.Duration
	sync.Mutex
}

// initializes a TCP peer
func NewTCPNetwork(localID, localIP string, localPort int) *TCPNetwork {
	n := &TCPNetwork{
		localPeer:   PeerInfo{ID: localID, IP: localIP, Port: localPort},
		connections: make(map[string]net.Conn),
		callbacks:   make(map[string]NetworkCallback),
		isConnected: false,
		retryDelay:  2 * time.Second,
	}

	go n.startServer()

	logger, _ = zap.NewProduction()

	return n
}

// Add a new peer connection to the local peer
func (n *TCPNetwork) AddPeer(remoteID string, remoteIP string, remotePort int) {
	go n.retryConnect(remoteID, remoteIP, remotePort)
}

// startServer starts a TCP server to accept connections.
func (n *TCPNetwork) startServer() {
	address := fmt.Sprintf("%s:%d", n.localPeer.IP, n.localPeer.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Sugar().Errorf("failed to start server: %v", err)
	}
	n.listener = listener
	logger.Sugar().Infof("server started on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Sugar().Errorf("failed to accept connection: %v\n", err)
			continue
		}

		remoteAddr := conn.RemoteAddr().String()
		n.Lock()
		n.connections[remoteAddr] = conn
		n.Unlock()
		n.isConnected = true
		n.retryDelay = 2 * time.Second

		logger.Sugar().Infof("connected to remote peer %s\n", remoteAddr)
		go n.listenForMessages(conn)
	}
}

// retryConnect attempts to connect to a remote peer.
func (n *TCPNetwork) retryConnect(remoteID, remoteIP string, remotePort int) {
	for {
		n.Lock()
		_, exists := n.connections[remoteID]
		n.Unlock()

		if exists {
			time.Sleep(5 * time.Second)
			continue
		}

		address := fmt.Sprintf("%s:%d", remoteIP, remotePort)
		conn, err := net.Dial("tcp", address)

		if err != nil {
			logger.Sugar().Errorf("failed to connect to %s: %v. Retrying in %v...", remoteID, err, n.retryDelay)
			time.Sleep(n.retryDelay)
			if n.retryDelay < 30*time.Second {
				n.retryDelay *= 2
			}
			continue
		}

		n.Lock()
		n.connections[remoteID] = conn
		n.Unlock()
		logger.Sugar().Infof("successfully connected to peer %s!", remoteID)

		go n.listenForMessages(conn)
	}
}

// Send sends a message to a specified remote peer.
func (n *TCPNetwork) Send(remoteID, messageType string, payload []byte) error {
	n.Lock()
	conn, exists := n.connections[remoteID]
	n.Unlock()

	if !exists {
		return fmt.Errorf("not connected to peer %s", remoteID)
	}

	msg := Message{
		Type:      messageType,
		Payload:   payload,
		Source:    n.localPeer,
		Target:    PeerInfo{ID: remoteID},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	_, err = conn.Write(append(data, '\n'))
	if err != nil {
		logger.Sugar().Errorf("failed to send message to %s: %v. Reconnecting...", remoteID, err)
		n.Lock()
		delete(n.connections, remoteID)
		n.Unlock()
		go n.retryConnect(remoteID, "", 0)
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// RegisterHandler registers a callback for a message type.
func (n *TCPNetwork) RegisterHandler(messageType string, callback NetworkCallback) {
	n.callbacksMu.Lock()
	n.callbacks[messageType] = callback
	n.callbacksMu.Unlock()
}

// listenForMessages listens for incoming messages.
func (n *TCPNetwork) listenForMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			logger.Debug("connection lost. Reconnecting...")
			n.Lock()
			for id, c := range n.connections {
				if c == conn {
					delete(n.connections, id)
					go n.retryConnect(id, "", 0)
					break
				}
			}
			n.Unlock()
			return
		}

		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			logger.Sugar().Errorf("failed to unmarshal message: %v\n", err)
			continue
		}

		n.callbacksMu.RLock()
		callback, exists := n.callbacks[message.Type]
		n.callbacksMu.RUnlock()

		if exists {
			go callback(message)
		}
	}
}

func (n *TCPNetwork) IsConnected() bool {
	n.Lock()
	defer n.Unlock()
	return n.isConnected
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		slog.Error("err", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func GetRandomAvailablePort() (int, error) {
	for i := 0; i < 100; i += 1 {
		port := rand.Intn(65535-1024) + 1024
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			defer ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("failed to find an available port after multiple attempts")
}
