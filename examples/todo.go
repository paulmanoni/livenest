package main

import (
	"html/template"
	"strconv"
	"time"

	"github.com/paulmanoni/livenest/liveview"
)

// TodoItem represents a single todo item
type TodoItem struct {
	ID        int
	Text      string
	Completed bool
	CreatedAt time.Time
}

// TodoListComponent demonstrates CRUD operations with LiveView
type TodoListComponent struct {
	liveview.TemplateComponent
}

// Mount initializes the todo list
func (t *TodoListComponent) Mount(socket *liveview.Socket) error {
	t.TemplateDir = "examples/templates"
	socket.Assign(map[string]interface{}{
		"todos":     []TodoItem{},
		"newTodo":   "",
		"nextID":    1,
		"filter":    "all", // all, active, completed
		"editingID": 0,
		"editText":  "",
	})
	return nil
}

// HandleAdd adds a new todo item
func (t *TodoListComponent) HandleAdd(socket *liveview.Socket, payload map[string]interface{}) error {
	text, ok := payload["text"].(string)
	if !ok || text == "" {
		return nil
	}

	todos := socket.Assigns["todos"].([]TodoItem)
	nextID := socket.Assigns["nextID"].(int)

	newTodo := TodoItem{
		ID:        nextID,
		Text:      text,
		Completed: false,
		CreatedAt: time.Now(),
	}

	socket.Assign(map[string]interface{}{
		"todos":   append(todos, newTodo),
		"newTodo": "",
		"nextID":  nextID + 1,
	})

	socket.PutFlash("success", "Todo added successfully!")
	return nil
}

// HandleToggle toggles a todo's completed status
func (t *TodoListComponent) HandleToggle(socket *liveview.Socket, payload map[string]interface{}) error {
	id, ok := payload["id"].(string)
	if !ok {
		return nil
	}

	todoID, _ := strconv.Atoi(id)
	todos := socket.Assigns["todos"].([]TodoItem)

	for i := range todos {
		if todos[i].ID == todoID {
			todos[i].Completed = !todos[i].Completed
			break
		}
	}

	socket.Assign(map[string]interface{}{
		"todos": todos,
	})
	return nil
}

// HandleDelete removes a todo item
func (t *TodoListComponent) HandleDelete(socket *liveview.Socket, payload map[string]interface{}) error {
	id, ok := payload["id"].(float64)
	if !ok {
		return nil
	}

	todoID := int(id)
	todos := socket.Assigns["todos"].([]TodoItem)
	filtered := []TodoItem{}

	for _, todo := range todos {
		if todo.ID != todoID {
			filtered = append(filtered, todo)
		}
	}

	socket.Assign(map[string]interface{}{
		"todos": filtered,
	})

	socket.PutFlash("info", "Todo deleted")
	return nil
}

// HandleFilter changes the filter view
func (t *TodoListComponent) HandleFilter(socket *liveview.Socket, payload map[string]interface{}) error {
	filter, ok := payload["filter"].(string)
	if !ok {
		return nil
	}

	socket.Assign(map[string]interface{}{
		"filter": filter,
	})
	return nil
}

// HandleClearCompleted removes all completed todos
func (t *TodoListComponent) HandleClearCompleted(socket *liveview.Socket, payload map[string]interface{}) error {
	var active []TodoItem
	socket.Assign(map[string]interface{}{
		"todos": active,
	})

	socket.PutFlash("info", "Completed todos cleared")
	return nil
}

// Render returns the HTML for the todo list
func (t *TodoListComponent) Render(socket *liveview.Socket) (template.HTML, error) {
	todos := socket.Assigns["todos"].([]TodoItem)
	filter := socket.Assigns["filter"].(string)

	// Filter todos based on current filter
	filteredTodos := []TodoItem{}
	for _, todo := range todos {
		if filter == "all" {
			filteredTodos = append(filteredTodos, todo)
		} else if filter == "active" && !todo.Completed {
			filteredTodos = append(filteredTodos, todo)
		} else if filter == "completed" && todo.Completed {
			filteredTodos = append(filteredTodos, todo)
		}
	}

	// Count stats
	activeCount := 0
	completedCount := 0
	for _, todo := range todos {
		if !todo.Completed {
			activeCount++
		} else {
			completedCount++
		}
	}

	// Prepare template data
	data := map[string]interface{}{
		"todos":          todos,
		"filteredTodos":  filteredTodos,
		"filter":         filter,
		"activeCount":    activeCount,
		"completedCount": completedCount,
	}

	return t.TemplateComponent.Render("todo.html", data)
}
