package mux

import (
	"io"
	"sync/atomic"
)

type CountingWriter struct {
	n *atomic.Int64
	w io.Writer
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n.Add(int64(n))
	return n, err
}
