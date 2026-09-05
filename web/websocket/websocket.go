package websocket

import (
	"encoding/json"
	"sync"
)

type Client struct {
	ID   string
	Send chan []byte
}

type Manager struct {
	clients map[string]*Client
	rooms   map[string]map[string]*Client
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
		rooms:   make(map[string]map[string]*Client),
	}
}

func (m *Manager) Register(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c.ID] = c
}

func (m *Manager) Unregister(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, c.ID)
	for r := range m.rooms {
		delete(m.rooms[r], c.ID)
	}
	close(c.Send)
}

func (m *Manager) Broadcast(payload interface{}) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.clients {
		select {
		case c.Send <- b:
		default:
		}
	}
}
