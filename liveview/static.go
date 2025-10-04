package liveview

import (
	_ "embed"
	"strings"
)

//go:embed static/liveview.js
var liveviewJS string

// GetLiveViewJS returns the LiveView client JavaScript
func GetLiveViewJS() string {
	// Combine LiveView socket + Component tag
	var js strings.Builder
	js.WriteString(liveviewJS)
	js.WriteString("\n\n")
	js.WriteString(GetComponentTagJS())
	return js.String()
}

// getLiveViewSocketJS returns just the socket implementation (deprecated, kept for compatibility)
func getLiveViewSocketJS() string {
	return `// LiveNest LiveView Client
class LiveViewSocket {
    constructor(componentName, socketId) {
        this.componentName = componentName;
        this.socketId = socketId;
        this.ws = null;
        this.container = document.getElementById('liveview');
    }

    connect() {
        this.attachEventListeners();
        this.connectWebSocket();
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = protocol + '//' + window.location.host + '/live/ws/' + this.componentName + '?socket_id=' + this.socketId;

        this.ws = new WebSocket(wsUrl);

        this.ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            if (msg.type === 'render') {
                this.container.innerHTML = msg.data.html;
                this.attachEventListeners();
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket closed, reconnecting...');
            setTimeout(() => this.connectWebSocket(), 1000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    attachEventListeners() {
        const elements = this.container.querySelectorAll('[lv-click]');
        elements.forEach(el => {
            const event = el.getAttribute('lv-click');
            el.addEventListener('click', () => {
                this.sendEvent(event, {});
            });
        });
    }

    sendEvent(event, payload) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                event: event,
                payload: payload
            }));
        }
    }
}

// Auto-initialize if liveview container exists
window.addEventListener('DOMContentLoaded', () => {
    const container = document.getElementById('liveview');
    if (container && container.dataset.component && container.dataset.socketId) {
        const liveview = new LiveViewSocket(
            container.dataset.component,
            container.dataset.socketId
        );
        liveview.connect();
    }
});`
}
