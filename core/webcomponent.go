package core

import (
	"livenest/liveview"
)

// RegisterWebComponent registers a custom web component with validation
func (a *App) RegisterWebComponent(config liveview.WebComponentConfig) {
	if a.webComponents == nil {
		a.webComponents = make(map[string]liveview.WebComponentConfig)
	}
	a.webComponents[config.TagName] = config
}

// GetWebComponentsJS returns the JavaScript for all registered web components
func (a *App) GetWebComponentsJS() string {
	if a.webComponents == nil || len(a.webComponents) == 0 {
		return ""
	}
	return liveview.BuildWebComponentJS(a.webComponents)
}
