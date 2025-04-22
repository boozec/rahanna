package p2p

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/boozec/rahanna/internal/logger"
)

// Connect a DNS to get the address
func GetOutboundIP() net.IP {
	log, _ := logger.GetLogger()
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Sugar().Error("err", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// Returns a random available port on the node in the range of ephemeral ports
func GetRandomAvailablePort() (int, error) {
	for i := 0; i < 100; i += 1 {
		port := rand.Intn(65535-49152) + 1024
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			defer ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("failed to find an available port after multiple attempts")
}
