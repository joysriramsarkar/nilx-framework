// Package state provides fine-grained reactive state management (Signals, Store, Computed) for Alap UI components.
package state

import (
	"sync"
)

// Listener is a notification callback invoked on state changes.
type Listener func(newValue interface{})

// Signal holds a single reactive value and notifies subscribers when mutated.
type Signal struct {
	mu        sync.RWMutex
	value     interface{}
	listeners []Listener
}

// NewSignal creates a reactive state signal with an initial value.
func NewSignal(initial interface{}) *Signal {
	return &Signal{
		value:     initial,
		listeners: make([]Listener, 0),
	}
}

// Get reads the current value.
func (s *Signal) Get() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set updates the value and triggers registered listener callbacks.
func (s *Signal) Set(val interface{}) {
	s.mu.Lock()
	s.value = val
	listeners := make([]Listener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	for _, l := range listeners {
		l(val)
	}
}

// Subscribe attaches a listener that fires whenever the signal value changes.
func (s *Signal) Subscribe(l Listener) func() {
	s.mu.Lock()
	s.listeners = append(s.listeners, l)
	idx := len(s.listeners) - 1
	s.mu.Unlock()

	// Return unsubscribe function
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if idx < len(s.listeners) {
			s.listeners = append(s.listeners[:idx], s.listeners[idx+1:]...)
		}
	}
}

// Store represents a central reactive state repository with key-value signals.
type Store struct {
	mu      sync.RWMutex
	signals map[string]*Signal
}

// NewStore initializes an empty state store.
func NewStore() *Store {
	return &Store{
		signals: make(map[string]*Signal),
	}
}

// Define registers a named state field.
func (st *Store) Define(key string, initial interface{}) *Signal {
	st.mu.Lock()
	defer st.mu.Unlock()
	sig := NewSignal(initial)
	st.signals[key] = sig
	return sig
}

// Get returns the value of a named state field.
func (st *Store) Get(key string) interface{} {
	st.mu.RLock()
	defer st.mu.RUnlock()
	if sig, ok := st.signals[key]; ok {
		return sig.Get()
	}
	return nil
}

// Set updates the value of a named state field.
func (st *Store) Set(key string, val interface{}) {
	st.mu.RLock()
	sig, ok := st.signals[key]
	st.mu.RUnlock()
	if ok {
		sig.Set(val)
	}
}
