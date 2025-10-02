package main

import (
	"fmt"
	"html/template"

	"livenest/liveview"
)

// CounterComponent is a simple counter LiveView component
type CounterComponent struct {
	liveview.TemplateComponent
	UseTemplate bool // Set to true to use counter.html template
}

// Mount initializes the counter component
func (c *CounterComponent) Mount(socket *liveview.Socket) error {
	socket.Assign(map[string]interface{}{
		"count": 0,
	})
	return nil
}

// HandleIncrement handles the increment event
func (c *CounterComponent) HandleIncrement(socket *liveview.Socket, payload map[string]interface{}) error {
	count := socket.Assigns["count"].(int)
	socket.Assign(map[string]interface{}{
		"count": count + 1,
	})
	return nil
}

// HandleDecrement handles the decrement event
func (c *CounterComponent) HandleDecrement(socket *liveview.Socket, payload map[string]interface{}) error {
	count := socket.Assigns["count"].(int)
	socket.Assign(map[string]interface{}{
		"count": count - 1,
	})
	return nil
}

// HandleReset handles the reset event
func (c *CounterComponent) HandleReset(socket *liveview.Socket, payload map[string]interface{}) error {
	socket.Assign(map[string]interface{}{
		"count": 0,
	})
	return nil
}

// Render returns the HTML for the counter component
func (c *CounterComponent) Render(socket *liveview.Socket) (template.HTML, error) {
	count, _ := socket.Get("count")

	// Use template file if configured
	if c.UseTemplate {
		return c.TemplateComponent.Render("counter.html", socket.Assigns)
	}

	// Otherwise use inline template
	html := fmt.Sprintf(`
		<div class="counter">
			<h1>LiveView Counter</h1>
			<div class="count-display">
				<h2>Count: %d</h2>
			</div>
			<div class="buttons">
				<button lv-click="decrement">-</button>
				<button lv-click="reset">Reset</button>
				<button lv-click="increment">+</button>
			</div>
		</div>
		<style>
			.counter {
				font-family: Arial, sans-serif;
				max-width: 400px;
				margin: 50px auto;
				text-align: center;
				padding: 20px;
				border: 2px solid #333;
				border-radius: 10px;
			}
			.count-display {
				margin: 30px 0;
				font-size: 2em;
				color: #2c3e50;
			}
			.buttons {
				display: flex;
				justify-content: center;
				gap: 10px;
			}
			button {
				padding: 10px 20px;
				font-size: 1.2em;
				border: none;
				border-radius: 5px;
				cursor: pointer;
				background-color: #3498db;
				color: white;
				transition: background-color 0.3s;
			}
			button:hover {
				background-color: #2980b9;
			}
		</style>
	`, count)

	return template.HTML(html), nil
}
