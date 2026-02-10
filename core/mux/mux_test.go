package mux

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
)

func TestMultiplexerStart(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	serverSession, _ := yamux.Server(serverConn, nil)
	clientSession, _ := yamux.Client(clientConn, nil)

	mux := NewMultiplexer(uuid.New().String(), 8080, "test.local", serverSession)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mux.Start(ctx)
	stream, _ := clientSession.Open()

	go func() {
		stream.Write([]byte("msg from client"))
		buf := make([]byte, 1024)
		n, _ := stream.Read(buf)
		fmt.Println("Client received:", string(buf[:n]))
		stream.Close()
	}()

	time.Sleep(1 * time.Second)
	fmt.Println("Test finished")
}
