package liveview

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRule represents a validation rule for a field
type ValidationRule[T any] func(value T) error

// FieldValidator manages validation for a single field
type FieldValidator[T any] struct {
	Rules []ValidationRule[T]
}

// NewFieldValidator creates a new field validator
func NewFieldValidator[T any]() *FieldValidator[T] {
	return &FieldValidator[T]{
		Rules: make([]ValidationRule[T], 0),
	}
}

// AddRule adds a validation rule
func (v *FieldValidator[T]) AddRule(rule ValidationRule[T]) *FieldValidator[T] {
	v.Rules = append(v.Rules, rule)
	return v
}

// Validate runs all validation rules
func (v *FieldValidator[T]) Validate(value T) error {
	for _, rule := range v.Rules {
		if err := rule(value); err != nil {
			return err
		}
	}
	return nil
}

// Common validation rules for strings
func Required(fieldName string) ValidationRule[string] {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		return nil
	}
}

func MinLength(min int) ValidationRule[string] {
	return func(value string) error {
		if len(value) < min {
			return fmt.Errorf("must be at least %d characters", min)
		}
		return nil
	}
}

func MaxLength(max int) ValidationRule[string] {
	return func(value string) error {
		if len(value) > max {
			return fmt.Errorf("must be at most %d characters", max)
		}
		return nil
	}
}

func Email() ValidationRule[string] {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return func(value string) error {
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}
}

func Pattern(pattern string, message string) ValidationRule[string] {
	regex := regexp.MustCompile(pattern)
	return func(value string) error {
		if !regex.MatchString(value) {
			return fmt.Errorf("%s", message)
		}
		return nil
	}
}

func Numeric() ValidationRule[string] {
	return func(value string) error {
		value = strings.TrimSpace(value)
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("must be a number")
		}
		return nil
	}
}

func Min(min float64) ValidationRule[string] {
	return func(value string) error {
		num, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if num < min {
			return fmt.Errorf("must be at least %.2f", min)
		}
		return nil
	}
}

func Max(max float64) ValidationRule[string] {
	return func(value string) error {
		num, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if num > max {
			return fmt.Errorf("must be at most %.2f", max)
		}
		return nil
	}
}

// Validation rule for booleans
func MustBeTrue(message string) ValidationRule[bool] {
	return func(value bool) error {
		if !value {
			return fmt.Errorf("%s", message)
		}
		return nil
	}
}

// FormValidator manages validation for all form fields
type FormValidator[T any] struct {
	validators map[string]func(*T) error
}

// NewFormValidator creates a new form validator
func NewFormValidator[T any]() *FormValidator[T] {
	return &FormValidator[T]{
		validators: make(map[string]func(*T) error),
	}
}

// AddFieldValidator adds a field validator
func (fv *FormValidator[T]) AddFieldValidator(fieldName string, validator func(*T) error) *FormValidator[T] {
	fv.validators[fieldName] = validator
	return fv
}

// Validate validates the entire form
func (fv *FormValidator[T]) Validate(data *T) map[string]string {
	errors := make(map[string]string)
	for fieldName, validator := range fv.validators {
		if err := validator(data); err != nil {
			errors[fieldName] = err.Error()
		}
	}
	return errors
}

// ValidateField validates a single field
func (fv *FormValidator[T]) ValidateField(fieldName string, data *T) error {
	if validator, ok := fv.validators[fieldName]; ok {
		return validator(data)
	}
	return nil
}