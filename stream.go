package portal

type Stream struct {
	id      uint32
	mux     *Multiplexer
	recvBuf chan []byte
	closed  bool
}
