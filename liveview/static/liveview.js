// LiveNest LiveView Client
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
        const wsUrl = `${protocol}//${window.location.host}/live/ws/${this.componentName}?socket_id=${this.socketId}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            if (msg.type === 'render') {
                this.container.innerHTML = msg.data.html;
                this.attachEventListeners();

                // Handle flash messages if present
                if (msg.data.flash) {
                    this.showFlash(msg.data.flash);
                }
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
        // Handle lv-click events
        const clickElements = this.container.querySelectorAll('[lv-click]');
        clickElements.forEach(el => {
            const event = el.getAttribute('lv-click');
            el.addEventListener('click', (e) => {
                e.preventDefault();
                const payload = this.getPayloadFromElement(el);
                this.pushEvent(event, payload);
            });
        });

        // Handle lv-change events
        const changeElements = this.container.querySelectorAll('[lv-change]');
        changeElements.forEach(el => {
            const event = el.getAttribute('lv-change');
            el.addEventListener('input', (e) => {
                const payload = this.getPayloadFromElement(el);
                payload.value = el.value;
                this.pushEvent(event, payload);
            });
        });

        // Handle lv-submit events
        const formElements = this.container.querySelectorAll('[lv-submit]');
        formElements.forEach(el => {
            const event = el.getAttribute('lv-submit');
            el.addEventListener('submit', (e) => {
                e.preventDefault();
                const payload = this.getPayloadFromElement(el);
                this.pushEvent(event, payload);
            });
        });
    }

    getPayloadFromElement(el) {
        const payload = {};
        // Collect all lv-value-* attributes
        Array.from(el.attributes).forEach(attr => {
            if (attr.name.startsWith('lv-value-')) {
                const key = attr.name.replace('lv-value-', '');
                payload[key] = attr.value;
            }
        });
        return payload;
    }

    pushEvent(event, payload) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                event: event,
                payload: payload
            }));
        }
    }

    showFlash(flash) {
        // Remove existing flash messages
        const existing = document.querySelectorAll('.lv-flash');
        existing.forEach(el => el.remove());

        // Create flash container
        const flashDiv = document.createElement('div');
        flashDiv.className = `lv-flash lv-flash-${flash.type || 'info'}`;
        flashDiv.innerHTML = `
            <span class="lv-flash-message">${flash.message}</span>
            <button class="lv-flash-close">&times;</button>
        `;

        // Add styles if not already present
        if (!document.getElementById('lv-flash-styles')) {
            const style = document.createElement('style');
            style.id = 'lv-flash-styles';
            style.textContent = `
                .lv-flash {
                    position: fixed;
                    top: 20px;
                    right: 20px;
                    padding: 15px 20px;
                    border-radius: 5px;
                    box-shadow: 0 4px 6px rgba(0,0,0,0.1);
                    display: flex;
                    align-items: center;
                    gap: 15px;
                    z-index: 9999;
                    animation: slideIn 0.3s ease-out;
                }
                @keyframes slideIn {
                    from { transform: translateX(100%); opacity: 0; }
                    to { transform: translateX(0); opacity: 1; }
                }
                .lv-flash-success {
                    background: #27ae60;
                    color: white;
                }
                .lv-flash-error {
                    background: #e74c3c;
                    color: white;
                }
                .lv-flash-info {
                    background: #3498db;
                    color: white;
                }
                .lv-flash-warning {
                    background: #f39c12;
                    color: white;
                }
                .lv-flash-close {
                    background: none;
                    border: none;
                    color: white;
                    font-size: 24px;
                    cursor: pointer;
                    padding: 0;
                    line-height: 1;
                }
            `;
            document.head.appendChild(style);
        }

        // Add to page
        document.body.appendChild(flashDiv);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            flashDiv.style.animation = 'slideIn 0.3s ease-out reverse';
            setTimeout(() => flashDiv.remove(), 300);
        }, 5000);

        // Close button
        flashDiv.querySelector('.lv-flash-close').addEventListener('click', () => {
            flashDiv.remove();
        });
    }

    // Expose pushEvent globally for custom usage
    static getInstance() {
        return window.liveSocket;
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
});
