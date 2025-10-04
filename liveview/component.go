package liveview

import (
	"html/template"
	"math/rand"
)

// Component represents a LiveView component
type Component interface {
	// Mount is called when the component is first loaded
	Mount(socket *Socket) error

	// Render returns the HTML template for this component
	Render(socket *Socket) (template.HTML, error)
}

// EventHandler is an optional interface for handling events
type EventHandler interface {
	HandleEvent(event string, payload map[string]interface{}, socket *Socket) error
}

// Socket represents a LiveView socket connection
type Socket struct {
	ID           string
	ComponentID  string
	Session      *Session
	Assigns      map[string]interface{}
	previousHTML string // Track previous render for diffing
}

// NewSocket creates a new socket
func NewSocket(id string) *Socket {
	return &Socket{
		ID:          id,
		ComponentID: generateComponentID(),
		Assigns:     make(map[string]interface{}),
		Session:     NewSession(),
	}
}

// generateComponentID generates a unique component ID
func generateComponentID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "lv-" + string(b)
}

// Assign sets multiple values in the socket assigns from a map
func (s *Socket) Assign(assigns map[string]interface{}) {
	for k, v := range assigns {
		s.Assigns[k] = v
	}
}

// Set sets a single value in the socket assigns
func (s *Socket) Set(key string, value interface{}) {
	s.Assigns[key] = value
}

// Get retrieves a value from socket assigns
func (s *Socket) Get(key string) (interface{}, bool) {
	val, ok := s.Assigns[key]
	return val, ok
}

// PutFlash sets a flash message
func (s *Socket) PutFlash(key, message string) {
	s.Session.PutFlash(key, message)
}

// GetFlash retrieves and clears a flash message
func (s *Socket) GetFlash(key string) (string, bool) {
	return s.Session.GetFlash(key)
}