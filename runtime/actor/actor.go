// Package actor implements the Actor Concurrency Model for NilLang.
// Actors maintain isolated state and communicate strictly via asynchronous message passing.
package actor

import (
	"context"
	"fmt"
	"sync"
)

// Message represents an envelope sent to an Actor.
type Message struct {
	Type    string
	Payload interface{}
	ReplyTo chan interface{}
}

// Actor encapsulates private state, a mailbox channel, and message handlers.
type Actor struct {
	Name     string
	mailbox  chan Message
	handlers map[string]func(msg Message) interface{}
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

// New creates a new named Actor with a buffered mailbox.
func New(name string, mailboxCapacity int) *Actor {
	if mailboxCapacity <= 0 {
		mailboxCapacity = 64
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Actor{
		Name:     name,
		mailbox:  make(chan Message, mailboxCapacity),
		handlers: make(map[string]func(msg Message) interface{}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// On registers a message handler for a specific message type.
func (a *Actor) On(msgType string, handler func(msg Message) interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handlers[msgType] = handler
}

// Start spawns the actor's event processing loop in a goroutine.
func (a *Actor) Start() {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.mu.Unlock()

	go a.loop()
}

func (a *Actor) loop() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case msg, ok := <-a.mailbox:
			if !ok {
				return
			}
			a.process(msg)
		}
	}
}

func (a *Actor) process(msg Message) {
	a.mu.RLock()
	handler, exists := a.handlers[msg.Type]
	a.mu.RUnlock()

	if exists && handler != nil {
		res := handler(msg)
		if msg.ReplyTo != nil {
			msg.ReplyTo <- res
		}
	} else if msg.ReplyTo != nil {
		msg.ReplyTo <- nil
	}
}

// Send dispatches an asynchronous one-way message (fire-and-forget).
func (a *Actor) Send(msgType string, payload interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.running {
		return fmt.Errorf("actor %s is not running", a.Name)
	}
	a.mailbox <- Message{Type: msgType, Payload: payload}
	return nil
}

// Ask sends a message and synchronously awaits a reply.
func (a *Actor) Ask(msgType string, payload interface{}) (interface{}, error) {
	a.mu.RLock()
	if !a.running {
		a.mu.RUnlock()
		return nil, fmt.Errorf("actor %s is not running", a.Name)
	}
	a.mu.RUnlock()

	replyChan := make(chan interface{}, 1)
	a.mailbox <- Message{Type: msgType, Payload: payload, ReplyTo: replyChan}

	select {
	case <-a.ctx.Done():
		return nil, fmt.Errorf("actor %s stopped", a.Name)
	case res := <-replyChan:
		return res, nil
	}
}

// Stop terminates the actor.
func (a *Actor) Stop() {
	a.cancel()
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()
}
