package tunnel

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/hashicorp/yamux"
)

type Tunnel struct {
	ID        string
	LocalPort int
	Domain    string
	Session   *yamux.Session
	Streams   map[string]*yamux.Stream
	CreatedAt time.Time
	Active    bool
}

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

// TODO: Implement tunnel start logic
// Accept incoming streams from the yamux session
// For each stream, reverse proxy data from domain:Domain to localhost:LocalPort
// run in a loop until tunnel is closed
func (t *Tunnel) Start() {

	for t.Active {
		stream, err := t.Session.AcceptStream()
		if err != nil {
			if t.Active {
				fmt.Printf("Error accepting stream: %v\n", err)
			}
			return
		}
		go t.HandleStream(stream, t.LocalPort)
	}

}

// TOOD: Implement stream handling logic
// Helper to proxy one request
// Connect to local service at localhost:localPort
// Bidirectionally copy data between the stream and the local connection
// copy from stream to local connection and vice versa
// io.Copy bidirectionally between stream and local connection (sorry for the mess)
// and finally, wait for either copy to finish
func (t *Tunnel) HandleStream(stream *yamux.Stream, localPort int) {

	localAddr := fmt.Sprintf("localhost:%d", localPort)
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		fmt.Printf("Error connecting to local service: %v\n", err)
		stream.Close()
		return
	}
	defer localConn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := io.Copy(localConn, stream)
		if err != nil {
			fmt.Printf("Error copying from stream to local connection: %v\n", err)
		}
		done <- struct{}{}
	}()

	go func() {
		defer close(done)
		_, err := io.Copy(stream, localConn)
		if err != nil {
			fmt.Printf("Error copying from local connection to stream: %v\n", err)
		}
		done <- struct{}{}
	}()

	<-done
}

// TODO: Implement tunnel close logic
// clean shutdown of the session
// mark tunnel as inactive
// close all streams
func (t *Tunnel) Close() {

}

func (t *Tunnel) ProxyStream() {
	// TODO: Implement stream proxying logic
	// Takes one stream (one HTTP request)
	// forward data between the stream and the local service
	// copies response back through the stream
}

func (t *Tunnel) ConnectToService(svcAddr, domain string, localPort int) (*Tunnel, error) {
	// TODO: Dial TCP to your server
	conn, err := net.Dial("tcp", svcAddr)
	if err != nil {
		return nil, err
	}
	// Create yamux client session
	session, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}
	// Generate a unique tunnel ID (e.g. UUID)
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	// assign the new Tunnel instance to the variable t
	t = NewTunnel(id, localPort, domain, session)
	// send a handshake with the domain and localPort
	// return the Tunnel instance
	return t, nil
}
