package liveview

import (
	"fmt"
	"reflect"
	"strings"
)

// BaseComponent provides a base for LiveView components
// Embedding this is optional - the framework automatically routes events to Handle* methods
type BaseComponent struct{}

// Note: BaseComponent doesn't need any methods because the socket handler
// automatically routes events using reflection on the component instance

// RouteEvent is a standalone helper that routes events to Handle* methods on any component
func RouteEvent(component interface{}, event string, payload map[string]interface{}, socket *Socket) error {
	// Convert event name to method name (e.g., "increment" -> "HandleIncrement")
	methodName := "Handle" + toTitle(event)

	// Get the component's value
	val := reflect.ValueOf(component)
	method := val.MethodByName(methodName)

	if !method.IsValid() {
		return fmt.Errorf("no handler found for event: %s (expected method: %s)", event, methodName)
	}

	// Prepare arguments
	args := []reflect.Value{
		reflect.ValueOf(socket),
		reflect.ValueOf(payload),
	}

	// Call the method
	results := method.Call(args)

	// Check if the method returned an error
	if len(results) > 0 {
		if err, ok := results[0].Interface().(error); ok && err != nil {
			return err
		}
	}

	return nil
}

// toTitle converts first character to uppercase
func toTitle(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
