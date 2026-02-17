package main

import (
	"context"
	"fmt"
	"net"

	"github.com/epxhsid/conduit/core/mux"
	"github.com/hashicorp/yamux"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:7000")
	if err != nil {
		panic(err)
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		panic(err)
	}

	m := mux.NewMultiplexer("test-client", 8080, "127.0.0.1:5678", session)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("Client logic started. Routing localhost:8080 to tunnel...\n")
	m.Start(ctx)
}
