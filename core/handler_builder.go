package core

import (
	"fmt"
	"log"

	"livenest/liveview"

	"github.com/gin-gonic/gin"
)

// HandlerBuilder provides a fluent API for building routes
type HandlerBuilder struct {
	app              *App
	path             string
	method           string
	handler          gin.HandlerFunc
	components       []liveview.Component
	componentNames   []string
	primaryComponent string
	isLive           bool
}

// NewHandler creates a new handler builder
func (a *App) NewHandler() *HandlerBuilder {
	if a.lvHandler == nil {
		a.lvHandler = liveview.NewHandler()
	}

	return &HandlerBuilder{
		app:        a,
		method:     "GET",
		components: make([]liveview.Component, 0),
	}
}

// Path sets the route path
func (b *HandlerBuilder) Path(path string) *HandlerBuilder {
	b.path = path
	return b
}

// AsGet sets the HTTP method to GET
func (b *HandlerBuilder) AsGet() *HandlerBuilder {
	b.method = "GET"
	return b
}

// AsPost sets the HTTP method to POST
func (b *HandlerBuilder) AsPost() *HandlerBuilder {
	b.method = "POST"
	return b
}

// AsPut sets the HTTP method to PUT
func (b *HandlerBuilder) AsPut() *HandlerBuilder {
	b.method = "PUT"
	return b
}

// AsDelete sets the HTTP method to DELETE
func (b *HandlerBuilder) AsDelete() *HandlerBuilder {
	b.method = "DELETE"
	return b
}

// AsPatch sets the HTTP method to PATCH
func (b *HandlerBuilder) AsPatch() *HandlerBuilder {
	b.method = "PATCH"
	return b
}

// AsLive marks this handler as a LiveView route
func (b *HandlerBuilder) AsLive() *HandlerBuilder {
	b.isLive = true
	return b
}

// Func sets the handler function for regular routes
func (b *HandlerBuilder) Func(handler gin.HandlerFunc) *HandlerBuilder {
	b.handler = handler
	return b
}

// AddComponent adds a LiveView component with optional name
// If name is provided after the component, it will be registered with that name
// Example: .AddComponent(&Counter{}).WithName("counter")
func (b *HandlerBuilder) AddComponent(component liveview.Component) *ComponentAdder {
	return &ComponentAdder{
		builder:   b,
		component: component,
	}
}

// ComponentAdder allows chaining WithName after AddComponent
type ComponentAdder struct {
	builder   *HandlerBuilder
	component liveview.Component
}

// WithName sets a custom name for this component and returns the builder
func (ca *ComponentAdder) WithName(name string) *HandlerBuilder {
	ca.builder.components = append(ca.builder.components, ca.component)
	ca.builder.componentNames = append(ca.builder.componentNames, name)
	return ca.builder
}

// AddComponent chains another component (when WithName is not called)
func (ca *ComponentAdder) AddComponent(component liveview.Component) *ComponentAdder {
	// Add current component without explicit name
	ca.builder.components = append(ca.builder.components, ca.component)
	ca.builder.componentNames = append(ca.builder.componentNames, "")
	return ca.builder.AddComponent(component)
}

// Build finalizes without WithName (used when component name is derived from path)
func (ca *ComponentAdder) Build() {
	ca.builder.components = append(ca.builder.components, ca.component)
	ca.builder.componentNames = append(ca.builder.componentNames, "")
	ca.builder.Build()
}

// AsLive marks as LiveView and allows continuing the chain
func (ca *ComponentAdder) AsLive() *ComponentAdder {
	ca.builder.isLive = true
	return ca
}

// Path is a convenience method to continue building
func (ca *ComponentAdder) Path(path string) *ComponentAdder {
	ca.builder.path = path
	return ca
}

// Build registers the route with the app
func (b *HandlerBuilder) Build() {
	if b.path == "" {
		b.path = "/"
	}

	if b.isLive {
		b.buildLiveView()
	} else {
		b.buildRegular()
	}
}

// buildRegular builds a regular HTTP route
func (b *HandlerBuilder) buildRegular() {
	if b.handler == nil {
		return
	}

	switch b.method {
	case "GET":
		b.app.GET(b.path, b.handler)
	case "POST":
		b.app.POST(b.path, b.handler)
	case "PUT":
		b.app.PUT(b.path, b.handler)
	case "DELETE":
		b.app.DELETE(b.path, b.handler)
	case "PATCH":
		b.app.PATCH(b.path, b.handler)
	}
}

// buildLiveView builds a LiveView route
func (b *HandlerBuilder) buildLiveView() {
	if len(b.components) == 0 {
		return
	}

	// Ensure LiveView handler exists
	if b.app.lvHandler == nil {
		b.app.lvHandler = liveview.NewHandler()
	}

	// Determine primary component name (for the route)
	primaryName := b.primaryComponent
	if primaryName == "" && len(b.componentNames) > 0 && b.componentNames[0] != "" {
		primaryName = b.componentNames[0]
	}
	if primaryName == "" {
		primaryName = b.path
		if primaryName == "/" {
			primaryName = "index"
		}
	}

	// Register all components with their names
	var registeredNames []string
	for i, component := range b.components {
		name := ""
		if i < len(b.componentNames) && b.componentNames[i] != "" {
			name = b.componentNames[i]
		} else {
			// Derive name from path if not specified
			name = primaryName
			if i > 0 {
				name = fmt.Sprintf("%s_%d", primaryName, i)
			}
		}

		b.app.lvHandler.Register(name, component)
		registeredNames = append(registeredNames, name)
	}

	// Register HTTP handler (uses first component)
	b.app.GET(b.path, b.app.lvHandler.HandleHTTP(primaryName))

	// Register WebSocket handlers for all components
	for _, name := range registeredNames {
		wsPath := "/live/ws/" + name
		componentName := name // capture for closure
		b.app.GET(wsPath, func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "component", Value: componentName})
			b.app.lvHandler.HandleWebSocket(c)
		})
	}

	log.Printf("LiveView registered: %s (Components: %v)", b.path, registeredNames)
}
