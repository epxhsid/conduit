package portal

import (
	"fmt"
	"net"
	"testing"
)

func TestOpenStream(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	mux := NewMultiplexer(client)

	stream1 := mux.OpenStream()
	fmt.Printf("Opening stream ID %d\n", stream1.id)
	fmt.Printf("Multiplexer has %d streams\n", len(mux.streams))
	if stream1.id != 1 {
		t.Errorf("Expected stream ID 1, got %d", stream1.id)
	}

	if stream1.mux != mux {
		t.Error("Stream's multiplexer does not match the original multiplexer")
	}

	if stream1.closed {
		t.Error("Newly opened stream should not be closed")
	}

	stream2 := mux.OpenStream()
	fmt.Printf("Opening stream ID %d\n", stream2.id)
	fmt.Printf("Multiplexer has %d streams\n", len(mux.streams))

	if stream2.id != 3 {
		t.Errorf("Expected stream ID 3, got %d", stream2.id)
	}

	if stream2.mux != mux {
		t.Error("Stream's multiplexer does not match the original multiplexer")
	}

	mux.mu.Lock()
	if len(mux.streams) != 2 {
		t.Errorf("Expected 2 streams in map, got %d", len(mux.streams))
	}
	if mux.streams[1] != stream1 {
		t.Error("Stream 1 not properly registered")
	}
	if mux.streams[3] != stream2 {
		t.Error("Stream 3 not properly registered")
	}
	mux.mu.Unlock()
}
