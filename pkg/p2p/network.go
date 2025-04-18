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
	Type      []byte    `json:"type"`
	Timestamp int64     `json:"timestamp"`
	Source    NetworkID `json:"source"`
	Payload   []byte    `json:"payload"`
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
type NetworkHandshakeFunc func(conn net.Conn) error

func DefaultHandshake(conn net.Conn) error {
	return nil
}

// Network options to define on new `TCPNetwork`
type TCPNetworkOpts struct {
	ListenAddr       string
	RetryDelay       time.Duration
	HandshakeFn      NetworkHandshakeFunc
	FirstHandshakeFn NetworkHandshakeFunc
	OnReceiveFn      NetworkMessageReceiveFunc
	Logger           *zap.Logger
}

// PeerConnection holds the connection and address of a peer.
type PeerConnection struct {
	Conn    net.Conn
	Address string
}

// TCPNetwork represents a TCP peer capable to send and receive messages
type TCPNetwork struct {
	sync.Mutex
	TCPNetworkOpts

	id              NetworkID
	listener        net.Listener
	connections     map[NetworkID]PeerConnection
	isClosed        bool
	handshakesCount uint
}

// Initiliaze a new TCP network
func NewTCPNetwork(localID NetworkID, opts TCPNetworkOpts) *TCPNetwork {
	n := &TCPNetwork{
		TCPNetworkOpts: opts,
		id:             localID,
		connections:    make(map[NetworkID]PeerConnection),
	}

	go n.startServer()

	return n
}

// Close listener' connection
func (n *TCPNetwork) Close() error {
	n.isClosed = true
	if n.listener != nil {
		err := n.listener.Close()
		if err != nil {
			return err
		}
	}
	n.Lock()
	for _, pc := range n.connections {
		if pc.Conn != nil {
			pc.Conn.Close()
		}
	}
	n.connections = nil
	n.Unlock()
	return nil
}

// Add a new peer connection to the local peer
func (n *TCPNetwork) AddPeer(remoteID NetworkID, addr string) {
	n.Lock()
	n.connections[remoteID] = PeerConnection{Address: addr}
	n.Unlock()

	go n.retryConnect(remoteID, addr)
}

// Send methods is used to send a message to a specified remote peer
func (n *TCPNetwork) Send(remoteID NetworkID, messageType []byte, payload []byte) error {
	n.Lock()
	peerConn, exists := n.connections[remoteID]
	n.Unlock()

	if !exists {
		return fmt.Errorf("not connected to peer %s", remoteID)
	}

	if peerConn.Conn == nil {
		n.Logger.Sugar().Warnf("connection to peer %s is nil, attempting reconnect", remoteID)
		go n.retryConnect(remoteID, peerConn.Address)
		return fmt.Errorf("connection to peer %s is nil", remoteID)
	}

	message := Message{
		Type:      messageType,
		Payload:   payload,
		Source:    n.id,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	_, err = peerConn.Conn.Write(append(data, '\n'))

	if err != nil {
		n.Logger.Sugar().Errorf("failed to send message to %s: %v. Reconnecting...", remoteID, err)
		n.removeConnection(remoteID)
		return fmt.Errorf("failed to send message: %v", err)
	} else {
		n.Logger.Sugar().Infof("sent message to '%s' (%s): %s", remoteID, peerConn.Address, string(message.Payload))
	}

	return nil
}

// RegisterHandler registers a callback for a message type.
func (n *TCPNetwork) RegisterHandler(callback NetworkMessageReceiveFunc) {
	n.OnReceiveFn = callback
}

// startServer starts a TCP server to accept connections.
func (n *TCPNetwork) startServer() {
	var err error

	n.listener, err = net.Listen("tcp", n.ListenAddr)
	if err != nil {
		n.Logger.Sugar().Errorf("failed to start server: %v", err)
		return
	}

	n.isClosed = false

	n.Logger.Sugar().Infof("server started on %s\n", n.ListenAddr)

	for {
		conn, err := n.listener.Accept()
		if n.isClosed {
			n.Logger.Info("server listener closed")
			return
		}
		if err != nil {
			n.Logger.Sugar().Errorf("failed to accept connection: %v\n", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

func (n *TCPNetwork) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	remoteID := NetworkID(remoteAddr)

	n.Lock()
	n.handshakesCount++
	n.connections[remoteID] = PeerConnection{Conn: conn, Address: remoteAddr}
	n.Unlock()

	if n.HandshakeFn != nil {
		if err := n.HandshakeFn(conn); err != nil {
			n.Logger.Sugar().Errorf("error on handshaking with %s: %v\n", remoteAddr, err)
			conn.Close()
			n.removeConnection(remoteID)
			return
		}
	}

	if n.FirstHandshakeFn != nil && n.handshakesCount == 1 {
		if err := n.FirstHandshakeFn(conn); err != nil {
			n.Logger.Sugar().Errorf("error on first handshake with %s: %v\n", remoteAddr, err)
			conn.Close()
			n.removeConnection(remoteID)
			return
		}

	}

	n.Logger.Sugar().Infof("connected to remote peer %s (%s)\n", remoteID, remoteAddr)

	n.listenForMessages(conn, remoteID)

	n.removeConnection(remoteID)
	conn.Close()
	n.Logger.Sugar().Infof("connection to %s closed\n", remoteAddr)
}

func (n *TCPNetwork) removeConnection(id NetworkID) {
	n.Lock()
	delete(n.connections, id)
	n.Unlock()
}

// listenForMessages listens for incoming messages on a specific connection.
func (n *TCPNetwork) listenForMessages(conn net.Conn, remoteID NetworkID) {
	reader := bufio.NewReader(conn)
	remoteAddr := conn.RemoteAddr().String()

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				n.Logger.Sugar().Debugf("connection to %s closed by remote peer", remoteAddr)
			} else {
				n.Logger.Sugar().Warnf("error reading from connection %s: %v", remoteAddr, err)
			}

			return
		}

		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			n.Logger.Sugar().Errorf("failed to unmarshal message from %s: %v\n", remoteAddr, err)
			continue
		}

		n.Logger.Sugar().Infof("received message from '%s' (%s): %s", message.Source, remoteAddr, string(message.Payload))

		if n.OnReceiveFn != nil {
			n.OnReceiveFn(message)
		}
	}
}

// retryConnect attempts to connect to a remote peer.
func (n *TCPNetwork) retryConnect(remoteID NetworkID, addr string) {
	if addr == "" {
		n.Logger.Sugar().Warnf("no address to reconnect to peer %s", remoteID)
		n.removeConnection(remoteID)
		return
	}

	retryDelay := n.RetryDelay
	for !n.isClosed {
		n.Lock()
		_, connected := n.connections[remoteID]
		n.Unlock()

		if connected {
			if n.connections[remoteID].Conn != nil {
				time.Sleep(5 * time.Second)
				continue
			}
		}

		conn, err := net.Dial("tcp", addr)

		if err == nil {
			n.Logger.Sugar().Infof("successfully connected to peer %s (%s)!", remoteID, addr)
			n.Lock()
			n.connections[remoteID] = PeerConnection{Conn: conn, Address: addr}
			n.Unlock()
			go n.handleConnection(conn)
			return
		} else {
			n.Logger.Sugar().Errorf("failed to connect to %s (%s): %v. Retrying in %v...", remoteID, addr, err, retryDelay)
			select {
			case <-time.After(retryDelay):
				if retryDelay < 2*time.Minute {
					retryDelay *= 2
				}
			case <-n.closed():
				n.Logger.Info("retryConnect stopped due to network closure (inner for)")
				return
			}
		}
	}
	n.Logger.Info("retryConnect stopped due to network closure")
}

func (n *TCPNetwork) closed() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		<-n.listenerClosed()
		close(ch)
	}()
	return ch
}

func (n *TCPNetwork) listenerClosed() <-chan struct{} {
	if n.listener == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	done := make(chan struct{})
	go func() {
		n.listener.Accept() // This will block until closed
		close(done)
	}()
	return done
}
