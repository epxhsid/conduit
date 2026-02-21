package router

import (
	"context"
	"fmt"
	"io"
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

		go handleRouterConnection(ctx, conn, registry)
	}
}

func handleRouterConnection(ctx context.Context, conn net.Conn, registry *Registry) {
	defer conn.Close()

	buf := make([]byte, 256)

	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	domain := strings.TrimSpace(string(buf[:n]))

	tunnel, ok := registry.Get(domain)
	if !ok {
		fmt.Println("No tunnel for domain:", domain)
		return
	}

	stream, err := tunnel.Session.OpenStream()
	if err != nil {
		fmt.Println("Failed to open yamux stream:", err)
		return
	}

	go pipe(conn, stream)
	go pipe(stream, conn)
}

func pipe(a io.ReadWriteCloser, b io.ReadWriteCloser) {
	defer a.Close()
	defer b.Close()

	io.Copy(a, b)
}
