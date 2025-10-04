// LiveNest LiveView Client
class LiveViewSocket {
    constructor(componentName, socketId) {
        this.componentName = componentName;
        this.socketId = socketId;
        this.ws = null;
        this.container = document.getElementById('liveview');
        this.debounceTimers = new Map(); // Store debounce timers per element
        this.focusedInput = null; // Track currently focused input
        this.cursorPosition = null; // Track cursor position
        this.inputStates = new Map(); // Track input values and cursor positions
        this.pendingInputs = new Set(); // Track inputs with pending server updates

        // Track focus/blur on inputs
        this.setupFocusTracking();

        // Expose globally immediately for form handlers
        window.liveSocket = this;
        // Dispatch event so form scripts know liveSocket is ready
        window.dispatchEvent(new CustomEvent('liveSocketReady'));
    }

    setupFocusTracking() {
        // Use event delegation to track focus on all inputs
        document.addEventListener('focusin', (e) => {
            const target = e.target;
            if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT') {
                this.focusedInput = target;
                // Capture initial state when focused
                this.captureInputState(target);
            }
        }, true);

        document.addEventListener('focusout', (e) => {
            const target = e.target;
            if (target === this.focusedInput) {
                // On blur, allow server updates to be applied
                this.pendingInputs.delete(target);
                this.inputStates.delete(target);
                this.focusedInput = null;
                this.cursorPosition = null;
            }
        }, true);

        // Track changes as user types
        document.addEventListener('input', (e) => {
            const target = e.target;
            if ((target.tagName === 'INPUT' || target.tagName === 'TEXTAREA')) {
                this.captureInputState(target);
                // Mark as having pending changes
                this.pendingInputs.add(target);
            }
        }, true);

        document.addEventListener('selectionchange', () => {
            if (this.focusedInput && (this.focusedInput.tagName === 'INPUT' || this.focusedInput.tagName === 'TEXTAREA')) {
                this.cursorPosition = this.focusedInput.selectionStart;
            }
        });
    }

    captureInputState(input) {
        if (input.tagName === 'INPUT' || input.tagName === 'TEXTAREA') {
            this.inputStates.set(input, {
                value: input.value,
                selectionStart: input.selectionStart,
                selectionEnd: input.selectionEnd
            });
            this.cursorPosition = input.selectionStart;
        }
    }

    restoreInputState(input) {
        const state = this.inputStates.get(input);
        if (state && input === this.focusedInput) {
            input.value = state.value;
            if (state.selectionStart !== null) {
                try {
                    input.setSelectionRange(state.selectionStart, state.selectionEnd);
                } catch (e) {
                    // Ignore errors for input types that don't support selection
                }
            }
        }
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
                // Handle diff-based updates (Phoenix LiveView style)
                if (msg.data.diff) {
                    this.applyDiff(msg.data.diff);
                } else if (msg.data.html) {
                    // Full HTML replacement (initial render)
                    this.patch(msg.data.html);
                }

                // Handle flash messages if present
                if (msg.data.flash) {
                    this.showFlash(msg.data.flash);
                }
            }
        };

        this.ws.onopen = () => {
            // WebSocket connected
        };

        this.ws.onclose = (event) => {
            setTimeout(() => this.connectWebSocket(), 1000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    attachEventListeners() {
        // Remove old listeners by cloning and replacing nodes (simple approach)
        // Mark elements so we don't re-attach listeners
        const clickElements = this.container.querySelectorAll('[lv-click]');
        clickElements.forEach(el => {
            if (el.__lv_click_attached) return;
            el.__lv_click_attached = true;

            const event = el.getAttribute('lv-click');
            el.addEventListener('click', (e) => {
                e.preventDefault();
                const payload = this.getPayloadFromElement(el);
                this.pushEvent(event, payload);
            });
        });

        // Handle lv-change events with debouncing
        const changeElements = this.container.querySelectorAll('[lv-change]');
        changeElements.forEach(el => {
            if (el.__lv_change_attached) return;
            el.__lv_change_attached = true;

            const event = el.getAttribute('lv-change');
            // Get debounce time from lv-debounce attribute (default: 300ms)
            const debounceMs = parseInt(el.getAttribute('lv-debounce') || '300');

            el.addEventListener('input', (e) => {
                // Clear existing timer for this element
                const timerId = this.debounceTimers.get(el);
                if (timerId) {
                    clearTimeout(timerId);
                }

                // Set new timer
                const newTimerId = setTimeout(() => {
                    const payload = this.getPayloadFromElement(el);
                    payload.value = el.value;
                    this.pushEvent(event, payload);
                    this.debounceTimers.delete(el);

                    // Clear pending after a short delay to allow server to catch up
                    // This gives the server time to process and respond
                    // If user keeps typing, it will be marked pending again
                    setTimeout(() => {
                        // Only clear if input is still focused but user hasn't typed more
                        if (this.focusedInput !== el) {
                            this.pendingInputs.delete(el);
                        }
                    }, 100);
                }, debounceMs);

                this.debounceTimers.set(el, newTimerId);
            });
        });

        // Handle lv-submit events
        const formElements = this.container.querySelectorAll('[lv-submit]');
        formElements.forEach(el => {
            if (el.__lv_submit_attached) return;
            el.__lv_submit_attached = true;

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

    applyDiff(diff) {
        // Apply Phoenix LiveView-style diff patches
        // Format: { "0": { "children": { "1": { "s": ["<span>New</span>"] } } } }
        const rootNode = this.container.firstElementChild;
        if (!rootNode) {
            return;
        }

        // The diff format has the root node changes under "0"
        // So we need to apply diff["0"] to the rootNode itself
        if (diff["0"]) {
            this.applyNodeChanges(this.container, rootNode, 0, diff["0"]);
        } else {
            // Otherwise apply diff to root's children
            this.applyDiffToNode(rootNode, diff);
        }

        // Re-attach event listeners after patching
        this.attachEventListeners();
    }

    applyDiffToNode(node, diff) {
        if (!node || !diff) return;

        for (const [key, value] of Object.entries(diff)) {
            // Check if this is a numeric index (child node)
            if (/^\d+$/.test(key)) {
                const index = parseInt(key);
                const child = this.getChildByIndex(node, index);

                if (!child) {
                    continue;
                }

                // Apply changes to this child
                this.applyNodeChanges(node, child, index, value);
            }
        }
    }

    getChildByIndex(node, index) {
        // Get child by index, skipping text nodes with only whitespace
        let currentIndex = 0;
        for (let child = node.firstChild; child; child = child.nextSibling) {
            if (currentIndex === index) {
                return child;
            }
            currentIndex++;
        }
        return null;
    }

    applyNodeChanges(parent, node, index, changes) {
        if (!node) {
            return;
        }

        // Handle static content replacement: "s": ["<html>"] or "s": ["text"]
        if (changes.s && Array.isArray(changes.s)) {
            const content = changes.s.join('');

            // Check if it's a text node update
            if (node.nodeType === Node.TEXT_NODE) {
                // Don't update text nodes inside focused inputs
                if (parent === this.focusedInput || (parent && parent.contains && parent.contains(this.focusedInput))) {
                    return;
                }
                node.nodeValue = content;
                return;
            }

            // Special handling for INPUT/TEXTAREA/SELECT elements
            // Use morphdom instead of replacement to preserve input state
            if (node.tagName === 'INPUT' || node.tagName === 'TEXTAREA' || node.tagName === 'SELECT') {
                const temp = document.createElement('div');
                temp.innerHTML = content;
                const newNode = temp.firstElementChild;

                if (newNode && newNode.tagName === node.tagName) {
                    // Use morphdom to preserve focus and cursor
                    this.morphdom(node, newNode);
                    return;
                }
            }

            // Check if this node contains a focused input
            // If so, use morphdom instead of replacement to preserve input state
            if (this.focusedInput && node.contains && node.contains(this.focusedInput)) {
                const temp = document.createElement('div');
                temp.innerHTML = content;
                const newNode = temp.firstElementChild;

                if (newNode) {
                    // Use morphdom to preserve descendant input states
                    this.morphdom(node, newNode);
                    return;
                }
            }

            // Otherwise it's HTML content - do full replacement
            const temp = document.createElement('div');
            temp.innerHTML = content;

            // Replace all children if multiple nodes
            const fragment = document.createDocumentFragment();
            while (temp.firstChild) {
                fragment.appendChild(temp.firstChild);
            }

            // If fragment has exactly one child, replace the node
            if (fragment.childNodes.length === 1) {
                parent.replaceChild(fragment.firstChild, node);
            } else if (fragment.childNodes.length > 1) {
                // Multiple nodes - replace with all of them
                parent.insertBefore(fragment, node);
                parent.removeChild(node);
            } else if (fragment.childNodes.length === 0) {
                // Empty content - might be text
                const textNode = document.createTextNode(content);
                parent.replaceChild(textNode, node);
            }
        }
        // Handle dynamic content replacement: "d": [["id", "content"]]
        else if (changes.d && Array.isArray(changes.d)) {
            changes.d.forEach(([id, content]) => {
                // Handle dynamic content (tracked by ID)
                const targetNode = document.getElementById(id);
                if (targetNode) {
                    targetNode.innerHTML = content;
                }
            });
        }
        // Handle children updates: "children": { ... }
        else if (changes.children) {
            this.applyDiffToNode(node, changes.children);
        }
        // Handle attribute updates: "attr": { "class": "new-class" }
        else if (changes.attr) {
            for (const [attrName, attrValue] of Object.entries(changes.attr)) {
                if (attrValue === null) {
                    node.removeAttribute(attrName);
                } else {
                    node.setAttribute(attrName, attrValue);
                }
            }
        }
        // Handle text content update: "text": "new text"
        else if (changes.text !== undefined) {
            node.textContent = changes.text;
        }
    }

    patch(html) {
        // Create a temporary container to parse the new HTML
        const temp = document.createElement('div');
        temp.innerHTML = html;
        const newContent = temp.firstElementChild;

        if (!newContent) {
            return;
        }

        // Use morphdom-like algorithm to efficiently patch the DOM
        this.morphdom(this.container.firstElementChild || this.container, newContent);

        // Re-attach event listeners after patching
        this.attachEventListeners();
    }

    morphdom(fromNode, toNode) {
        // Simple morphdom implementation
        // Preserves input values and focus state

        if (!fromNode || !toNode) {
            if (toNode) {
                this.container.appendChild(toNode);
            }
            return;
        }

        // If nodes are different types, replace entirely
        if (fromNode.nodeName !== toNode.nodeName) {
            fromNode.parentNode.replaceChild(toNode.cloneNode(true), fromNode);
            return;
        }

        // Update attributes
        this.updateAttributes(fromNode, toNode);

        // Preserve form input values and cursor position (Phoenix LiveView style)
        if (fromNode.tagName === 'INPUT' || fromNode.tagName === 'TEXTAREA' || fromNode.tagName === 'SELECT') {
            // Update attributes first (they don't interfere with typing)
            this.updateAttributes(fromNode, toNode);

            // Handle value updates carefully
            const isFocused = fromNode === this.focusedInput;
            const hasPendingChanges = this.pendingInputs.has(fromNode);

            if (isFocused && hasPendingChanges) {
                // User is actively typing - preserve their input completely
                // Don't update value at all (prevents race condition)
                // Example: User types "test@example.com" but server only has "test@"
                // We keep "test@example.com" locally until user blurs

                // Restore cursor position if it was somehow lost
                setTimeout(() => {
                    if (fromNode === this.focusedInput) {
                        this.restoreInputState(fromNode);
                    }
                }, 0);
            } else if (isFocused && !hasPendingChanges) {
                // Focused but no pending changes - update but preserve cursor
                // User is focused but hasn't typed anything new
                const cursorStart = fromNode.selectionStart;
                const cursorEnd = fromNode.selectionEnd;

                if (fromNode.type === 'checkbox' || fromNode.type === 'radio') {
                    fromNode.checked = toNode.checked;
                } else {
                    fromNode.value = toNode.value || '';
                }

                // Restore cursor
                if (cursorStart !== null) {
                    setTimeout(() => {
                        try {
                            fromNode.setSelectionRange(cursorStart, cursorEnd);
                        } catch (e) {
                            // Ignore
                        }
                    }, 0);
                }
            } else {
                // Not focused - safe to update from server
                if (fromNode.type === 'checkbox' || fromNode.type === 'radio') {
                    fromNode.checked = toNode.checked;
                } else {
                    fromNode.value = toNode.value || '';
                }
                // Clear pending state since server value is now applied
                this.pendingInputs.delete(fromNode);
            }

            // Skip the default attribute update since we already did it above
            return;
        }

        // Update text nodes
        if (fromNode.nodeType === Node.TEXT_NODE) {
            if (fromNode.nodeValue !== toNode.nodeValue) {
                fromNode.nodeValue = toNode.nodeValue;
            }
            return;
        }

        // Morph children
        const fromChildren = Array.from(fromNode.childNodes);
        const toChildren = Array.from(toNode.childNodes);

        // Simple algorithm: match by position
        const maxLength = Math.max(fromChildren.length, toChildren.length);

        for (let i = 0; i < maxLength; i++) {
            const fromChild = fromChildren[i];
            const toChild = toChildren[i];

            if (!toChild) {
                // Remove extra nodes
                if (fromChild) {
                    fromNode.removeChild(fromChild);
                }
            } else if (!fromChild) {
                // Add new nodes
                fromNode.appendChild(toChild.cloneNode(true));
            } else if (fromChild.nodeType === Node.TEXT_NODE && toChild.nodeType === Node.TEXT_NODE) {
                // Update text content
                if (fromChild.nodeValue !== toChild.nodeValue) {
                    fromChild.nodeValue = toChild.nodeValue;
                }
            } else if (fromChild.nodeType === Node.ELEMENT_NODE && toChild.nodeType === Node.ELEMENT_NODE) {
                // Recursively morph element nodes
                this.morphdom(fromChild, toChild);
            } else {
                // Different node types, replace
                fromNode.replaceChild(toChild.cloneNode(true), fromChild);
            }
        }
    }

    updateAttributes(fromNode, toNode) {
        // Remove old attributes
        const fromAttrs = Array.from(fromNode.attributes || []);
        fromAttrs.forEach(attr => {
            if (!toNode.hasAttribute(attr.name)) {
                fromNode.removeAttribute(attr.name);
            }
        });

        // Add/update new attributes
        const toAttrs = Array.from(toNode.attributes || []);
        toAttrs.forEach(attr => {
            if (fromNode.getAttribute(attr.name) !== attr.value) {
                fromNode.setAttribute(attr.name, attr.value);
            }
        });
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
        // Expose globally for custom form handlers
        window.liveSocket = liveview;
    }
});
