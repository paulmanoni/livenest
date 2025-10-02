package main

import (
	"log"

	"livenest/core"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
)

func main() {
	// Create app with default config
	app := core.New(nil)

	// Connect to database
	if err := app.ConnectDB(sqlite.Open("example.db")); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// LiveView route using fluent API
	// Component is automatically registered as "index" for <component name="index">
	app.NewHandler().
		Path("/").
		AsLive().
		AddComponent(&CounterComponent{}).
		Build()

	// Register same component with custom name "counter" for <component name="counter">
	app.NewHandler().
		Path("/counter").
		AsLive().
		AddComponent(&CounterComponent{}).WithName("counter").
		AddComponent(&CounterComponent{}).WithName("counter2").
		Build()

	// Register counter using template file
	app.NewHandler().
		Path("/counter-template").
		AsLive().
		AddComponent(&CounterComponent{UseTemplate: true}).WithName("counter-template").
		Build()

	// Register dashboard component with subdirectory template
	app.NewHandler().
		Path("/dashboard").
		AsLive().
		AddComponent(&DashboardComponent{}).WithName("dashboard").
		Build()

	// Register todo list component
	app.NewHandler().
		Path("/todo").
		AsLive().
		AddComponent(&TodoListComponent{}).WithName("todo").
		Build()

	// Register form component with validation
	app.NewHandler().
		Path("/form").
		AsLive().
		AddComponent(&FormComponent{}).WithName("contact-form").
		Build()

	// Register chat component
	app.NewHandler().
		Path("/chat").
		AsLive().
		AddComponent(&ChatComponent{}).WithName("chat").
		Build()

	// Component tag example page
	app.NewHandler().
		Path("/component-tag").
		AsGet().
		Func(func(c *gin.Context) {
			c.Header("Content-Type", "text/html; charset=utf-8")
			html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LiveNest Component Tag Example</title>
    <script src="/livenest/liveview.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        component {
            display: block;
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
        }
    </style>
</head>
<body>
    <h1>LiveNest Component Tag Examples</h1>

    <div class="container">
        <component name="counter" id="counter-1"></component>
        <component name="counter" id="counter-2"></component>
        <component name="counter" id="counter-3"></component>
    </div>

    <script>
        document.querySelectorAll('component').forEach(comp => {
            comp.addEventListener('component-loaded', (e) => {
                console.log('Component loaded:', e.detail);
            });
        });
    </script>
</body>
</html>`
			c.Data(200, "text/html; charset=utf-8", []byte(html))
		}).
		Build()

	// API route example
	app.NewHandler().
		Path("/api/hello").
		AsGet().
		Func(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Hello from LiveNest!"})
		}).
		Build()

	// Serve static files
	app.Router.Static("/static", "./static")

	// Start server
	log.Println("Starting LiveNest example server...")
	log.Println("Routes:")
	log.Println("  http://localhost:8080                  - Counter (inline template)")
	log.Println("  http://localhost:8080/counter          - Counter (fluent API)")
	log.Println("  http://localhost:8080/counter-template - Counter (file template)")
	log.Println("  http://localhost:8080/dashboard        - Dashboard (subdirectory template)")
	log.Println("  http://localhost:8080/todo             - Todo List (CRUD operations)")
	log.Println("  http://localhost:8080/form             - Contact Form (with validation)")
	log.Println("  http://localhost:8080/chat             - Real-time Chat")
	log.Println("  http://localhost:8080/component-tag    - <component> tag examples")
	if err := app.Run(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
