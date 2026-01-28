package conduit

import "fmt"

var (
	MaxMessageSize = 10 << 20 // 10MB
)

var (
	ErrMessageTooLarge = fmt.Errorf("Message size exceeds maximum allowed size of %d bytes", MaxMessageSize)
)
