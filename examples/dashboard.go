package main

import (
	"html/template"
	"math/rand"

	"livenest/liveview"
)

// DashboardComponent demonstrates template file usage with subdirectories
type DashboardComponent struct {
	liveview.TemplateComponent
}

// Mount initializes the dashboard
func (d *DashboardComponent) Mount(socket *liveview.Socket) error {
	socket.Assign(map[string]interface{}{
		"user_name":       "John Doe",
		"total_users":     1234,
		"active_sessions": 89,
		"revenue":         45678.90,
	})
	return nil
}

// HandleRefresh refreshes the dashboard data
func (d *DashboardComponent) HandleRefresh(socket *liveview.Socket, payload map[string]interface{}) error {
	// Simulate data refresh
	socket.Assign(map[string]interface{}{
		"total_users":     rand.Intn(2000) + 1000,
		"active_sessions": rand.Intn(200) + 50,
		"revenue":         float64(rand.Intn(100000)) + 10000.50,
	})
	return nil
}

// HandleExport handles export action
func (d *DashboardComponent) HandleExport(socket *liveview.Socket, payload map[string]interface{}) error {
	// In a real app, this would trigger a file download
	socket.PutFlash("info", "Report exported successfully!")
	return nil
}

// Render uses a template from pages subdirectory
func (d *DashboardComponent) Render(socket *liveview.Socket) (template.HTML, error) {
	return d.TemplateComponent.Render("pages/dashboard.html", socket.Assigns)
}
