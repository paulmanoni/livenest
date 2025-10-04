package main

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/paulmanoni/livenest/liveview"
)

// ChatMessage represents a single chat message
type ChatMessage struct {
	ID        int
	Username  string
	Message   string
	Timestamp time.Time
}

// Global chat state (shared across all users)
var (
	chatMessages   = []ChatMessage{}
	chatMessagesMu sync.RWMutex
	nextMessageID  = 1
)

// ChatComponent demonstrates real-time chat with LiveView
type ChatComponent struct {
	liveview.TemplateComponent
}

// Mount initializes the chat component
func (ch *ChatComponent) Mount(socket *liveview.Socket) error {
	// Generate random username for this session
	username := fmt.Sprintf("User%d", time.Now().Unix()%1000)

	socket.Assign(map[string]interface{}{
		"username":   username,
		"newMessage": "",
		"messages":   getChatMessages(),
	})
	return nil
}

// HandleSend sends a new chat message
func (ch *ChatComponent) HandleSend(socket *liveview.Socket, payload map[string]interface{}) error {
	message, ok := payload["message"].(string)
	if !ok || message == "" {
		return nil
	}

	username := socket.Assigns["username"].(string)

	// Add message to global chat
	addChatMessage(username, message)

	// Update local state
	socket.Assign(map[string]interface{}{
		"newMessage": "",
		"messages":   getChatMessages(),
	})

	return nil
}

// HandleRefresh refreshes the chat messages
func (ch *ChatComponent) HandleRefresh(socket *liveview.Socket, payload map[string]interface{}) error {
	socket.Assign(map[string]interface{}{
		"messages": getChatMessages(),
	})
	return nil
}

// HandleClear clears all chat messages (admin action)
func (ch *ChatComponent) HandleClear(socket *liveview.Socket, payload map[string]interface{}) error {
	clearChatMessages()
	socket.Assign(map[string]interface{}{
		"messages": getChatMessages(),
	})
	socket.PutFlash("info", "Chat cleared")
	return nil
}

// Render returns the HTML for the chat component
func (ch *ChatComponent) Render(socket *liveview.Socket) (template.HTML, error) {
	username := socket.Assigns["username"].(string)
	messages := socket.Assigns["messages"].([]ChatMessage)

	html := `
		<div class="chat-app">
			<div class="chat-header">
				<h2>💬 Real-Time Chat</h2>
				<p class="username">You are: <strong>` + username + `</strong></p>
			</div>

			<div class="chat-messages" id="chatMessages">
	`

	if len(messages) == 0 {
		html += `<div class="empty-state">No messages yet. Start the conversation!</div>`
	} else {
		for _, msg := range messages {
			isOwnMessage := msg.Username == username
			messageClass := "message"
			if isOwnMessage {
				messageClass += " own-message"
			}

			html += fmt.Sprintf(`
				<div class="%s">
					<div class="message-header">
						<span class="message-username">%s</span>
						<span class="message-time">%s</span>
					</div>
					<div class="message-content">%s</div>
				</div>
			`, messageClass, msg.Username, msg.Timestamp.Format("15:04"), msg.Message)
		}
	}

	html += `
			</div>

			<div class="chat-input">
				<input
					type="text"
					id="messageInput"
					placeholder="Type a message..."
					lv-keyup="send"
					lv-key="Enter"
				/>
				<button lv-click="refresh" class="refresh-btn">🔄</button>
				<button lv-click="clear" class="clear-btn">Clear</button>
			</div>
		</div>

		<style>
			.chat-app {
				width: 100%;
				max-width: 800px;
				height: 600px;
				display: flex;
				flex-direction: column;
				font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
				background: white;
				border-radius: 10px;
				overflow: hidden;
			}
			.chat-header {
				background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
				color: white;
				padding: 20px;
				text-align: center;
			}
			.chat-header h2 {
				margin: 0 0 10px 0;
			}
			.username {
				margin: 0;
				opacity: 0.9;
				font-size: 14px;
			}
			.chat-messages {
				flex: 1;
				overflow-y: auto;
				padding: 20px;
				background: #f5f5f5;
			}
			.empty-state {
				text-align: center;
				color: #95a5a6;
				padding: 40px;
			}
			.message {
				margin-bottom: 15px;
				padding: 10px 15px;
				background: white;
				border-radius: 10px;
				max-width: 70%;
				box-shadow: 0 2px 4px rgba(0,0,0,0.1);
			}
			.message.own-message {
				margin-left: auto;
				background: #667eea;
				color: white;
			}
			.message-header {
				display: flex;
				justify-content: space-between;
				margin-bottom: 5px;
				font-size: 12px;
				opacity: 0.8;
			}
			.message.own-message .message-header {
				opacity: 0.9;
			}
			.message-username {
				font-weight: bold;
			}
			.message-content {
				font-size: 15px;
				word-wrap: break-word;
			}
			.chat-input {
				display: flex;
				padding: 20px;
				background: white;
				border-top: 1px solid #e0e0e0;
				gap: 10px;
			}
			.chat-input input {
				flex: 1;
				padding: 12px 15px;
				border: 2px solid #e0e0e0;
				border-radius: 25px;
				font-size: 15px;
				outline: none;
			}
			.chat-input input:focus {
				border-color: #667eea;
			}
			.refresh-btn, .clear-btn {
				padding: 12px 20px;
				border: none;
				border-radius: 25px;
				cursor: pointer;
				font-size: 14px;
				transition: background-color 0.3s;
			}
			.refresh-btn {
				background: #3498db;
				color: white;
			}
			.refresh-btn:hover {
				background: #2980b9;
			}
			.clear-btn {
				background: #e74c3c;
				color: white;
			}
			.clear-btn:hover {
				background: #c0392b;
			}
		</style>

		<script>
			// Auto-scroll to bottom
			const chatMessages = document.getElementById('chatMessages');
			if (chatMessages) {
				chatMessages.scrollTop = chatMessages.scrollHeight;
			}

			// Handle Enter key for sending messages
			const messageInput = document.getElementById('messageInput');
			if (messageInput) {
				messageInput.addEventListener('keyup', function(e) {
					if (e.key === 'Enter' && this.value.trim()) {
						window.liveSocket.pushEvent('send', { message: this.value.trim() });
						this.value = '';
					}
				});
			}

			// Auto-refresh every 3 seconds to get new messages
			setInterval(() => {
				if (window.liveSocket) {
					window.liveSocket.pushEvent('refresh', {});
				}
			}, 3000);
		</script>
	`

	return template.HTML(html), nil
}

// Global chat message management
func addChatMessage(username, message string) {
	chatMessagesMu.Lock()
	defer chatMessagesMu.Unlock()

	chatMessages = append(chatMessages, ChatMessage{
		ID:        nextMessageID,
		Username:  username,
		Message:   message,
		Timestamp: time.Now(),
	})
	nextMessageID++

	// Keep only last 50 messages
	if len(chatMessages) > 50 {
		chatMessages = chatMessages[len(chatMessages)-50:]
	}
}

func getChatMessages() []ChatMessage {
	chatMessagesMu.RLock()
	defer chatMessagesMu.RUnlock()

	// Return a copy
	messages := make([]ChatMessage, len(chatMessages))
	copy(messages, chatMessages)
	return messages
}

func clearChatMessages() {
	chatMessagesMu.Lock()
	defer chatMessagesMu.Unlock()
	chatMessages = []ChatMessage{}
	nextMessageID = 1
}
