package portal

import "errors"

const MaxMessageSize uint32 = 10 << 20

var (
	ErrStreamClosed = errors.New("stream closed")
)
