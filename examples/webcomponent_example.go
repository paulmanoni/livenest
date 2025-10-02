package main

import (
	"livenest/core"
	"livenest/liveview"
)

func setupWebComponents(app *core.App) {
	// Register a user-card component with validation
	minAge := 18
	maxAge := 120

	app.RegisterWebComponent(liveview.WebComponentConfig{
		TagName: "user-card",
		Attributes: map[string]liveview.AttributeConfig{
			"name": {
				Required: true,
				Type:     "string",
			},
			"email": {
				Required: true,
				Type:     "email",
			},
			"age": {
				Required: false,
				Type:     "number",
				Min:      &minAge,
				Max:      &maxAge,
			},
			"website": {
				Required: false,
				Type:     "url",
			},
		},
	})

	// Register a counter-widget component
	minCount := 0
	maxCount := 100

	app.RegisterWebComponent(liveview.WebComponentConfig{
		TagName: "counter-widget",
		Attributes: map[string]liveview.AttributeConfig{
			"initial": {
				Required: false,
				Type:     "number",
				Min:      &minCount,
				Max:      &maxCount,
				Default:  "0",
			},
			"step": {
				Required: false,
				Type:     "number",
				Default:  "1",
			},
		},
	})

	// Register a validated form input
	app.RegisterWebComponent(liveview.WebComponentConfig{
		TagName: "validated-input",
		Attributes: map[string]liveview.AttributeConfig{
			"type": {
				Required: true,
				Type:     "string",
			},
			"pattern": {
				Required: false,
				Type:     "string",
			},
			"required": {
				Required: false,
				Type:     "boolean",
			},
		},
	})
}

// Usage in HTML:
/*
<!DOCTYPE html>
<html>
<head>
    <script src="/livenest/components.js"></script>
</head>
<body>
    <!-- Valid usage -->
    <user-card
        name="John Doe"
        email="john@example.com"
        age="25"
        website="https://example.com">
    </user-card>

    <!-- Invalid - will show validation error -->
    <user-card email="not-an-email"></user-card>

    <!-- Counter widget -->
    <counter-widget initial="10" step="5"></counter-widget>

    <!-- Validated input -->
    <validated-input
        type="email"
        pattern="^[a-zA-Z0-9]+@example\\.com$"
        required="true">
    </validated-input>
</body>
</html>
*/
