package mux

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

type Multiplexer struct {
	ID        string
	LocalPort int
	Domain    string
	Session   *yamux.Session
	Streams   map[string]*yamux.Stream
	mu        sync.Mutex
	CreatedAt time.Time
	Active    bool
}

func NewMultiplexer(id string, localPort int, domain string, session *yamux.Session) *Multiplexer {
	return &Multiplexer{
		ID:        id,
		LocalPort: localPort,
		Domain:    domain,
		Session:   session,
		Streams:   make(map[string]*yamux.Stream),
		CreatedAt: time.Now(),
		Active:    true,
	}
}

type MultiplexerInterface interface {
	Start(ctx context.Context)
	HandleStream(stream *yamux.Stream, localPort int)
	HandleStreamWithContext(ctx context.Context, stream *yamux.Stream, localPort int)
	Close()
	ProxyStream()
	ConnectToService(svcAddr, domain string, localPort int) (*Multiplexer, error)
}

// TODO: Implement tunnel start logic
// Accept incoming streams from the yamux session
// For each stream, reverse proxy data from domain:Domain to localhost:LocalPort
// run in a loop until tunnel is closed
func (t *Multiplexer) Start(ctx context.Context) {
	fmt.Printf("Multiplexer started: domain=%s localPort=%d\n", t.Domain, t.LocalPort)
	var wg sync.WaitGroup
	defer func() {
		fmt.Println("Waiting for active streams to finish...")
		wg.Wait()
		fmt.Println("Multiplexer stopped")
	}()

	for {
		stream, err := t.Session.AcceptStreamWithContext(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("Multiplexer shutting down...")
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

// Helper to proxy one request
// Connect to local service at localhost:localPort
// Bidirectionally copy data between the stream and the local connection
// copy from stream to local connection and vice versa
// io.Copy bidirectionally between stream and local connection (sorry for the mess)
// and finally, wait for either copy to finish
func (t *Multiplexer) HandleStream(stream *yamux.Stream, localPort int) {
	localAddr := fmt.Sprintf("localhost:%d", localPort)
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		fmt.Printf("Error connecting to local service: %v\n", err)
		stream.Close()
		return
	}
	defer localConn.Close()
	defer stream.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if _, err := io.Copy(localConn, stream); err != nil {
			fmt.Printf("stream -> local error: %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, localConn); err != nil {
			fmt.Printf("local -> stream error: %v\n", err)
		}
	}()

	wg.Wait()
}

func (t *Multiplexer) HandleStreamWithContext(ctx context.Context, stream *yamux.Stream, localPort int) {
	localAddr := fmt.Sprintf("localhost:%d", localPort)
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		fmt.Printf("Error connecting to local service: %v\n", err)
		stream.Close()
		return
	}
	defer localConn.Close()
	defer stream.Close()

	done := make(chan struct{})

	copyFunc := func(dst, src net.Conn) {
		_, _ = io.Copy(dst, src)
		done <- struct{}{}
	}

	go copyFunc(localConn, stream)
	go copyFunc(stream, localConn)

	select {
	case <-ctx.Done():
		fmt.Println("Stream handling context cancelled, closing connections...")
		localConn.Close()
		stream.Close()
		return
	case <-done:
		fmt.Println("Stream handling completed")
		return
	}
}

// TODO: Implement tunnel close logic
// clean shutdown of the session
// mark tunnel as inactive
// close all streams
func (t *Multiplexer) Close() {
	t.Active = false

	for _, stream := range t.Streams {
		stream.Close()
	}

	if t.Session != nil {
		t.Session.Close()
	}
}

func (t *Multiplexer) ProxyStream() {
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
func (t *Multiplexer) ConnectToService(svcAddr, domain string, localPort int) (*Multiplexer, error) {
	conn, err := net.Dial("tcp", svcAddr)
	if err != nil {
		return nil, err
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	t = NewMultiplexer(id, localPort, domain, session)
	return t, nil
}
