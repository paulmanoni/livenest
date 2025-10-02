package main

import (
	"log"

	"livenest/core"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
)

func mainFluent() {
	// Create app with default config
	app := core.New(nil)

	// Connect to database
	if err := app.ConnectDB(sqlite.Open("example.db")); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// LiveView route - fluent API
	app.NewHandler().
		Path("/").
		AsLive().
		AddComponent(&CounterComponent{}).
		Build()

	// Regular POST route - fluent API
	app.NewHandler().
		Path("/api/data").
		AsPost().
		Func(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Data received"})
		}).
		Build()

	// Regular GET route - fluent API
	app.NewHandler().
		Path("/api/hello").
		AsGet().
		Func(func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Hello from LiveNest!"})
		}).
		Build()

	// Start server
	log.Println("Starting LiveNest server with fluent API...")
	if err := app.Run(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
