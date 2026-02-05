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

func (t *Tunnel) Start() {
	// TODO: Implement tunnel start logic
	// Accept incoming streams from the yamux session
	// For each stream, reverse proxy data from domain:Domain to localhost:LocalPort
	// run in a loop until tunnel is closed
}

func (t *Tunnel) ProxyStream() {
	// TODO: Implement stream proxying logic
	// Takes one stream (one HTTP request)
	// forward data between the stream and the local service
	// copies response back through the stream
}

func (t *Tunnel) Close() {
	// TODO: Implement tunnel close logic
	// clean shutdown of the session
	// mark tunnel as inactive
	// close all streams
}

func (t *Tunnel) ConnectToService(svcAddr, domain string, localPort int) *Tunnel {
	// TODO: Dial TCP to your server
	// Create yamux client session
	// send a handshake with the domain and localPort
	// return the Tunnel instance
	return nil
}

func (t *Tunnel) HandleStream(stream *yamux.Stream, localPort int) {
	// TOOD: Implement stream handling logic
	// Helper to proxy one request
	// Connect to local service at localhost:localPort
	// io.Copy bidirectionally between stream and local connection
}
