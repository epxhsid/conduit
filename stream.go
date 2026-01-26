package portal

type Stream struct {
	id      uint32
	mux     *Multiplexer
	recvBuf chan []byte
	closed  bool
}

func (s *Stream) Write(payload []byte) error {
	if s.closed {
		return ErrStreamClosed
	}

	frame := Frame{
		StreamID: s.id,
		Payload:  payload,
	}

	s.mux.outgoingQueue <- frame
	return nil
}
