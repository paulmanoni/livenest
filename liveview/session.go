package liveview

import (
	"sync"
)

// Session manages LiveView session state
type Session struct {
	mu     sync.RWMutex
	Data   map[string]interface{}
	Flashes map[string]string
}

// NewSession creates a new session
func NewSession() *Session {
	return &Session{
		Data:    make(map[string]interface{}),
		Flashes: make(map[string]string),
	}
}

// Put stores a value in the session
func (s *Session) Put(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data[key] = value
}

// Get retrieves a value from the session
func (s *Session) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Data[key]
	return val, ok
}

// Delete removes a value from the session
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Data, key)
}

// PutFlash sets a flash message
func (s *Session) PutFlash(key, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Flashes[key] = message
}

// GetFlash retrieves and clears a flash message
func (s *Session) GetFlash(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg, ok := s.Flashes[key]
	if ok {
		delete(s.Flashes, key)
	}
	return msg, ok
}

// Clear clears all session data
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data = make(map[string]interface{})
	s.Flashes = make(map[string]string)
}
