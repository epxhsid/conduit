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
	ActiveStreams atomic.Int64
	TotalStreams  atomic.Uint64
	MaxStreams    int64
	activeWG      sync.WaitGroup
	mu            sync.Mutex
	CreatedAt     time.Time
	Active        atomic.Bool
}

func NewMultiplexer(id string, localPort int, domain string, session *yamux.Session) *Multiplexer {
	m := &Multiplexer{
		ID:        id,
		LocalPort: localPort,
		Domain:    domain,
		Session:   session,
		CreatedAt: time.Now(),
	}
	m.Active.Store(true)
	return m
}

func (t *Multiplexer) Start(ctx context.Context) {
	fmt.Printf("Multiplexer started: domain=%s localPort=%d\n", t.Domain, t.LocalPort)

	for {
		stream, err := t.Session.AcceptStreamWithContext(ctx)
		if err != nil {
			if ctx.Err() != nil || t.Session.IsClosed() || errors.Is(err, io.EOF) {
				fmt.Println("Multiplexer shutting down...")
				return
			}

			fmt.Printf("Error accepting stream: %v\n", err)
			continue
		}

		if t.MaxStreams > 0 {
			active := t.ActiveStreams.Load()
			if active >= t.MaxStreams {
				stream.Close()
				continue
			}

			t.ActiveStreams.Add(1)
		} else {
			t.ActiveStreams.Add(1)
		}

		t.TotalStreams.Add(1)
		t.activeWG.Add(1)

		go func(s *yamux.Stream) {
			defer t.activeWG.Done()
			defer t.ActiveStreams.Add(-1)
			defer s.Close()
			if ctx.Err() != nil {
				t.HandleStream(s, t.LocalPort)
			} else {
				t.HandleStreamWithContext(ctx, s, t.LocalPort)
			}
		}(stream)
	}
}

func (t *Multiplexer) HandleStream(stream *yamux.Stream, localPort int) {
	localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		return
	}
	defer localConn.Close()

	tcpConn, ok := localConn.(*net.TCPConn)
	if !ok {
		localConn.Close()
		stream.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if _, err := io.Copy(tcpConn, stream); err != nil && !errors.Is(err, net.ErrClosed) {
			fmt.Printf("Error copying from stream to local connection: %v\n", err)
		}
		tcpConn.CloseWrite()
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, tcpConn); err != nil && !errors.Is(err, net.ErrClosed) {
			fmt.Printf("Error copying from local connection to stream: %v\n", err)
		}
		stream.Close()
	}()

	wg.Wait()
}

func (t *Multiplexer) HandleStreamWithContext(ctx context.Context, stream *yamux.Stream, localPort int) {
	localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		stream.Close()
		return
	}
	defer localConn.Close()
	defer stream.Close()

	tcpConn, ok := localConn.(*net.TCPConn)
	if !ok {
		localConn.Close()
		stream.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		<-ctx.Done()
		stream.SetDeadline(time.Now())
		tcpConn.SetDeadline(time.Now())
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(tcpConn, stream); err != nil && !errors.Is(err, net.ErrClosed) {
			fmt.Printf("Error copying from stream to local connection: %v\n", err)
		}
		tcpConn.CloseWrite()
	}()

	go func() {
		defer wg.Done()
		if _, err := io.Copy(stream, tcpConn); err != nil && !errors.Is(err, net.ErrClosed) {
			fmt.Printf("Error copying from local connection to stream: %v\n", err)
		}
		stream.Close()
	}()

	wg.Wait()
}

func (m *Multiplexer) Shutdown() {
	m.Active.Store(false)
	m.Session.Close()
	m.activeWG.Wait()
	fmt.Printf("Multiplexer for domain=%s shut down\n", m.Domain)
}

func (t *Multiplexer) Stats() string {
	return fmt.Sprintf(
		"active=%d total=%d",
		t.ActiveStreams.Load(),
		t.TotalStreams.Load(),
	)
}
