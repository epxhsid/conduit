package server

import (
	"sync"

	"github.com/epxhsid/conduit/router"
)

type Server struct {
	registry   *router.Registry
	httpAddr   string
	tunnelAddr string
	mu         sync.Mutex
}

func NewServer(httpAddr, tunnelAddr string) *Server {
	return &Server{
		registry:   router.NewRegistry(),
		httpAddr:   httpAddr,
		tunnelAddr: tunnelAddr,
	}
}
