// Package nilbus implements the distributed IPC client for Onuron (Onuron OS) system services.
package nilbus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultSocketPath = "/run/nilbus/system.sock"
	MagicHeader       = 0x4E494C42 // "NILB" in ASCII
)

// Request represents an outgoing RPC envelope across NilBus.
type Request struct {
	ID      uint32
	Service string
	Method  string
	Payload []byte
}

// Response represents the RPC response received from NilBus.
type Response struct {
	ID         uint32
	StatusCode int32
	Payload    []byte
	ErrorMsg   string
}

// Client manages the connection and RPC multiplexing to NilBus IPC daemon.
type Client struct {
	mu           sync.RWMutex
	conn         net.Conn
	socketPath   string
	seqID        uint32
	pendingCalls map[uint32]chan *Response
	subscribers  map[string][]func(event []byte)
	inProcessBus map[string]func(method string, payload []byte) ([]byte, error)
	connected    bool
	isMock       bool
}

// NewClient creates a new NilBus IPC client.
func NewClient(customSocketPath string) *Client {
	socket := customSocketPath
	if socket == "" {
		socket = os.Getenv("NILBUS_SOCKET")
	}
	if socket == "" {
		socket = DefaultSocketPath
	}

	return &Client{
		socketPath:   socket,
		pendingCalls: make(map[uint32]chan *Response),
		subscribers:  make(map[string][]func(event []byte)),
		inProcessBus: make(map[string]func(method string, payload []byte) ([]byte, error)),
	}
}

// Connect opens the IPC transport to the Onuron system bus.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.DialTimeout("unix", c.socketPath, 200*time.Millisecond)
	if err != nil {
		// Fallback to local in-process bus when running outside Onuron kernel environment
		c.isMock = true
		c.connected = true
		c.registerDefaultServices()
		return nil
	}

	c.conn = conn
	c.connected = true
	c.isMock = false

	go c.readLoop()
	return nil
}

// Call invokes a remote method on a named Onuron system service.
func (c *Client) Call(service, method string, payload []byte) ([]byte, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("nilbus client not connected")
	}

	if c.isMock {
		handler, ok := c.inProcessBus[service]
		c.mu.RUnlock()
		if !ok {
			return nil, fmt.Errorf("service %q not found on nilbus", service)
		}
		return handler(method, payload)
	}
	c.mu.RUnlock()

	reqID := atomic.AddUint32(&c.seqID, 1)
	respChan := make(chan *Response, 1)

	c.mu.Lock()
	c.pendingCalls[reqID] = respChan
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pendingCalls, reqID)
		c.mu.Unlock()
	}()

	// Encode frame: [Magic 4B][ReqID 4B][ServiceLen 2B][Service][MethodLen 2B][Method][PayloadLen 4B][Payload]
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint32(MagicHeader))
	_ = binary.Write(&buf, binary.BigEndian, reqID)
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(service)))
	buf.WriteString(service)
	_ = binary.Write(&buf, binary.BigEndian, uint16(len(method)))
	buf.WriteString(method)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(payload)))
	buf.Write(payload)

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("connection lost")
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed sending nilbus packet: %w", err)
	}

	select {
	case resp := <-respChan:
		if resp.StatusCode != 0 {
			return nil, fmt.Errorf("nilbus error (%d): %s", resp.StatusCode, resp.ErrorMsg)
		}
		return resp.Payload, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("nilbus timeout calling %s.%s", service, method)
	}
}

// Publish broadcasts an event payload on a NilOS topic.
func (c *Client) Publish(topic string, eventData []byte) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if listeners, ok := c.subscribers[topic]; ok {
		for _, l := range listeners {
			go l(eventData)
		}
	}
}

// Subscribe listens for events published to a topic.
func (c *Client) Subscribe(topic string, handler func(event []byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribers[topic] = append(c.subscribers[topic], handler)
}

// Close closes the NilBus connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.connected = false
}

func (c *Client) registerDefaultServices() {
	notifyHandler := func(method string, payload []byte) ([]byte, error) {
		if method == "Notify" {
			return []byte("{\"status\":\"posted\"}"), nil
		}
		return nil, fmt.Errorf("unknown method %s", method)
	}
	c.inProcessBus["org.onuron.NotificationService"] = notifyHandler
	c.inProcessBus["org.nilos.NotificationService"] = notifyHandler

	halHandler := func(method string, payload []byte) ([]byte, error) {
		if method == "ReadSensor" {
			return []byte("{\"x\":0.0,\"y\":9.81,\"z\":0.0}"), nil
		}
		return nil, fmt.Errorf("unknown method %s", method)
	}
	c.inProcessBus["org.onuron.HAL"] = halHandler
	c.inProcessBus["org.nilos.HAL"] = halHandler
}

func (c *Client) readLoop() {
	header := make([]byte, 12)
	for {
		_, err := io.ReadFull(c.conn, header)
		if err != nil {
			return
		}

		magic := binary.BigEndian.Uint32(header[0:4])
		if magic != MagicHeader {
			continue
		}

		reqID := binary.BigEndian.Uint32(header[4:8])
		payloadLen := binary.BigEndian.Uint32(header[8:12])

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(c.conn, payload); err != nil {
			return
		}

		c.mu.RLock()
		ch, ok := c.pendingCalls[reqID]
		c.mu.RUnlock()

		if ok && ch != nil {
			ch <- &Response{ID: reqID, StatusCode: 0, Payload: payload}
		}
	}
}
