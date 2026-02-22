package router

import (
	"context"
	"fmt"
	"net"
	"strings"
)

func StartRouter(ctx context.Context, addr string, registry *Registry) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	fmt.Println("Router listening on", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		go handleRouterConnection(conn, registry, NewPipe())
	}
}

func handleRouterConnection(conn net.Conn, registry *Registry, pipe *Pipe) {
	defer conn.Close()

	buf := make([]byte, 256)

	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	domain := strings.TrimSpace(string(buf[:n]))

	tunnel, ok := registry.Get(domain)
	if !ok {
		fmt.Println("Unknown domain:", domain)
		return
	}

	stream, err := tunnel.Session.OpenStream()
	if err != nil {
		fmt.Println("Failed to open tunnel stream:", err)
		return
	}

	pipe.Pipe(conn, stream)
}
