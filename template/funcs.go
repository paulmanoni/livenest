package template

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// DefaultFuncs returns default template functions
func DefaultFuncs() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"trim":     strings.TrimSpace,
		"join":     strings.Join,
		"split":    strings.Split,
		"replace":  strings.ReplaceAll,
		"contains": strings.Contains,

		// Time functions
		"now":        time.Now,
		"formatDate": formatDate,
		"formatTime": formatTime,

		// Utility functions
		"default": defaultValue,
		"safe":    safe,
		"dict":    dict,
		"list":    list,

		// Math functions
		"add": add,
		"sub": sub,
		"mul": mul,
		"div": div,
		"mod": mod,

		// Comparison
		"eq":  eq,
		"ne":  ne,
		"lt":  lt,
		"lte": lte,
		"gt":  gt,
		"gte": gte,
	}
}

// formatDate formats a time.Time to a date string
func formatDate(t time.Time, format string) string {
	if format == "" {
		format = "2006-01-02"
	}
	return t.Format(format)
}

// formatTime formats a time.Time to a time string
func formatTime(t time.Time, format string) string {
	if format == "" {
		format = "15:04:05"
	}
	return t.Format(format)
}

// defaultValue returns a default value if the given value is empty
func defaultValue(defaultVal, val interface{}) interface{} {
	if val == nil || val == "" {
		return defaultVal
	}
	return val
}

// safe marks a string as safe HTML
func safe(s string) template.HTML {
	return template.HTML(s)
}

// dict creates a map from key-value pairs
func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}

	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// list creates a slice from arguments
func list(values ...interface{}) []interface{} {
	return values
}

// Math functions
func add(a, b interface{}) (interface{}, error) {
	return mathOp(a, b, func(x, y float64) float64 { return x + y })
}

func sub(a, b interface{}) (interface{}, error) {
	return mathOp(a, b, func(x, y float64) float64 { return x - y })
}

func mul(a, b interface{}) (interface{}, error) {
	return mathOp(a, b, func(x, y float64) float64 { return x * y })
}

func div(a, b interface{}) (interface{}, error) {
	return mathOp(a, b, func(x, y float64) float64 { return x / y })
}

func mod(a, b interface{}) (interface{}, error) {
	ai, aok := a.(int)
	bi, bok := b.(int)
	if !aok || !bok {
		return nil, fmt.Errorf("mod requires integer arguments")
	}
	return ai % bi, nil
}

func mathOp(a, b interface{}, op func(float64, float64) float64) (interface{}, error) {
	var af, bf float64

	switch v := a.(type) {
	case int:
		af = float64(v)
	case float64:
		af = v
	default:
		return nil, fmt.Errorf("unsupported type for math operation")
	}

	switch v := b.(type) {
	case int:
		bf = float64(v)
	case float64:
		bf = v
	default:
		return nil, fmt.Errorf("unsupported type for math operation")
	}

	return op(af, bf), nil
}

// Comparison functions
func eq(a, b interface{}) bool  { return a == b }
func ne(a, b interface{}) bool  { return a != b }
func lt(a, b interface{}) bool  { return compare(a, b) < 0 }
func lte(a, b interface{}) bool { return compare(a, b) <= 0 }
func gt(a, b interface{}) bool  { return compare(a, b) > 0 }
func gte(a, b interface{}) bool { return compare(a, b) >= 0 }

func compare(a, b interface{}) int {
	switch av := a.(type) {
	case int:
		if bv, ok := b.(int); ok {
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		}
	case float64:
		if bv, ok := b.(float64); ok {
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		}
	case string:
		if bv, ok := b.(string); ok {
			return strings.Compare(av, bv)
		}
	}
	return 0
}
