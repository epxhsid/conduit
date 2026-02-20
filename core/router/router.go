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

// GetSession retrieves the yamux stream associated with the given domain.
// It returns the stream and a boolean indicating whether the stream exists.
// example of domain: "test.local" maps to a stream that proxies to localhost:8080
// This is used by the HTTP handler to find the correct stream to proxy the request to based on the Host header.
// If a request comes in with Host "test.local", the handler will call GetSession("test.local")
// to get the stream to proxy the request through. If the stream exists, it will forward the request to the local service.
// If the stream does not exist, it can return an error or a default response indicating that the domain is not available.
func (r *Registry) Get(domain string) (*Tunnel, bool) {
	r.RLock()
	defer r.RUnlock()
	t, ok := r.Tunnels[domain]
	return t, ok
}

// AddSession adds a new yamux stream to the registry for the specified domain.
// This is called when a new client connects and establishes a session for a specific domain.
// example: when a client connects and wants to proxy requests for "test.local",
// we create a yamux stream for that client and add it to the registry with the key "test.local".
// RemoveSession removes the yamux stream associated with the given domain from the registry.
// This is called when a client disconnects or when we want to clean up resources for a domain.
func (r *Registry) Register(domain string, session *yamux.Session, targetPort int) {
	r.Lock()
	defer r.Unlock()
	r.Tunnels[domain] = &Tunnel{
		Session:    session,
		TargetPort: targetPort,
		CreatedAt:  time.Now(),
	}
}

// RemoveSession removes the yamux stream associated with the given domain from the registry.
// This is called when a client disconnects or when we want to clean up resources for a domain.
// example: if the client that was proxying requests for "test.local" disconnects, we call
// RemoveSession("test.local") to remove the stream from the registry, so that future requests
// for "test.local" will not find a stream and can return an error or a default response.
func (r *Registry) Remove(domain string) {
	r.Lock()
	defer r.Unlock()
	delete(r.Tunnels, domain)
}
