package main

import (
	"net"
	"os"

	"github.com/boozec/rahanna/relay"
	pb "github.com/boozec/rahanna/relay/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		logger.Sugar().Errorln("Failed to listen", "err", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	server := &relay.Server{}
	pb.RegisterRelayServer(s, server)

	reflection.Register(s)

	logger.Sugar().Infoln("Server listening", "address", lis.Addr())

	if err := s.Serve(lis); err != nil {
		logger.Sugar().Errorln("Failed to serve", "err", err)
		os.Exit(1)
	}
}
