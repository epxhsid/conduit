package conduit

import (
	"encoding/binary"
	"net"

	"io"
)

type Codec struct {
	streamID uint32
	payload  []byte
}

type ProtocolImpl interface {
	WriteFrame(conn net.Conn, c *Codec) error
	ReadFrame(conn net.Conn) (*Codec, error)
}

type Protocol struct{}

// Frame format:
// 4 bytes: stream ID (big-endian uint32)
// 4 bytes: payload length (big-endian uint32)
// N bytes: payload

// writeFrame writes a single framed message to the connection.
// this creates a 8-byte buffer for the header; this avoids multiple system calls
// by writing the header and payload in a single write operation.
func (p *Protocol) WriteFrame(conn net.Conn, c *Codec) error {
	length := uint32(len(c.payload))
	if length > uint32(MaxMessageSize) {
		return ErrMessageTooLarge
	}

	buf := make([]byte, 8+length)
	binary.BigEndian.PutUint32(buf[0:4], c.streamID)
	binary.BigEndian.PutUint32(buf[4:8], length)
	copy(buf[8:], c.payload)

	_, err := conn.Write(buf)
	return err
}

// readFrame reads a single framed message from the connection.
// it first reads the 8-byte header to determine the stream ID and payload length,
// then reads the payload based on the specified length.
func (p *Protocol) ReadFrame(conn net.Conn) (*Codec, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	streamID := binary.BigEndian.Uint32(header[0:4])
	length := binary.BigEndian.Uint32(header[4:8])

	if length > uint32(MaxMessageSize) {
		return nil, ErrMessageTooLarge
	}

	payload := make([]byte, length)

	if length > 0 {
		if _, err := io.ReadFull(conn, payload); err != nil {
			return nil, err
		}
	}

	return &Codec{
		streamID: streamID,
		payload:  payload,
	}, nil
}
