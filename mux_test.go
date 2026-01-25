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

func TestOpenStreamIDSequence(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	mux := NewMultiplexer(client)

	expectedIDs := []uint32{1, 3, 5, 7, 9}
	for _, expectedID := range expectedIDs {
		stream := mux.OpenStream()
		if stream.id != expectedID {
			t.Errorf("Expected stream ID %d, got %d", expectedID, stream.id)
		}
	}
}

func TestOpenStreamConcurrent(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	mux := NewMultiplexer(client)

	const numStreams = 10
	streams := make([]*Stream, numStreams)
	done := make(chan bool, numStreams)

	for i := 1; i <= numStreams; i++ {
		go func(index int) {
			streams[index-1] = mux.OpenStream()
			done <- true
		}(i)
	}

	for i := 1; i <= numStreams; i++ {
		<-done
	}

	seenIDs := make(map[uint32]bool)
	for _, stream := range streams {
		if seenIDs[stream.id] {
			t.Errorf("Duplicate stream ID: %d", stream.id)
		}
		seenIDs[stream.id] = true
	}

	mux.mu.Lock()
	if len(mux.streams) != numStreams {
		t.Errorf("Expected %d streams, got %d", numStreams, len(mux.streams))
	}
	fmt.Printf("Multiplexer has %d streams\n", len(mux.streams))
	mux.mu.Unlock()
}
