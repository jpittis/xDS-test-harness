package main

import (
	"net"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"
)

func NewShim(addr string) *Shim {
	return &Shim{
		Addr:   addr,
		Errors: make(chan error, 1),
	}
}

type Shim struct {
	Addr          string
	Errors        chan error
	snapshotCache cache.SnapshotCache
	grpcServer    *grpc.Server
	envoyServer   server.Server
	ln            net.Listener
}

func (s *Shim) StartServer() error {
	s.snapshotCache = cache.NewSnapshotCache(false, NodeIDHash{}, nil)
	s.grpcServer = grpc.NewServer()
	s.envoyServer = server.NewServer(s.snapshotCache, nil)

	discovery.RegisterAggregatedDiscoveryServiceServer(s.grpcServer, s.envoyServer)

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go func() {
		err := s.grpcServer.Serve(ln)
		s.Errors <- err
	}()

	return nil
}

func (s *Shim) StopServer() error {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}

	if s.ln != nil {
		s.ln.Close()
	}

	s.grpcServer = nil
	s.envoyServer = nil
	s.snapshotCache = nil
	s.ln = nil

	return nil
}

func (s *Shim) SetSnapshot(node string, snapshot cache.Snapshot) error {
	return s.snapshotCache.SetSnapshot(node, snapshot)
}

type NodeIDHash struct{}

func (h NodeIDHash) ID(node *core.Node) string {
	if node != nil {
		return node.Id
	}
	return "node"
}
