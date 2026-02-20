package mux

import (
	"time"

	"github.com/hashicorp/yamux"
)

type StreamState struct {
	ID        string
	Stream    *yamux.Stream
	StartedAt time.Time
}

func NewStreamState(id string, stream *yamux.Stream) *StreamState {
	return &StreamState{
		ID:        id,
		Stream:    stream,
		StartedAt: time.Now(),
	}
}

func (s *StreamState) Close() error {
	return s.Stream.Close()
}

func (s *StreamState) IsActive() bool {
	return time.Since(s.StartedAt) < 5*time.Minute
}
