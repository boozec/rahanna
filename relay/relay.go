package relay

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/boozec/rahanna/relay/proto"
)

type Server struct {
	pb.UnimplementedRelayServer
}

type name string

// TODO: use pair of ips and ports
type ips struct {
	ip0 string
	ip1 string
}

var mu sync.Mutex

// Map each name to a pair of IPs
var table = make(map[name]ips)

func (s *Server) RegisterName(ctx context.Context, in *pb.RelayRequest) (*pb.RelayResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	if in.Ip == "" {
		return nil, fmt.Errorf("IP address cannot be empty")
	}

	sessionName := newSession()
	for {
		if _, ok := table[name(sessionName)]; !ok {
			break
		}
		sessionName = newSession()
	}

	table[name(sessionName)] = ips{ip0: in.Ip, ip1: ""}
	return &pb.RelayResponse{Name: sessionName, Ip: in.Ip}, nil
}

func (s *Server) Lookup(ctx context.Context, in *pb.LookupRequest) (*pb.RelayResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	if in.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	entry, ok := table[name(in.Name)]
	if !ok {
		return nil, fmt.Errorf("name not found")
	}

	return &pb.RelayResponse{Name: in.Name, Ip: entry.ip0}, nil
}

func (s *Server) CloseName(ctx context.Context, in *pb.LookupRequest) (*pb.CloseResponse, error) {
	mu.Lock()
	defer mu.Unlock()

	if in.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	_, ok := table[name(in.Name)]
	if !ok {
		return &pb.CloseResponse{Status: false}, fmt.Errorf("name not found")
	}

	delete(table, name(in.Name))
	return &pb.CloseResponse{Status: true}, nil
}
