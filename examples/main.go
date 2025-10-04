package main

import (
	"log"

	"livenest/core"

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

	// Register new form examples with auto-generation
	app.NewHandler().
		Path("/registration").
		AsLive().
		AddComponent(NewUserForm()).WithName("user-registration").
		Build()

	app.NewHandler().
		Path("/contact").
		AsLive().
		AddComponent(NewContactForm()).WithName("contact").
		Build()

	app.NewHandler().
		Path("/review").
		AsLive().
		AddComponent(NewProductReview()).WithName("product-review").
		Build()

	app.NewHandler().
		Path("/login").
		AsLive().
		AddComponent(NewLoginForm()).WithName("login").
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
	log.Println("  http://localhost:8080/form             - Contact Form (old implementation)")
	log.Println("  http://localhost:8080/chat             - Real-time Chat")
	log.Println("  http://localhost:8080/registration     - User Registration (auto-generated)")
	log.Println("  http://localhost:8080/contact          - Contact Form (auto-generated)")
	log.Println("  http://localhost:8080/review           - Product Review (auto-generated)")
	log.Println("  http://localhost:8080/login            - Login Form (auto-generated)")
	log.Println("  http://localhost:8080/component-tag    - <component> tag examples")
	if err := app.Run(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
