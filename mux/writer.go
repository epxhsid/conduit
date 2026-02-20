package mux

import (
	"io"
	"sync/atomic"
)

type CountingWriter struct {
	N atomic.Int64
	W io.Writer
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{W: w}
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.W.Write(p)
	c.N.Add(int64(n))
	return n, err
}

func (c *CountingWriter) Count() int64 {
	return c.N.Load()
}
