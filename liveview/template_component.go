package liveview

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// TemplateComponent is a base component that loads templates from files
type TemplateComponent struct {
	TemplateDir  string
	TemplateName string
	templateContent string
}

// LoadTemplate loads the template from a file
func (t *TemplateComponent) LoadTemplate() error {
	if t.templateContent != "" {
		return nil // Already loaded
	}

	templatePath := filepath.Join(t.TemplateDir, t.TemplateName)

	// Try with .html extension if not present
	if !strings.HasSuffix(templatePath, ".html") {
		templatePath += ".html"
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	t.templateContent = string(content)
	return nil
}

// RenderTemplate renders the template with the given data
func (t *TemplateComponent) RenderTemplate(data interface{}) (template.HTML, error) {
	if err := t.LoadTemplate(); err != nil {
		return "", err
	}

	// Parse and execute template
	tmpl, err := template.New(t.TemplateName).Parse(t.templateContent)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}

// SetTemplateContent sets the template content directly (useful for testing or inline templates)
func (t *TemplateComponent) SetTemplateContent(content string) {
	t.templateContent = content
}

// Render loads and renders a template file with the given data
// Usage: return c.Render("counter.html", socket.Assigns)
// or:    return c.Render("pages/dashboard.html", socket.Assigns)
func (t *TemplateComponent) Render(templatePath string, data interface{}) (template.HTML, error) {
	// Set template path
	t.TemplateName = templatePath
	if t.TemplateDir == "" {
		t.TemplateDir = "templates" // default directory
	}

	// Force reload for this render
	t.templateContent = ""

	// Load and render
	return t.RenderTemplate(data)
}
