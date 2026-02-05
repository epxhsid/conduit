package tunnel

import (
	"time"

	"github.com/hashicorp/yamux"
)

// Tunnel represents a multiplexed TCP tunnel session between the local service and the public domain.
type Tunnel struct {
	ID        string
	LocalPort int
	Domain    string
	Session   *yamux.Session
	Streams   map[string]*yamux.Stream
	CreatedAt time.Time
	Active    bool
}

// NewTunnel initializes a new Tunnel instance.
func NewTunnel(id string, localPort int, domain string, session *yamux.Session) *Tunnel {
	return &Tunnel{
		ID:        id,
		LocalPort: localPort,
		Domain:    domain,
		Session:   session,
		Streams:   make(map[string]*yamux.Stream),
		CreatedAt: time.Now(),
		Active:    true,
	}
}
