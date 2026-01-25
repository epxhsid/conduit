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

	testMsg := []byte("sample message stream")

	errChan := make(chan error, 1)
	go func() {
		errChan <- SendMessage(client, testMsg)
	}()

	received, err := ReadMessage(server)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if !bytes.Equal(received, testMsg) {
		t.Errorf("Expected %q, got %q", testMsg, received)
	}

	fmt.Printf("Received message: %s\n", string(received))
}

func TestMaxMessageSize(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	largeMessage := make([]byte, MaxMessageSize+1)

	err := SendMessage(client, largeMessage)
	if err == nil {
		t.Error("Expected error for oversized message, got nil")
	}
}

func TestEmptyMessage(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	emptyMessage := []byte{}

	errChan := make(chan error, 1)
	go func() {
		errChan <- SendMessage(client, emptyMessage)
	}()
	received, err := ReadMessage(server)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if !bytes.Equal(received, emptyMessage) {
		t.Errorf("Expected %q, got %q", emptyMessage, received)
	}

	if len(received) != 0 {
		t.Errorf("Expected empty message, got length %d", len(received))
	}

	fmt.Println("Empty message test passed")
}

func TestAll(t *testing.T) {
	if (t.Run("TestSendMessageAndReadMessage", TestSendMessageAndReadMessage)) == false {
		t.Fail()
	}

	fmt.Println("TestSendMessageAndReadMessage passed")

	if (t.Run("TestMaxMessageSize", TestMaxMessageSize)) == false {
		t.Fail()
	}

	fmt.Println("TestMaxMessageSize passed")

	if (t.Run("TestEmptyMessage", TestEmptyMessage)) == false {
		t.Fail()
	}

	fmt.Println("TestEmptyMessage passed")
}
