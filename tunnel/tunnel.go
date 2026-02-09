package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
)

type Tunnel struct {
	ID        string
	LocalPort int
	Domain    string
	Session   *yamux.Session
	Streams   map[string]*yamux.Stream
	mu        sync.Mutex
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

type TunnelInterface interface {
	Start(ctx context.Context)
	HandleStream(stream *yamux.Stream, localPort int)
	Close()
	ProxyStream()
	ConnectToService(svcAddr, domain string, localPort int) (*Tunnel, error)
}

// TODO: Implement tunnel start logic
// Accept incoming streams from the yamux session
// For each stream, reverse proxy data from domain:Domain to localhost:LocalPort
// run in a loop until tunnel is closed
func (t *Tunnel) Start(ctx context.Context) {
	fmt.Printf("Tunnel started: domain=%s localPort=%d\n", t.Domain, t.LocalPort)
	var wg sync.WaitGroup
	defer func() {
		fmt.Println("Waiting for active streams to finish...")
		wg.Wait()
		fmt.Println("Tunnel stopped")
	}()

	for {
		stream, err := t.Session.AcceptStreamWithContext(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("Tunnel shutting down...")
				return
			default:
				fmt.Printf("Accept stream error: %v\n", err)
				continue
			}

		}

		streamID := uuid.New().String()
		t.mu.Lock()
		t.Streams[streamID] = stream
		t.mu.Unlock()

		fmt.Printf("New stream %s accepted for domain=%s\n", streamID, t.Domain)

		wg.Add(1)
		go func(id string, s *yamux.Stream) {
			defer wg.Done()
			defer func() {
				t.mu.Lock()
				delete(t.Streams, id)
				t.mu.Unlock()
				s.Close()
				fmt.Printf("Stream %s finished\n", id)
			}()

			t.HandleStream(s, t.LocalPort)
		}(streamID, stream)
	}
}

// TODO: Implement stream handling logic
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
	t.Active = false

	for _, stream := range t.Streams {
		stream.Close()
	}

	if t.Session != nil {
		t.Session.Close()
	}
}

func (t *Tunnel) ProxyStream() {
	// TODO: Implement stream proxying logic
	// Takes one stream (one HTTP request)
	// forward data between the stream and the local service
	// copies response back through the stream
}

// TODO: Dial TCP to server
// Create yamux client session
// Generate a unique tunnel ID (e.g. UUID)
// assign the new Tunnel instance to the variable t
// send a handshake with the domain and localPort
// return the Tunnel instance
func (t *Tunnel) ConnectToService(svcAddr, domain string, localPort int) (*Tunnel, error) {
	conn, err := net.Dial("tcp", svcAddr)
	if err != nil {
		return nil, err
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	t = NewTunnel(id, localPort, domain, session)
	return t, nil
}
