package websocket

import (
	"testing"
)

func TestWebSocketManager(t *testing.T) {
	mgr := NewManager()
	c1 := &Client{ID: "c1", Send: make(chan []byte, 10)}
	c2 := &Client{ID: "c2", Send: make(chan []byte, 10)}

	mgr.Register(c1)
	mgr.Register(c2)

	mgr.Broadcast(map[string]string{"type": "ping"})

	select {
	case <-c1.Send:
		// OK
	default:
		t.Errorf("c1 did not receive broadcast")
	}

	mgr.Unregister(c1)
}
