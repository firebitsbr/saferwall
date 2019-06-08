// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"

	pb "github.com/saferwall/saferwall/core/multiav/clamav/proto"
	"github.com/saferwall/saferwall/pkg/multiav/clamav"
	"github.com/saferwall/saferwall/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"

	// grpc library default is 4MB
	maxMsgSize = 1024 * 1024 * 20
)

// DefaultServerOpts returns the set of default grpc ServerOption's that Tiller requires.
func DefaultServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	}
}

// server is used to implement clamav.ClamAVScanner.
type server struct{}

// GetVersion implements clamav.ClamAVScanner.
func (s *server) GetVersion(ctx context.Context, in *pb.VersionRequest) (*pb.VersionResponse, error) {
	version, err := clamav.GetVersion()
	return &pb.VersionResponse{Version: version}, err
}

// ScanFile implements clamav.ClamAVScanner.
func (s *server) ScanFile(ctx context.Context, in *pb.ScanFileRequest) (*pb.ScanResponse, error) {
	log.Infoln("ScanFile ", in.Filepath)
	res, err := clamav.ScanFile(in.Filepath)
	return &pb.ScanResponse{Infected: res.Infected, Output: res.Output}, err
}

// NewServer creates a new grpc server.
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append(DefaultServerOpts(), opts...)...)
}

// StartService starts the Avast service.
func StartService() error {
	_, err := utils.ExecCommand("clamd")
	return err
}

// main start a gRPC server and waits for connection.
func main() {

	// start by running clamd
	log.Infoln("Starting clamav daemon")
	err := StartService()
	if err != nil {
		log.Fatal(err)
	}

	// create a listener on TCP port 50051
	log.Infoln("Starting clamav gRPC server")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	// create a gRPC server object
	s := NewServer()

	// attach the AvastScanner service to the server
	pb.RegisterClamAVScannerServer(s, &server{})

	// register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		grpclog.Fatalf("failed to serve: %v", err)
	}

}