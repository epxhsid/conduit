package portal

import (
	"net"
	"sync"
)

type Multiplexer struct {
	conn         net.Conn
	streams      map[uint32]*Stream
	mu           sync.Mutex
	nextStreamID uint32
}

func NewMultiplexer(conn net.Conn) *Multiplexer {
	return &Multiplexer{
		conn:         conn,
		streams:      make(map[uint32]*Stream),
		nextStreamID: 1,
	}
}

func (m *Multiplexer) OpenStream() *Stream {
	m.mu.Lock()         // Lock to protect access to nextStreamID and streams map
	defer m.mu.Unlock() // Unlock after the function completes

	stream := &Stream{
		id:      m.nextStreamID,
		mux:     m,
		recvBuf: make(chan []byte, 10),
		closed:  false,
	}

	m.streams[m.nextStreamID] = stream
	m.nextStreamID += 2 // Increment by 2 to maintain odd/even stream ID separation
	return stream
}
