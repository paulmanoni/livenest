package liveview

import (
	"fmt"
	"strings"
)

// GetComponentTagJS returns the JavaScript for the universal <component> tag
func GetComponentTagJS() string {
	return `
// Universal <component> Web Component for LiveNest
class LiveNestComponent extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.liveview = null;
    }

    async connectedCallback() {
        const componentName = this.getAttribute('name');
        const componentId = this.getAttribute('id') || this.generateId();

        if (!componentName) {
            this.shadowRoot.innerHTML = '<div style="color: red;">Error: component name is required</div>';
            return;
        }

        // Set ID if not provided
        if (!this.hasAttribute('id')) {
            this.setAttribute('id', componentId);
        }

        // Fetch initial component HTML from server
        try {
            const response = await fetch('/livenest/component/' + componentName);
            if (!response.ok) {
                throw new Error('Component not found: ' + componentName);
            }

            const data = await response.json();

            // Create LiveView container
            const container = document.createElement('div');
            container.id = 'liveview-' + componentId;
            container.dataset.component = componentName;
            container.dataset.socketId = data.socket_id;
            container.dataset.componentId = data.component_id;
            container.innerHTML = data.html;

            this.shadowRoot.appendChild(container);

            // Initialize LiveView WebSocket connection
            this.liveview = new LiveViewSocket(componentName, data.socket_id);
            this.liveview.container = container;
            this.liveview.connect();

            // Dispatch loaded event
            this.dispatchEvent(new CustomEvent('component-loaded', {
                detail: { componentId: data.component_id, componentName }
            }));

        } catch (error) {
            console.error('Failed to load component:', error);
            this.shadowRoot.innerHTML = '<div style="color: red;">Error loading component: ' + error.message + '</div>';
        }
    }

    disconnectedCallback() {
        // Clean up WebSocket connection
        if (this.liveview && this.liveview.ws) {
            this.liveview.ws.close();
        }
    }

    generateId() {
        return 'cmp-' + Math.random().toString(36).substr(2, 9);
    }

    // Get component state
    getState() {
        if (!this.liveview) return null;
        const container = this.shadowRoot.querySelector('[data-component-id]');
        return {
            componentId: container?.dataset.componentId,
            socketId: container?.dataset.socketId,
            componentName: container?.dataset.component
        };
    }

    // Send event to component
    sendEvent(eventName, payload = {}) {
        if (this.liveview) {
            this.liveview.sendEvent(eventName, payload);
        }
    }

    static get observedAttributes() {
        return ['name'];
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'name' && oldValue !== newValue && oldValue !== null) {
            // Reload component if name changes
            this.connectedCallback();
        }
    }
}

// Custom element names must contain a hyphen
customElements.define('lv-component', LiveNestComponent);
`
}

// ComponentTagHandler handles requests for component instances
func (h *Handler) ComponentTagHandler(componentName string) interface{} {
	return func(c interface{}) {
		// This will be implemented to handle /livenest/component/:name requests
		// Returns JSON with: {html, socket_id, component_id}
	}
}

// GetComponentTagHTML generates HTML for server-side rendering
func GetComponentTagHTML(name string, attrs map[string]string) string {
	var attrStr strings.Builder
	attrStr.WriteString(fmt.Sprintf(`name="%s"`, name))

	for key, val := range attrs {
		attrStr.WriteString(fmt.Sprintf(` %s="%s"`, key, val))
	}

	return fmt.Sprintf(`<component %s></component>`, attrStr.String())
}
