package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/hashicorp/yamux"
)

type Client struct {
	VpsAddr   string
	Domain    string
	LocalPort int
}

func NewClient(vpsAddr, domain string, localPort int) *Client {
	return &Client{
		VpsAddr:   vpsAddr,
		Domain:    domain,
		LocalPort: localPort,
	}
}

func (c *Client) Run(ctx context.Context) {
	delay := time.Second

	for {
		err := c.Start(ctx)
		if err == nil {
			return
		}

		time.Sleep(delay)

		delay *= 2
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
	}
}

func (c *Client) Start(ctx context.Context) error {
	conn, err := net.Dial("tcp", c.VpsAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to VPS: %w", err)
	}

	// Temporary: send domain as plain text header
	// (Binary protocol will replace this later)
	fmt.Fprintf(conn, "%s\n", c.Domain)

	session, err := yamux.Client(conn, nil)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create yamux session: %w", err)
	}

	fmt.Println("Client connected:", c.Domain)

	for {
		stream, err := session.AcceptStream()
		if err != nil {
			if ctx.Err() != nil || session.IsClosed() {
				return nil
			}
			fmt.Printf("Failed to accept stream: %v\n", err)
			return err
		}

		go func(s *yamux.Stream) {
			localConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", c.LocalPort))

			if err != nil {
				fmt.Printf("Failed to connect to local service: %v\n", err)
				s.Close()
				return
			}

			defer localConn.Close()
			defer s.Close()

			go io.Copy(localConn, s)
			io.Copy(s, localConn)
		}(stream)
	}
}
