package mux

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/yamux"
)

type Multiplexer struct {
	ID            string
	LocalPort     int
	Domain        string
	Session       *yamux.Session
	StreamCounter atomic.Uint64
	mu            sync.Mutex
	CreatedAt     time.Time
	Active        bool
}

func NewMultiplexer(id string, localPort int, domain string, session *yamux.Session) *Multiplexer {
	return &Multiplexer{
		ID:        id,
		LocalPort: localPort,
		Domain:    domain,
		Session:   session,
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
		wg.Wait()
		fmt.Printf("Multiplexer for domain=%s shutting down\n", t.Domain)
	}()

	for {
		stream, err := t.Session.AcceptStreamWithContext(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("Multiplexer shutting down...")
				return
			default:
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					if errors.Is(err, io.EOF) || t.Session.IsClosed() {
						return
					}
					continue
				}
			}

			wg.Add(1)

			go func(s *yamux.Stream) {
				defer wg.Done()
				defer s.Close()

				t.HandleStream(s, t.LocalPort)
			}(stream)
		}
	}
}

// Helper to proxy one request
// Connect to local service at localhost:localPort
// Bidirectionally copy data between the stream and the local connection
// copy from stream to local connection and vice versa
// io.Copy bidirectionally between stream and local connection (sorry for the mess)
// and finally, wait for either copy to finish
func (t *Multiplexer) HandleStream(stream *yamux.Stream, localPort int) {
	localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		return
	}
	defer localConn.Close()

	tcpConn := localConn.(*net.TCPConn)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(tcpConn, stream)
		tcpConn.CloseWrite()
	}()

	go func() {
		defer wg.Done()
		io.Copy(stream, tcpConn)
		stream.Close()
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

// Close the multiplexer and all active streams.
// Set Active to false, close all streams and the yamux session.
// This will cause the Start loop to exit and the multiplexer to stop.
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
