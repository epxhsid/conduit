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

// Same as HandleStream but with context cancellation support
// so that if the context is cancelled, we can close the connections and stop handling the stream
// This is specifically needed for graceful shutdowns and timeouts
// We can use a select statement to wait for either the context to be cancelled or the copying to finish
// If the context is cancelled, otherwise we can just close the connections and return
// NOTE: goroutines can yeet their completion signal and exit properly even if the main function has already moved on [1]
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

	done := make(chan struct{}, 2) // [1]

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

func (t *Multiplexer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Active {
		fmt.Printf("Closing multiplexer for domain=%s\n", t.Domain)
		t.Active = false

		for id, s := range t.Streams {
			fmt.Printf("Closing stream %s\n", id)
			s.Close()
			delete(t.Streams, id)
		}

		if t.Session != nil {
			t.Session.Close()
		}
	}
}
