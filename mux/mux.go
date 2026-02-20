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

func (t *Multiplexer) HandleStreamWithContext(ctx context.Context, stream *yamux.Stream, localPort int) {
	localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		stream.Close()
		return
	}
	defer localConn.Close()
	defer stream.Close()

	tcpConn := localConn.(*net.TCPConn)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		<-ctx.Done()
		stream.SetDeadline(time.Now())
		tcpConn.SetDeadline(time.Now())
	}()

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
