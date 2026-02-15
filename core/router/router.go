package router

import (
	"sync"

	"github.com/hashicorp/yamux"
)

type Registry struct {
	sync.RWMutex
	Streams map[string]*yamux.Stream
}

func NewRegistry() *Registry {
	return &Registry{
		Streams: make(map[string]*yamux.Stream),
	}
}
