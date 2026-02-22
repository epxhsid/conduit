package proto

import "errors"

const (
	Version uint16 = 0x0001
)

const HeaderSize = 8

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

const (
	CmdMin uint16 = 0x0001
	CmdMax uint16 = 0x0006
)

var (
	ErrFrameTooLarge      = errors.New("frame too large")
	ErrUnsupportedCommand = errors.New("unsupported command")
	ErrInvalidHandshake   = errors.New("invalid handshake")
	ErrInvalidPayload     = errors.New("invalid payload")
	ErrDomainTooLarge     = errors.New("domain too large")
)
