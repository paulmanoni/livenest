package main

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/paulmanoni/livenest/liveview"
)

// FormComponent demonstrates form handling with validation
type FormComponent struct {
	liveview.TemplateComponent
}

// FormData holds the form values
type FormData struct {
	Name        string
	Email       string
	Age         string
	Message     string
	AcceptTerms bool
}

// ValidationErrors holds validation error messages
type ValidationErrors map[string]string

// Mount initializes the form component
func (f *FormComponent) Mount(socket *liveview.Socket) error {
	socket.Assign(map[string]interface{}{
		"formData":  FormData{},
		"errors":    ValidationErrors{},
		"submitted": false,
	})
	return nil
}

// HandleChange handles input changes with live validation
func (f *FormComponent) HandleChange(socket *liveview.Socket, payload map[string]interface{}) error {
	field, _ := payload["field"].(string)
	value, _ := payload["value"].(string)

	formData := socket.Assigns["formData"].(FormData)
	errors := socket.Assigns["errors"].(ValidationErrors)

	// Update form data
	switch field {
	case "name":
		formData.Name = value
		delete(errors, "name")
		if value == "" {
			errors["name"] = "Name is required"
		} else if len(value) < 2 {
			errors["name"] = "Name must be at least 2 characters"
		}
	case "email":
		formData.Email = value
		delete(errors, "email")
		if value == "" {
			errors["email"] = "Email is required"
		} else if !isValidEmail(value) {
			errors["email"] = "Invalid email format"
		}
	case "age":
		formData.Age = value
		delete(errors, "age")
		if value == "" {
			errors["age"] = "Age is required"
		} else if !isNumeric(value) {
			errors["age"] = "Age must be a number"
		}
	case "message":
		formData.Message = value
		delete(errors, "message")
		if value == "" {
			errors["message"] = "Message is required"
		} else if len(value) < 10 {
			errors["message"] = "Message must be at least 10 characters"
		}
	case "terms":
		formData.AcceptTerms = value == "true"
		delete(errors, "terms")
	}

	socket.Assign(map[string]interface{}{
		"formData": formData,
		"errors":   errors,
	})

	return nil
}

// HandleSubmit handles form submission
func (f *FormComponent) HandleSubmit(socket *liveview.Socket, payload map[string]interface{}) error {
	formData := socket.Assigns["formData"].(FormData)
	errors := f.validateForm(formData)

	if len(errors) > 0 {
		socket.Assign(map[string]interface{}{
			"errors": errors,
		})
		socket.PutFlash("error", "Please fix the errors below")
		return nil
	}

	// Form is valid - process it
	socket.Assign(map[string]interface{}{
		"submitted": true,
		"errors":    ValidationErrors{},
	})

	socket.PutFlash("success", fmt.Sprintf("Form submitted successfully! Welcome, %s", formData.Name))
	return nil
}

// HandleReset resets the form
func (f *FormComponent) HandleReset(socket *liveview.Socket, payload map[string]interface{}) error {
	socket.Assign(map[string]interface{}{
		"formData":  FormData{},
		"errors":    ValidationErrors{},
		"submitted": false,
	})
	socket.PutFlash("info", "Form reset")
	return nil
}

// validateForm validates all form fields
func (f *FormComponent) validateForm(data FormData) ValidationErrors {
	errors := ValidationErrors{}

	if data.Name == "" {
		errors["name"] = "Name is required"
	} else if len(data.Name) < 2 {
		errors["name"] = "Name must be at least 2 characters"
	}

	if data.Email == "" {
		errors["email"] = "Email is required"
	} else if !isValidEmail(data.Email) {
		errors["email"] = "Invalid email format"
	}

	if data.Age == "" {
		errors["age"] = "Age is required"
	} else if !isNumeric(data.Age) {
		errors["age"] = "Age must be a number"
	}

	if data.Message == "" {
		errors["message"] = "Message is required"
	} else if len(data.Message) < 10 {
		errors["message"] = "Message must be at least 10 characters"
	}

	if !data.AcceptTerms {
		errors["terms"] = "You must accept the terms and conditions"
	}

	return errors
}

// Render returns the HTML for the form
func (f *FormComponent) Render(socket *liveview.Socket) (template.HTML, error) {
	return f.TemplateComponent.Render("form.html", socket.Assigns)
}

// Helper functions
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
