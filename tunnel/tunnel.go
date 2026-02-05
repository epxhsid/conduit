package tunnel

import (
	"github.com/hashicorp/yamux"
)

// Tunnel represents a multiplexed TCP tunnel session between the local service and the public domain.
type Tunnel struct {
	ID        string
	LocalPort int
	Domain    string
	Session   *yamux.Session
	Stream    map[string]*yamux.Stream
}
