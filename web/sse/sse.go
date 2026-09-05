package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Stream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewStream(w http.ResponseWriter) (*Stream, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("flusher not supported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	return &Stream{w: w, flusher: flusher}, nil
}

func (s *Stream) Emit(event string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(b))
	if _, err := io.WriteString(s.w, msg); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}
