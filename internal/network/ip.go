package network

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net"
)

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
