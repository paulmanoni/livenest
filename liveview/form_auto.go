package liveview

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
)

// FormComponent automatically generates forms from struct tags
// It implements Component and EventHandler interfaces automatically
type FormComponent[T any] struct {
	validator  *FormValidator[T]
	onSubmit   func(*Socket, *T) error
	title      string
	submitText string
	showReset  bool
}

// Ensure FormComponent implements Component and EventHandler
var _ Component = (*FormComponent[struct{}])(nil)
var _ EventHandler = (*FormComponent[struct{}])(nil)

// NewFormComponent creates a form component from struct tags
func NewFormComponent[T any](title string) *FormComponent[T] {
	comp := &FormComponent[T]{
		validator:  buildValidatorFromTags[T](),
		title:      title,
		submitText: "Submit",
		showReset:  true,
	}
	return comp
}

// OnSubmit sets the submit handler
func (fc *FormComponent[T]) OnSubmit(handler func(*Socket, *T) error) *FormComponent[T] {
	fc.onSubmit = handler
	return fc
}

// Mount initializes the form component
func (fc *FormComponent[T]) Mount(socket *Socket) error {
	var formData T
	socket.Assign(map[string]interface{}{
		"formData":  formData,
		"errors":    make(map[string]string),
		"submitted": false,
	})
	return nil
}

// HandleChange handles input changes with live validation
func (fc *FormComponent[T]) HandleChange(socket *Socket, payload map[string]interface{}) error {
	field, ok := payload["field"].(string)
	if !ok {
		return fmt.Errorf("field name not provided")
	}

	value := payload["value"]

	// Get current form data
	formData, ok := socket.Assigns["formData"].(T)
	if !ok {
		var zero T
		formData = zero
	}

	errors, ok := socket.Assigns["errors"].(map[string]string)
	if !ok {
		errors = make(map[string]string)
	}

	// Update the field value
	if err := setFieldValue(&formData, field, value); err != nil {
		return err
	}

	// Validate the specific field
	if fc.validator != nil {
		if err := fc.validator.ValidateField(field, &formData); err != nil {
			errors[field] = err.Error()
		} else {
			delete(errors, field)
		}
	}

	socket.Assign(map[string]interface{}{
		"formData": formData,
		"errors":   errors,
	})

	return nil
}

// HandleSubmit handles form submission
func (fc *FormComponent[T]) HandleSubmit(socket *Socket, payload map[string]interface{}) error {
	formData, ok := socket.Assigns["formData"].(T)
	if !ok {
		return fmt.Errorf("form data not found")
	}

	// Validate all fields
	var errors map[string]string
	if fc.validator != nil {
		errors = fc.validator.Validate(&formData)
	} else {
		errors = make(map[string]string)
	}

	if len(errors) > 0 {
		socket.Assign(map[string]interface{}{
			"errors": errors,
		})
		socket.PutFlash("error", "Please fix the errors below")
		return nil
	}

	// Call custom submit handler
	if fc.onSubmit != nil {
		if err := fc.onSubmit(socket, &formData); err != nil {
			socket.PutFlash("error", err.Error())
			return nil
		}
	}

	// Form is valid and submitted
	socket.Assign(map[string]interface{}{
		"submitted": true,
		"errors":    make(map[string]string),
	})

	socket.PutFlash("success", "Form submitted successfully!")
	return nil
}

// HandleReset resets the form
func (fc *FormComponent[T]) HandleReset(socket *Socket, payload map[string]interface{}) error {
	var formData T
	socket.Assign(map[string]interface{}{
		"formData":  formData,
		"errors":    make(map[string]string),
		"submitted": false,
	})
	socket.PutFlash("info", "Form reset")
	return nil
}

// Render generates HTML from struct tags
func (fc *FormComponent[T]) Render(socket *Socket) (template.HTML, error) {
	var zero T
	fields := parseStructTags(zero)
	return fc.buildHTML(fields, socket.Assigns), nil
}

// HandleEvent handles all form events
func (fc *FormComponent[T]) HandleEvent(event string, payload map[string]interface{}, socket *Socket) error {
	switch event {
	case "change":
		return fc.HandleChange(socket, payload)
	case "submit":
		return fc.HandleSubmit(socket, payload)
	case "reset":
		return fc.HandleReset(socket, payload)
	default:
		return fmt.Errorf("unknown event: %s", event)
	}
}

// field represents a form field configuration
type field struct {
	Name        string
	Label       string
	Type        string
	Placeholder string
	Required    bool
	Min         interface{}
	Max         interface{}
	Rows        int
}

// buildHTML generates the complete HTML form
func (fc *FormComponent[T]) buildHTML(fields []field, assigns map[string]interface{}) template.HTML {
	var html strings.Builder

	submitted, _ := assigns["submitted"].(bool)
	formData := assigns["formData"]
	errors, _ := assigns["errors"].(map[string]string)

	html.WriteString(`<div class="form-container">`)
	html.WriteString(fmt.Sprintf(`<h1>%s</h1>`, fc.title))

	if submitted {
		html.WriteString(`<div class="success-message">
			<h2>✅ Form Submitted Successfully!</h2>
			<p>Thank you for your submission.</p>
			<button lv-click="reset" class="btn btn-primary">Submit Another</button>
		</div>`)
	} else {
		html.WriteString(`<form class="contact-form">`)

		for _, field := range fields {
			html.WriteString(fc.buildField(field, formData, errors))
		}

		html.WriteString(`<div class="form-actions">`)
		html.WriteString(fmt.Sprintf(`<button type="button" lv-click="submit" class="btn btn-primary">%s</button>`, fc.submitText))
		if fc.showReset {
			html.WriteString(`<button type="button" lv-click="reset" class="btn btn-secondary">Reset</button>`)
		}
		html.WriteString(`</div></form>`)
	}

	html.WriteString(`</div>`)
	html.WriteString(buildCSS())
	html.WriteString(buildScript())

	return template.HTML(html.String())
}

// buildField generates HTML for a single field
func (fc *FormComponent[T]) buildField(f field, formData interface{}, errors map[string]string) string {
	var html strings.Builder

	isCheckbox := f.Type == "checkbox"
	groupClass := "form-group"
	if isCheckbox {
		groupClass += " checkbox-group"
	}

	html.WriteString(fmt.Sprintf(`<div class="%s">`, groupClass))

	fieldValue := getFieldValue(formData, f.Name)
	hasError := errors[f.Name] != ""
	errorClass := ""
	if hasError {
		errorClass = "error"
	}

	if !isCheckbox {
		required := ""
		if f.Required {
			required = " *"
		}
		html.WriteString(fmt.Sprintf(`<label for="%s">%s%s</label>`, f.Name, f.Label, required))
	}

	switch f.Type {
	case "textarea":
		rows := f.Rows
		if rows == 0 {
			rows = 5
		}
		html.WriteString(fmt.Sprintf(
			`<textarea id="%s" rows="%d" data-field="%s" class="form-input %s" placeholder="%s">%v</textarea>`,
			f.Name, rows, f.Name, errorClass, f.Placeholder, fieldValue,
		))

	case "checkbox":
		checked := ""
		if boolVal, ok := fieldValue.(bool); ok && boolVal {
			checked = " checked"
		}
		html.WriteString(`<label>`)
		html.WriteString(fmt.Sprintf(
			`<input type="checkbox" id="%s"%s data-field="%s" />`,
			f.Name, checked, f.Name,
		))
		required := ""
		if f.Required {
			required = " *"
		}
		html.WriteString(fmt.Sprintf(`%s%s`, f.Label, required))
		html.WriteString(`</label>`)

	default:
		attrs := fmt.Sprintf(
			`type="%s" id="%s" value="%v" data-field="%s" name="%s" class="form-input %s" placeholder="%s"`,
			f.Type, f.Name, fieldValue, f.Name, f.Name, errorClass, f.Placeholder,
		)

		if f.Min != nil {
			attrs += fmt.Sprintf(` min="%v"`, f.Min)
		}
		if f.Max != nil {
			attrs += fmt.Sprintf(` max="%v"`, f.Max)
		}

		html.WriteString(fmt.Sprintf(`<input %s />`, attrs))
	}

	if hasError {
		html.WriteString(fmt.Sprintf(`<span class="error-message">%s</span>`, errors[f.Name]))
	}

	html.WriteString(`</div>`)

	return html.String()
}

// getFieldValue gets the value of a field from the form data
func getFieldValue(formData interface{}, fieldName string) interface{} {
	if formData == nil {
		return ""
	}

	v := reflect.ValueOf(formData)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return ""
	}

	return field.Interface()
}

// buildCSS generates the default CSS
func buildCSS() string {
	return `<style>
    .form-container {
        max-width: 600px;
        margin: 40px auto;
        padding: 30px;
        background: white;
        border-radius: 10px;
        box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    }
    .form-container h1 {
        text-align: center;
        color: #2c3e50;
        margin-bottom: 30px;
    }
    .contact-form {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }
    .form-group {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }
    .form-group label {
        font-weight: 600;
        color: #34495e;
        font-size: 14px;
    }
    .form-input {
        padding: 12px;
        border: 2px solid #e0e0e0;
        border-radius: 5px;
        font-size: 16px;
        transition: border-color 0.3s;
        font-family: inherit;
    }
    .form-input:focus {
        outline: none;
        border-color: #3498db;
    }
    .form-input.error {
        border-color: #e74c3c;
    }
    .error-message {
        color: #e74c3c;
        font-size: 13px;
        font-weight: 500;
    }
    .checkbox-group label {
        display: flex;
        align-items: center;
        gap: 10px;
        font-weight: 400;
        cursor: pointer;
    }
    .checkbox-group input[type="checkbox"] {
        width: 18px;
        height: 18px;
        cursor: pointer;
    }
    .form-actions {
        display: flex;
        gap: 10px;
        margin-top: 10px;
    }
    .btn {
        flex: 1;
        padding: 12px 24px;
        border: none;
        border-radius: 5px;
        font-size: 16px;
        font-weight: 600;
        cursor: pointer;
        transition: background-color 0.3s;
    }
    .btn-primary {
        background: #3498db;
        color: white;
    }
    .btn-primary:hover {
        background: #2980b9;
    }
    .btn-secondary {
        background: #95a5a6;
        color: white;
    }
    .btn-secondary:hover {
        background: #7f8c8d;
    }
    .success-message {
        text-align: center;
        padding: 40px 20px;
    }
    .success-message h2 {
        color: #27ae60;
        margin-bottom: 20px;
    }
    .success-message p {
        font-size: 16px;
        color: #34495e;
        margin: 10px 0;
    }
    .success-message button {
        margin-top: 30px;
        padding: 12px 30px;
    }
</style>`
}

// buildScript generates the JavaScript for form handling
func buildScript() string {
	// With morphdom, event listeners are preserved, so we only need to attach once
	return `<script>
	(function() {
		// Check if listeners already attached (avoid duplicates)
		if (window.__formListenersAttached) return;
		window.__formListenersAttached = true;

		// Use event delegation for efficiency and to handle dynamically added inputs
		document.addEventListener('input', function(e) {
			const field = e.target.getAttribute('data-field');
			if (field && window.liveSocket) {
				const value = e.target.type === 'checkbox' ? e.target.checked.toString() : e.target.value;
				window.liveSocket.pushEvent('change', { field, value });
			}
		});

		document.addEventListener('change', function(e) {
			const field = e.target.getAttribute('data-field');
			if (field && window.liveSocket) {
				const value = e.target.type === 'checkbox' ? e.target.checked.toString() : e.target.value;
				window.liveSocket.pushEvent('change', { field, value });
			}
		});
	})();
	</script>`
}

// parseStructTags parses struct tags to build form fields
func parseStructTags(data interface{}) []field {
	fields := make([]field, 0)
	t := reflect.TypeOf(data)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)

		// Skip unexported fields
		if !structField.IsExported() {
			continue
		}

		f := field{
			Name:  structField.Name,
			Label: structField.Name,
			Type:  "text",
		}

		// Parse form tag
		if formTag := structField.Tag.Get("form"); formTag != "" {
			parseFormTag(&f, formTag)
		}

		// Parse validate tag
		if validateTag := structField.Tag.Get("validate"); validateTag != "" {
			parseValidateTag(&f, validateTag)
		}

		// Infer type from field type if not specified
		if f.Type == "text" {
			f.Type = inferFieldType(structField.Type)
		}

		fields = append(fields, f)
	}

	return fields
}

// parseFormTag parses the form tag
// Format: form:"label:Email Address;type:email;placeholder:Enter email"
func parseFormTag(f *field, tag string) {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "label":
			f.Label = value
		case "type":
			f.Type = value
		case "placeholder":
			f.Placeholder = value
		case "rows":
			if rows, err := strconv.Atoi(value); err == nil {
				f.Rows = rows
			}
		case "-":
			// Skip this field
			f.Name = ""
		}
	}
}

// parseValidateTag parses the validate tag
// Format: validate:"required;min:3;max:100;email"
func parseValidateTag(f *field, tag string) {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "required" {
			f.Required = true
		} else if strings.HasPrefix(part, "min:") {
			if val := strings.TrimPrefix(part, "min:"); val != "" {
				if num, err := strconv.Atoi(val); err == nil {
					f.Min = num
				}
			}
		} else if strings.HasPrefix(part, "max:") {
			if val := strings.TrimPrefix(part, "max:"); val != "" {
				if num, err := strconv.Atoi(val); err == nil {
					f.Max = num
				}
			}
		} else if part == "email" {
			f.Type = "email"
		}
	}
}

// inferFieldType infers the HTML input type from Go type
func inferFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "checkbox"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "text"
	default:
		return "text"
	}
}

// buildValidatorFromTags builds a validator from struct tags
func buildValidatorFromTags[T any]() *FormValidator[T] {
	validator := NewFormValidator[T]()
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)

		if !structField.IsExported() {
			continue
		}

		validateTag := structField.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		fieldName := structField.Name
		rules := parseValidationRules(validateTag, structField.Name, structField.Type)

		if len(rules) > 0 {
			// Capture variables in closure to avoid loop variable capture bug
			capturedFieldName := fieldName
			capturedRules := rules

			validator.AddFieldValidator(capturedFieldName, func(data *T) error {
				v := reflect.ValueOf(data).Elem()
				fieldValue := v.FieldByName(capturedFieldName)

				for _, rule := range capturedRules {
					if err := rule(fieldValue.Interface()); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}

	return validator
}

// parseValidationRules parses validation rules from tag
func parseValidationRules(tag string, fieldName string, fieldType reflect.Type) []func(interface{}) error {
	rules := make([]func(interface{}) error, 0)
	parts := strings.Split(tag, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "required" {
			rules = append(rules, func(val interface{}) error {
				return Required(fieldName)(val.(string))
			})
		} else if strings.HasPrefix(part, "min:") {
			minStr := strings.TrimPrefix(part, "min:")
			if fieldType.Kind() == reflect.String {
				if minLen, err := strconv.Atoi(minStr); err == nil {
					rules = append(rules, func(val interface{}) error {
						return MinLength(minLen)(val.(string))
					})
				}
			} else {
				if minVal, err := strconv.ParseFloat(minStr, 64); err == nil {
					rules = append(rules, func(val interface{}) error {
						return Min(minVal)(fmt.Sprintf("%v", val))
					})
				}
			}
		} else if strings.HasPrefix(part, "max:") {
			maxStr := strings.TrimPrefix(part, "max:")
			if fieldType.Kind() == reflect.String {
				if maxLen, err := strconv.Atoi(maxStr); err == nil {
					rules = append(rules, func(val interface{}) error {
						return MaxLength(maxLen)(val.(string))
					})
				}
			} else {
				if maxVal, err := strconv.ParseFloat(maxStr, 64); err == nil {
					rules = append(rules, func(val interface{}) error {
						return Max(maxVal)(fmt.Sprintf("%v", val))
					})
				}
			}
		} else if part == "email" {
			rules = append(rules, func(val interface{}) error {
				return Email()(val.(string))
			})
		}
	}

	return rules
}

// setFieldValue sets a field value using reflection
func setFieldValue(data interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(data).Elem()
	field := v.FieldByName(fieldName)

	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		}
	case reflect.Bool:
		if str, ok := value.(string); ok {
			field.SetBool(str == "true")
		} else if b, ok := value.(bool); ok {
			field.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if str, ok := value.(string); ok {
			var num int64
			fmt.Sscanf(str, "%d", &num)
			field.SetInt(num)
		}
	case reflect.Float32, reflect.Float64:
		if str, ok := value.(string); ok {
			var num float64
			fmt.Sscanf(str, "%f", &num)
			field.SetFloat(num)
		}
	default:
		val := reflect.ValueOf(value)
		if val.Type().AssignableTo(field.Type()) {
			field.Set(val)
		}
	}

	return nil
}