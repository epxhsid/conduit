package portal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const MaxMessageSize uint32 = 10 << 20

// Framing:
// [4 bytes length][message bytes]
// big-endian uint32 (4 bytes) length prefix, standard for network protocols
// maximum message size is 10 MB
func SendMessage(conn net.Conn, msg []byte) error {
	length := uint32(len(msg))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	if length > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max %d)", length, MaxMessageSize)
	}

	written := 0
	for written < len(msg) {
		n, err := conn.Write(msg[written:])
		fmt.Printf("sent: %d bytes\n", n)
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

func ReadMessage(conn net.Conn) ([]byte, error) {
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length > MaxMessageSize {
		return nil, fmt.Errorf("message length %d exceeds maximum %d", length, MaxMessageSize)
	}

	if length == 0 {
		return []byte{}, nil
	}

	msg := make([]byte, length)
	_, err := io.ReadFull(conn, msg)
	return msg, err
}
