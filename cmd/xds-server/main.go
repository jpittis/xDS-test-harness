package main

import (
	"log"
	"net/http"
)

const (
	xDSListenAddr  = "0.0.0.0:5678"
	shimListenAddr = "0.0.0.0:4678"
)

func main() {
	shim := NewShim(xDSListenAddr)
	err := http.ListenAndServe(shimListenAddr, NewShimHandler(shim))
	if err != nil {
		log.Fatal(err)
	}
}
