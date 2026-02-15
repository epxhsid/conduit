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

func (r *Registry) GetSession(domain string) (*yamux.Stream, bool) {
	r.RLock()
	defer r.RUnlock()
	stream, exists := r.Streams[domain]
	return stream, exists
}
