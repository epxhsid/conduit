package router

import (
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

type Registry struct {
	sync.RWMutex
	Tunnels map[string]*Tunnel
}

func NewRegistry() *Registry {
	return &Registry{
		Tunnels: make(map[string]*Tunnel),
	}
}

type Tunnel struct {
	Session    *yamux.Session
	TargetPort int
	CreatedAt  time.Time
}

func (r *Registry) Get(domain string) (*Tunnel, bool) {
	r.RLock()
	defer r.RUnlock()
	t, ok := r.Tunnels[domain]
	return t, ok
}

func (r *Registry) Register(domain string, session *yamux.Session, targetPort int) {
	r.Lock()
	defer r.Unlock()
	r.Tunnels[domain] = &Tunnel{
		Session:    session,
		TargetPort: targetPort,
		CreatedAt:  time.Now(),
	}
}

func (r *Registry) Remove(domain string) {
	r.Lock()
	defer r.Unlock()
	delete(r.Tunnels, domain)
}
