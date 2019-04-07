package main

import (
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/envoyproxy/go-control-plane/pkg/test/resource"
	"google.golang.org/grpc"
)

const xDSListenAddr = "0.0.0.0:5678"

func main() {
	ln, err := net.Listen("tcp", xDSListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	snapshotCache := cache.NewSnapshotCache(false, NodeIDHash{}, nil)

	grpcServer := grpc.NewServer()
	envoyServer := server.NewServer(snapshotCache, nil)

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, envoyServer)

	// This snapshot has no values. I have no idea what this will do to Envoy!
	var clusters, endpoints, routes, listeners []cache.Resource
	// TODO: For some reason the endpoint is not being picked up but the cluster is!
	clusters = append(clusters, resource.MakeCluster(resource.Ads, "some_service"))
	endpoints = append(endpoints, resource.MakeEndpoint("some_service", 7777))
	snapshot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners)
	err = snapshotCache.SetSnapshot("node1", snapshot)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Listening on %s...", xDSListenAddr)
	log.Fatal(grpcServer.Serve(ln))
}

type NodeIDHash struct{}

func (h NodeIDHash) ID(node *core.Node) string {
	if node != nil {
		return node.Id
	}
	return "node"
}
