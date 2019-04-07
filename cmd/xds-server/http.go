package main

import (
	"net/http"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/test/resource"
)

func NewShimHandler(shim *Shim) *ShimHandler {
	h := &ShimHandler{
		ServeMux: http.NewServeMux(),
		shim:     shim,
	}
	h.ServeMux.HandleFunc("/start_server", h.StartServer)
	h.ServeMux.HandleFunc("/stop_server", h.StopServer)
	h.ServeMux.HandleFunc("/set_snapshot", h.SetSnapshot)
	return h
}

type ShimHandler struct {
	*http.ServeMux
	shim *Shim
}

func (h *ShimHandler) StartServer(w http.ResponseWriter, req *http.Request) {
	err := h.shim.StartServer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ShimHandler) StopServer(w http.ResponseWriter, req *http.Request) {
	err := h.shim.StopServer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ShimHandler) SetSnapshot(w http.ResponseWriter, req *http.Request) {
	// TODO: Parse this data from the HTTP body.
	var clusters, endpoints, routes, listeners []cache.Resource
	clusters = append(clusters, resource.MakeCluster(resource.Ads, "some_service"))
	snapshot := cache.NewSnapshot("1.0", endpoints, clusters, routes, listeners)
	node := "node1"

	err := h.shim.SetSnapshot(node, snapshot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
