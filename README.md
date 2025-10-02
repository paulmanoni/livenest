<p align="center">
  <img src="./livenest_logo.png" alt="LiveNest Logo" width="400"/>
</p>

<h1 align="center">LiveNest 🪺</h1>
<p align="center">
  <b>Django-like Web Framework for Go</b><br/>
  <i>with Phoenix LiveView-inspired real-time capabilities</i>
</p>

<p align="center">
  <a href="https://github.com/paulmanoni/livenest/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/paulmanoni/livenest/ci.yml?branch=main" alt="Build Status"/>
  </a>
  <a href="https://pkg.go.dev/github.com/paulmanoni/livenest">
    <img src="https://pkg.go.dev/badge/github.com/paulmanoni/livenest" alt="Go Reference"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/paulmanoni/livenest">
    <img src="https://goreportcard.com/badge/github.com/paulmanoni/livenest" alt="Go Report Card"/>
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"/>
  </a>
</p>

A Django-like web framework for Go with Phoenix LiveView-inspired real-time capabilities.

Build real-time, interactive web applications with the simplicity of server-side rendering and the power of WebSockets.

## Features

- **Gin-based Routing**: Fast HTTP routing powered by Gin with fluent API
- **GORM ORM**: Database access with GORM, supporting SQLite, PostgreSQL, and MySQL
- **LiveView**: Real-time, interactive components using WebSockets
- **Template Engine**: HTML template rendering with custom functions and file-based templates
- **Django-like QuerySets**: Familiar API for database queries
- **Configuration**: Support for JSON and TOML configuration files
- **Flash Messages**: Built-in user notification system with success, error, info, and warning types
- **Event Routing**: Automatic routing of events to `Handle*` methods using reflection
- **Form Validation**: Client-side and server-side validation support
- **Component System**: Reusable components with `<lv-component>` web component tag
- **Template Components**: File-based templates with `TemplateComponent` base class

## Project Structure

```
livenest/
├── core/           # Core application and context
├── orm/            # ORM manager and querysets
├── liveview/       # LiveView components and WebSocket handling
├── template/       # Template engine and functions
├── admin/          # Admin interface (coming soon)
└── examples/       # Example applications
```

## Quick Start

### Installation

```bash
go get livenest
```

### Basic Usage

```go
package main

import (
    "livenest/core"
    "gorm.io/driver/sqlite"
)

func main() {
    // Create app
    app := core.New(nil)

    // Connect database
    app.ConnectDB(sqlite.Open("app.db"))

    // Define routes
    app.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello LiveNest!"})
    })

    // Run server
    app.Run(":8080")
}
```

## LiveView Example

Create an interactive counter component with automatic event routing:

```go
type CounterComponent struct {
    liveview.TemplateComponent
}

func (c *CounterComponent) Mount(socket *liveview.Socket) error {
    socket.Assign(map[string]interface{}{
        "count": 0,
    })
    return nil
}

// Events are automatically routed to Handle* methods
func (c *CounterComponent) HandleIncrement(socket *liveview.Socket, payload map[string]interface{}) error {
    count := socket.Assigns["count"].(int)
    socket.Assign(map[string]interface{}{
        "count": count + 1,
    })
    return nil
}

func (c *CounterComponent) HandleDecrement(socket *liveview.Socket, payload map[string]interface{}) error {
    count := socket.Assigns["count"].(int)
    socket.Assign(map[string]interface{}{
        "count": count - 1,
    })
    return nil
}

func (c *CounterComponent) Render(socket *liveview.Socket) (template.HTML, error) {
    count, _ := socket.Get("count")
    html := fmt.Sprintf(`
        <div>
            <h2>Count: %d</h2>
            <button lv-click="decrement">-</button>
            <button lv-click="increment">+</button>
        </div>
    `, count)
    return template.HTML(html), nil
}
```

### Using Template Files

Components can also render from HTML template files:

```go
type TodoListComponent struct {
    liveview.TemplateComponent
}

func (t *TodoListComponent) Mount(socket *liveview.Socket) error {
    t.TemplateDir = "examples/templates"
    socket.Assign(map[string]interface{}{
        "todos": []TodoItem{},
    })
    return nil
}

func (t *TodoListComponent) Render(socket *liveview.Socket) (template.HTML, error) {
    return t.TemplateComponent.Render("todo.html", socket.Assigns)
}
```

## Running Examples

```bash
cd examples
go run main.go
```

The example server includes multiple LiveView demonstrations:

- **http://localhost:8080** - Counter (inline template)
- **http://localhost:8080/counter** - Counter (fluent API)
- **http://localhost:8080/counter-template** - Counter (file template)
- **http://localhost:8080/dashboard** - Dashboard (subdirectory template)
- **http://localhost:8080/todo** - Todo List (CRUD operations)
- **http://localhost:8080/form** - Contact Form (with validation)
- **http://localhost:8080/chat** - Real-time Chat
- **http://localhost:8080/component-tag** - `<lv-component>` web component examples

## Core Components

### App

The main application wrapper that combines Gin and GORM:

```go
app := core.New(&core.Config{
    Debug: true,
    TemplateDir: "templates",
    StaticDir: "static",
})
```

### ORM Manager

Django-like queryset API:

```go
qs := orm.NewQuerySet(db)
users, err := qs.Filter("age > ?", 18).OrderBy("name").All(&users)
```

### LiveView

Real-time components with WebSocket communication:

- `Component` interface for defining interactive components
- `Socket` for managing state and assigns
- Automatic DOM diffing and updates
- Flash messages for user notifications (`socket.PutFlash("success", "Message")`)
- Event attributes: `lv-click`, `lv-change`, `lv-submit`
- Automatic event routing to `Handle*` methods

### Template Engine

```go
engine := template.NewEngine("templates")
engine.Load()
html, err := engine.Render("index.html", data)
```

## Configuration

Create a `config.json`:

```json
{
  "debug": true,
  "template_dir": "templates",
  "static_dir": "static",
  "database": {
    "driver": "sqlite",
    "database": "app.db"
  },
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  }
}
```

Load configuration:

```go
config, err := core.LoadConfig("config.json")
app := core.New(config)
```

## Roadmap

- [ ] Admin interface (Django-like)
- [ ] Form handling and validation
- [ ] Authentication and authorization
- [ ] Migrations system
- [ ] CLI tool for scaffolding
- [ ] Middleware library
- [ ] Session management
- [ ] CSRF protection
- [ ] File uploads
- [ ] Testing utilities

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Inspired By

- Django (Python web framework)
- Phoenix LiveView (Elixir real-time framework)
- Gin (Go HTTP framework)
- GORM (Go ORM)
