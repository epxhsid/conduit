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

	testStreamID := uint32(1)
	testPayload := []byte("sample message stream")
	errChan := make(chan error, 1)

	go func() {
		errChan <- WriteFrame(client, &Frame{
			StreamID: testStreamID,
			Payload:  testPayload,
		})
	}()

	frame, err := ReadFrame(server)
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}

	if sendErr := <-errChan; sendErr != nil {
		t.Fatalf("WriteFrame failed: %v", sendErr)
	}

	if frame.StreamID != testStreamID {
		t.Errorf("Expected StreamID %d, got %d", testStreamID, frame.StreamID)
	}

	if !bytes.Equal(frame.Payload, testPayload) {
		t.Errorf("Expected %q, got %q", testPayload, frame.Payload)
	}

	fmt.Printf("Received message: %s\n", string(frame.Payload))
}

func TestMaxMessageSize(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	testStreamID := uint32(1)
	largeMessage := make([]byte, MaxMessageSize+1)

	err := WriteFrame(client, &Frame{
		StreamID: testStreamID,
		Payload:  largeMessage,
	})

	if err == nil {
		t.Error("Expected error for oversized message, got nil")
	}
}

func TestEmptyMessage(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	testStreamID := uint32(1)
	emptyMessage := []byte{}

	errChan := make(chan error, 1)
	go func() {
		errChan <- WriteFrame(client, &Frame{
			StreamID: testStreamID,
			Payload:  emptyMessage,
		})
	}()
	received, err := ReadFrame(server)
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("WriteFrame failed: %v", err)
	}

	if !bytes.Equal(received.Payload, emptyMessage) {
		t.Errorf("Expected %q, got %q", emptyMessage, received.Payload)
	}

	if len(received.Payload) != 0 {
		t.Errorf("Expected empty message, got length %d", len(received.Payload))
	}

	fmt.Println("Empty message test passed")
}

func TestAll(t *testing.T) {
	if (t.Run("TestSendMessageAndReadMessage", TestSendMessageAndReadMessage)) == false {
		t.Fail()
	}

	if (t.Run("TestMaxMessageSize", TestMaxMessageSize)) == false {
		t.Fail()
	}

	if (t.Run("TestEmptyMessage", TestEmptyMessage)) == false {
		t.Fail()
	}
}
