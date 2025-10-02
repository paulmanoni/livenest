package liveview

import (
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Handler manages LiveView WebSocket connections
type Handler struct {
	components map[string]Component
	sockets    map[string]*Socket
	mu         sync.RWMutex
}

// NewHandler creates a new LiveView handler
func NewHandler() *Handler {
	return &Handler{
		components: make(map[string]Component),
		sockets:    make(map[string]*Socket),
	}
}

// Register registers a component with a route
func (h *Handler) Register(name string, component Component) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.components[name] = component
}

// HandleWebSocket handles WebSocket connections for LiveView
func (h *Handler) HandleWebSocket(c *gin.Context) {
	componentName := c.Param("component")

	h.mu.RLock()
	component, exists := h.components[componentName]
	h.mu.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Component not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Create socket
	socket := NewSocket(c.Query("socket_id"))

	// Mount component
	if err := component.Mount(socket); err != nil {
		log.Printf("Component mount error: %v", err)
		return
	}

	// Store socket
	h.mu.Lock()
	h.sockets[socket.ID] = socket
	h.mu.Unlock()

	// Send initial render
	html, err := component.Render(socket)
	if err != nil {
		log.Printf("Render error: %v", err)
		return
	}

	renderData := map[string]interface{}{
		"html": string(html),
	}
	h.addFlashToData(socket, renderData)

	if err := h.sendMessage(conn, "render", renderData); err != nil {
		log.Printf("Send error: %v", err)
		return
	}

	// Listen for events
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle event - try reflection-based routing first, then EventHandler interface
		err := RouteEvent(component, msg.Event, msg.Payload, socket)
		if err != nil {
			// Fallback to EventHandler interface if routing fails
			if handler, ok := component.(EventHandler); ok {
				if err := handler.HandleEvent(msg.Event, msg.Payload, socket); err != nil {
					log.Printf("Event handling error: %v", err)
					continue
				}
			} else {
				log.Printf("Event handling error: %v", err)
				continue
			}
		}

		// Re-render
		html, err := component.Render(socket)
		if err != nil {
			log.Printf("Render error: %v", err)
			continue
		}

		renderData := map[string]interface{}{
			"html": string(html),
		}
		h.addFlashToData(socket, renderData)

		if err := h.sendMessage(conn, "render", renderData); err != nil {
			log.Printf("Send error: %v", err)
			break
		}
	}

	// Cleanup
	h.mu.Lock()
	delete(h.sockets, socket.ID)
	h.mu.Unlock()
}

// Message represents a WebSocket message
type Message struct {
	Event   string                 `json:"event"`
	Payload map[string]interface{} `json:"payload"`
}

// sendMessage sends a message to the WebSocket client
func (h *Handler) sendMessage(conn *websocket.Conn, msgType string, data map[string]interface{}) error {
	msg := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	return conn.WriteJSON(msg)
}

// addFlashToData adds flash messages from socket to render data
func (h *Handler) addFlashToData(socket *Socket, data map[string]interface{}) {
	// Check for all flash types
	flashTypes := []string{"success", "error", "info", "warning"}
	for _, flashType := range flashTypes {
		if msg, ok := socket.GetFlash(flashType); ok {
			data["flash"] = map[string]string{
				"type":    flashType,
				"message": msg,
			}
			break // Only send one flash message at a time
		}
	}
}

// HandleComponentTag handles requests from <component> tags
func (h *Handler) HandleComponentTag(c *gin.Context) {
	componentName := c.Param("name")

	h.mu.RLock()
	component, exists := h.components[componentName]
	h.mu.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Component not found"})
		return
	}

	// Create temporary socket for initial render
	socket := NewSocket("")

	if err := component.Mount(socket); err != nil {
		c.JSON(500, gin.H{"error": "Mount failed"})
		return
	}

	html, err := component.Render(socket)
	if err != nil {
		c.JSON(500, gin.H{"error": "Render failed"})
		return
	}

	// Generate socket ID
	socketID := generateSocketID()

	// Return JSON for component tag
	c.JSON(200, gin.H{
		"html":         string(html),
		"socket_id":    socketID,
		"component_id": socket.ComponentID,
	})
}

// HandleHTTP handles initial HTTP request and serves the LiveView page
func (h *Handler) HandleHTTP(componentName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.mu.RLock()
		component, exists := h.components[componentName]
		h.mu.RUnlock()

		if !exists {
			c.JSON(404, gin.H{"error": "Component not found"})
			return
		}

		// Create temporary socket for initial render
		socket := NewSocket("")

		if err := component.Mount(socket); err != nil {
			c.JSON(500, gin.H{"error": "Mount failed"})
			return
		}

		html, err := component.Render(socket)
		if err != nil {
			c.JSON(500, gin.H{"error": "Render failed"})
			return
		}

		// Generate socket ID
		socketID := generateSocketID()

		// Serve full HTML page with LiveView wrapper
		htmlWrapper := generateHTMLWrapper(componentName, string(html), socketID, socket.ComponentID)
		c.Data(200, "text/html; charset=utf-8", []byte(htmlWrapper))
	}
}

// generateSocketID generates a unique socket ID
func generateSocketID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "socket_" + string(b)
}

// generateHTMLWrapper generates the full HTML page with LiveView JavaScript
func generateHTMLWrapper(componentName, componentHTML, socketID, componentID string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LiveNest - ` + componentName + `</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .liveview-container {
            background: white;
            border-radius: 15px;
            padding: 40px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        }
    </style>
    <script src="/livenest/liveview.js"></script>
</head>
<body>
    <div class="liveview-container">
        <div id="liveview" data-component="` + componentName + `" data-socket-id="` + socketID + `" data-component-id="` + componentID + `">` + componentHTML + `</div>
    </div>
</body>
</html>`
}
