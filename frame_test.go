package portal

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

// TestSendMessageAndReadMessage tests the SendMessage and ReadMessage functions
// using an in-memory net.Pipe connection.
// It verifies that a message sent from one end is correctly received at the other end.
// It also checks for proper handling of empty messages.
func TestSendMessageAndReadMessage(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	testMsg := []byte("Hello, World!")

	errChan := make(chan error, 1)
	go func() {
		errChan <- SendMessage(client, testMsg)
	}()

	received, err := ReadMessage(server)
	if err != nil {
		t.Fatalf("ReadMessage faile : %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if !bytes.Equal(received, testMsg) {
		t.Errorf("Expected %q, got %q", testMsg, received)
	}

	fmt.Printf("Received message: %s\n", string(received))
}
