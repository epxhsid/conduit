package proto

import "errors"

const (
	Version uint16 = 0x0001
)

const HeaderSize = 4

const (
	MaxFrameSize  = 64 * 1024
	MaxDomainSize = 255
)

const (
	CmdHandshake uint16 = 0x0001
	CmdData      uint16 = 0x0002
	CmdPing      uint16 = 0x0003
	CmdPong      uint16 = 0x0004
	CmdClose     uint16 = 0x0005
	CmdError     uint16 = 0x0006
)

var (
	ErrInvalidVersion     = errors.New("invalid protocol version")
	ErrInvalidType        = errors.New("invalid command type")
	ErrFrameTooLarge      = errors.New("frame payload exceeds maximum allowed size")
	ErrDomainTooLarge     = errors.New("domain name exceeds maximum allowed size")
	ErrInvalidPayload     = errors.New("invalid command payload")
	ErrUnexpectedEOF      = errors.New("unexpected end of file")
	ErrInternalError      = errors.New("internal server error")
	ErrUnsupportedCommand = errors.New("unsupported command type")
	ErrInvalidHandshake   = errors.New("invalid handshake payload")
	ErrInvalidData        = errors.New("invalid data payload")
)
