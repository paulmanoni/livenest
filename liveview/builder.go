package liveview

import (
	"github.com/gin-gonic/gin"
)

// HandlerBuilder provides a fluent API for building LiveView handlers
type HandlerBuilder struct {
	handler    *Handler
	path       string
	components []Component
	isLive     bool
}

// NewHandlerBuilder creates a new handler builder
func NewHandlerBuilder(handler *Handler) *HandlerBuilder {
	return &HandlerBuilder{
		handler:    handler,
		components: make([]Component, 0),
	}
}

// Path sets the route path
func (b *HandlerBuilder) Path(path string) *HandlerBuilder {
	b.path = path
	return b
}

// AsLive marks this handler as a LiveView route
func (b *HandlerBuilder) AsLive() *HandlerBuilder {
	b.isLive = true
	return b
}

// AddComponent adds a component to this route
func (b *HandlerBuilder) AddComponent(component Component) *HandlerBuilder {
	b.components = append(b.components, component)
	return b
}

// Build registers the route and returns a gin.HandlerFunc
func (b *HandlerBuilder) Build() gin.HandlerFunc {
	if !b.isLive || len(b.components) == 0 {
		return func(c *gin.Context) {
			c.JSON(400, gin.H{"error": "Invalid LiveView configuration"})
		}
	}

	// Register the first component with the path as name
	componentName := b.path
	if componentName == "/" {
		componentName = "index"
	}

	// For now, use the first component (can be extended to support multiple)
	b.handler.Register(componentName, b.components[0])

	return b.handler.HandleHTTP(componentName)
}

// BuildWebSocket builds the WebSocket handler for this LiveView
func (b *HandlerBuilder) BuildWebSocket() gin.HandlerFunc {
	return b.handler.HandleWebSocket
}
