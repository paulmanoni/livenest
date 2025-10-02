package template

import (
	"bytes"
	"html/template"
	"io"
	"os"
	"path/filepath"
)

// Engine wraps Go's html/template with additional functionality
type Engine struct {
	templates *template.Template
	dir       string
	funcs     template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine(dir string) *Engine {
	return &Engine{
		dir:   dir,
		funcs: DefaultFuncs(),
	}
}

// AddFunc adds a template function
func (e *Engine) AddFunc(name string, fn interface{}) {
	e.funcs[name] = fn
}

// AddFuncs adds multiple template functions
func (e *Engine) AddFuncs(funcs template.FuncMap) {
	for name, fn := range funcs {
		e.funcs[name] = fn
	}
}

// Load loads all templates from the template directory
func (e *Engine) Load() error {
	if _, err := os.Stat(e.dir); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(e.dir, 0755); err != nil {
			return err
		}
		// No templates to load yet
		e.templates = template.New("").Funcs(e.funcs)
		return nil
	}

	tmpl := template.New("").Funcs(e.funcs)

	err := filepath.Walk(e.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only parse .html and .tmpl files
		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".tmpl" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Get relative path for template name
		relPath, err := filepath.Rel(e.dir, path)
		if err != nil {
			return err
		}

		_, err = tmpl.New(relPath).Parse(string(data))
		return err
	})

	if err != nil {
		return err
	}

	e.templates = tmpl
	return nil
}

// Render renders a template with the given data
func (e *Engine) Render(name string, data interface{}) (template.HTML, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// RenderTo renders a template to a writer
func (e *Engine) RenderTo(w io.Writer, name string, data interface{}) error {
	return e.templates.ExecuteTemplate(w, name, data)
}

// Parse parses a template string
func (e *Engine) Parse(name, tmpl string) error {
	_, err := e.templates.New(name).Parse(tmpl)
	return err
}

// Exists checks if a template exists
func (e *Engine) Exists(name string) bool {
	return e.templates.Lookup(name) != nil
}
