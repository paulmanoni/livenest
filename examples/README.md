# LiveNest Examples

This directory contains comprehensive examples demonstrating LiveNest's features.

## Running the Examples

```bash
go run *.go
```

Then visit http://localhost:8080 to explore the examples.

## Available Examples

### 1. Counter (`/`, `/counter`, `/counter-template`)

**What it demonstrates:**
- Basic LiveView component structure
- Event handling with `lv-click`
- State management with `socket.Assign()`
- Inline templates vs file-based templates
- Automatic event routing to `Handle*` methods

**Key concepts:**
```go
func (c *CounterComponent) HandleIncrement(socket *liveview.Socket, payload map[string]interface{}) error {
    count := socket.Assigns["count"].(int)
    socket.Assign(map[string]interface{}{"count": count + 1})
    return nil
}
```

### 2. Dashboard (`/dashboard`)

**What it demonstrates:**
- Template files with subdirectories (`templates/pages/dashboard.html`)
- Multiple data types in assigns
- Flash messages for user notifications
- Professional UI with gradients and cards

**Key concepts:**
```go
func (d *DashboardComponent) Render(socket *liveview.Socket) (template.HTML, error) {
    return d.TemplateComponent.Render("pages/dashboard.html", socket.Assigns)
}
```

### 3. Todo List (`/todo`)

**What it demonstrates:**
- CRUD operations (Create, Read, Update, Delete)
- Complex state management with arrays
- Filtering and data transformation
- Multiple event handlers

**Key features:**
- Add new todos
- Toggle completion status
- Delete todos
- Filter by all/active/completed
- Clear completed items

### 4. Contact Form (`/form`)

**What it demonstrates:**
- Form handling with validation
- Real-time validation feedback
- Multiple input types (text, email, tel, textarea)
- Error state management
- Success/error messages

**Validation examples:**
```go
func validateEmail(email string) string {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return "Invalid email format"
    }
    return ""
}
```

### 5. Real-time Chat (`/chat`)

**What it demonstrates:**
- Shared global state across multiple connections
- Real-time updates with auto-refresh
- User sessions and identification
- Message history management

**Key concepts:**
- Global mutex-protected state
- Auto-refresh every 3 seconds
- Message broadcasting across users
- Timestamps and user attribution

### 6. Component Tag Examples (`/component-tag`)

**What it demonstrates:**
- Using `<component>` tags in static HTML
- Multiple instances of the same component
- Component registration and discovery
- Client-side component initialization

**Usage:**
```html
<component name="counter" id="counter-1"></component>
<component name="counter" id="counter-2"></component>
```

## Event Attributes

LiveNest supports several event attributes:

- **`lv-click="event"`**: Triggered on click
- **`lv-change="event"`**: Triggered on input/change
- **`lv-submit="event"`**: Triggered on form submit
- **`lv-keyup="event"`**: Triggered on keyup
- **`lv-value-*="value"`**: Pass custom values with events

Example:
```html
<button lv-click="delete" lv-value-id="123">Delete</button>
```

This will call `HandleDelete()` with `payload["id"] = "123"`.

## Flash Messages

Display notifications to users:

```go
socket.PutFlash("success", "Item saved successfully!")
socket.PutFlash("error", "Something went wrong")
socket.PutFlash("info", "New messages available")
socket.PutFlash("warning", "Please review your input")
```

Flash messages automatically:
- Display in the top-right corner
- Auto-dismiss after 5 seconds
- Can be manually closed
- Support 4 types: success, error, info, warning

## Component Structure

Every LiveView component follows this pattern:

```go
type MyComponent struct {
    liveview.TemplateComponent  // Embed for template support
}

// Initialize state
func (c *MyComponent) Mount(socket *liveview.Socket) error {
    socket.Assign(map[string]interface{}{
        "key": "value",
    })
    return nil
}

// Handle events (automatically routed)
func (c *MyComponent) HandleMyEvent(socket *liveview.Socket, payload map[string]interface{}) error {
    // Update state
    socket.Assign(map[string]interface{}{
        "updated": true,
    })
    return nil
}

// Render the component
func (c *MyComponent) Render(socket *liveview.Socket) (template.HTML, error) {
    // Option 1: Use template file
    return c.TemplateComponent.Render("mytemplate.html", socket.Assigns)

    // Option 2: Return inline HTML
    return template.HTML("<div>Hello</div>"), nil
}
```

## Fluent API

Register components using the fluent API:

```go
app.NewHandler().
    Path("/myroute").
    AsLive().
    AddComponent(&MyComponent{}).WithName("my-component").
    Build()
```

For regular HTTP routes:

```go
app.NewHandler().
    Path("/api/hello").
    AsGet().
    Func(func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello!"})
    }).
    Build()
```

## Tips and Best Practices

1. **State Management**: Keep component state minimal and focused
2. **Event Naming**: Use descriptive event names that map to clear `Handle*` methods
3. **Validation**: Validate on both client and server side
4. **Flash Messages**: Use appropriate types (success/error/info/warning)
5. **Templates**: Use file-based templates for complex UIs
6. **Global State**: Use mutexes when sharing state across components
7. **Auto-refresh**: Be mindful of refresh intervals to avoid performance issues

## Learn More

- See the main [README](../README.md) for framework documentation
- Check individual component files for detailed implementations
- Explore the LiveView client code in `liveview/static/liveview.js`
