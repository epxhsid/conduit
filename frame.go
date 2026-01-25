package portal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// Framing:
// [4 bytes length][message bytes]
// big-endian uint32 (4 bytes) length prefix, standard for network protocols
// maximum message size is 10 MB
type Frame struct {
	StreamID uint32
	Payload  []byte
}

func WriteFrame(conn net.Conn, f *Frame) error {
	length := uint32(len(f.Payload))
	if length > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", length, MaxMessageSize)
	}

	if err := binary.Write(conn, binary.BigEndian, f.StreamID); err != nil {
		return err
	}

	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	written := 0
	for written < len(f.Payload) {
		n, err := conn.Write(f.Payload[written:])
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

func ReadFrame(conn net.Conn) (Frame, error) {
	var frame Frame
	var length uint32

	if err := binary.Read(conn, binary.BigEndian, &frame.StreamID); err != nil {
		return frame, err
	}

	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return frame, err
	}

	if length > MaxMessageSize {
		return frame, fmt.Errorf("message length %d exceeds maximum %d", length, MaxMessageSize)
	}

	if length == 0 {
		return frame, nil
	}

	frame.Payload = make([]byte, length)
	_, err := io.ReadFull(conn, frame.Payload)
	return frame, err
}
