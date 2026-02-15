package mux

import (
	"context"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
)

// basic test to verify multiplexer can accept incoming streams and proxy data to local service
func TestMultiplexerStart(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer ln.Close()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	localPort, _ := strconv.Atoi(portStr)

	go func() {
		conn, _ := ln.Accept()
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		conn.Write([]byte("echo: " + string(buf[:n])))
	}()

	clientConn, serverConn := net.Pipe()
	serverSession, _ := yamux.Server(serverConn, nil)
	clientSession, _ := yamux.Client(clientConn, nil)

	mux := NewMultiplexer("test_id", localPort, "test.local", serverSession)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mux.Start(ctx)

	stream, err := clientSession.OpenStream()
	if err != nil {
		t.Fatalf("Failed to open stream: %v", err)
	}
	defer stream.Close()

	payload := "msg from client"
	stream.Write([]byte(payload))

	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := "echo: " + payload
	if string(buf[:n]) != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, string(buf[:n]))
	}
}

func TestMultiplexerLocalServiceUnavailable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	localPort, _ := strconv.Atoi(portStr)
	ln.Close()

	clientConn, serverConn := net.Pipe()
	serverSession, _ := yamux.Server(serverConn, nil)
	clientSession, _ := yamux.Client(clientConn, nil)

	mux := NewMultiplexer("test_id_unavail", localPort, "test.local", serverSession)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mux.Start(ctx)

	stream, err := clientSession.OpenStream()
	if err != nil {
		t.Fatalf("Failed to open stream: %v", err)
	}
	defer stream.Close()

	_, _ = stream.Write([]byte("hello"))
	_ = stream.SetReadDeadline(time.Now().Add(1 * time.Second))

	buf := make([]byte, 1024)
	_, err = stream.Read(buf)
	if err == nil {
		t.Fatal("Expected read error due to local service unavailability, got none")
	}

	if err == io.EOF || strings.Contains(err.Error(), "reset") || strings.Contains(err.Error(), "closed") {
		return
	}
	t.Fatalf("Expected error due to local service unavailability, got: %v", err)
}
