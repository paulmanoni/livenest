package liveview

import (
	"fmt"
	"strings"
)

// WebComponentConfig defines validation and configuration for a web component
type WebComponentConfig struct {
	TagName    string
	Attributes map[string]AttributeConfig
}

// AttributeConfig defines validation rules for an attribute
type AttributeConfig struct {
	Required bool
	Type     string // "string", "number", "boolean", "email", "url"
	Pattern  string // regex pattern for validation
	Min      *int
	Max      *int
	Default  string
}

// GenerateWebComponent generates the JavaScript for a custom web component
func GenerateWebComponent(config WebComponentConfig) string {
	return `class ${className} extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        const errors = this.validate();
        if (errors.length > 0) {
            console.error('${tagName} validation errors:', errors);
            this.shadowRoot.innerHTML = '<div style="color: red;">Validation Error: ' + errors.join(', ') + '</div>';
            return;
        }

        this.render();
    }

    validate() {
        const errors = [];
        ${validationCode}
        return errors;
    }

    render() {
        this.shadowRoot.innerHTML = '<slot></slot>';
        this.classList.add('livenest-component');
    }

    static get observedAttributes() {
        return [${observedAttrs}];
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (oldValue !== newValue) {
            this.render();
        }
    }
}

customElements.define('${tagName}', ${className});`
}

// BuildWebComponentJS builds the complete JavaScript for web components
func BuildWebComponentJS(components map[string]WebComponentConfig) string {
	var js strings.Builder

	js.WriteString("// LiveNest Web Components\n\n")

	for _, config := range components {
		className := toPascalCase(config.TagName)
		validationCode := generateValidationCode(config.Attributes)
		observedAttrs := generateObservedAttributes(config.Attributes)

		componentJS := GenerateWebComponent(config)
		componentJS = strings.ReplaceAll(componentJS, "${className}", className)
		componentJS = strings.ReplaceAll(componentJS, "${tagName}", config.TagName)
		componentJS = strings.ReplaceAll(componentJS, "${validationCode}", validationCode)
		componentJS = strings.ReplaceAll(componentJS, "${observedAttrs}", observedAttrs)

		js.WriteString(componentJS)
		js.WriteString("\n\n")
	}

	return js.String()
}

// generateValidationCode generates validation JavaScript code
func generateValidationCode(attrs map[string]AttributeConfig) string {
	var code strings.Builder

	for name, config := range attrs {
		attrVar := fmt.Sprintf("const %s = this.getAttribute('%s');", name, name)
		code.WriteString(attrVar + "\n        ")

		// Required check
		if config.Required {
			code.WriteString(fmt.Sprintf(
				"if (!%s) { errors.push('%s is required'); }\n        ",
				name, name,
			))
		}

		// Type validation
		if config.Type != "" {
			switch config.Type {
			case "number":
				code.WriteString(fmt.Sprintf(
					"if (%s && isNaN(%s)) { errors.push('%s must be a number'); }\n        ",
					name, name, name,
				))
			case "email":
				code.WriteString(fmt.Sprintf(
					"if (%s && !/^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/.test(%s)) { errors.push('%s must be a valid email'); }\n        ",
					name, name, name,
				))
			case "url":
				code.WriteString(fmt.Sprintf(
					"if (%s && !/^https?:\\/\\/.+/.test(%s)) { errors.push('%s must be a valid URL'); }\n        ",
					name, name, name,
				))
			case "boolean":
				code.WriteString(fmt.Sprintf(
					"if (%s && %s !== 'true' && %s !== 'false') { errors.push('%s must be true or false'); }\n        ",
					name, name, name, name,
				))
			}
		}

		// Pattern validation
		if config.Pattern != "" {
			code.WriteString(fmt.Sprintf(
				"if (%s && !/%s/.test(%s)) { errors.push('%s does not match required pattern'); }\n        ",
				name, config.Pattern, name, name,
			))
		}

		// Min/Max validation for numbers
		if config.Min != nil {
			code.WriteString(fmt.Sprintf(
				"if (%s && Number(%s) < %d) { errors.push('%s must be at least %d'); }\n        ",
				name, name, *config.Min, name, *config.Min,
			))
		}
		if config.Max != nil {
			code.WriteString(fmt.Sprintf(
				"if (%s && Number(%s) > %d) { errors.push('%s must be at most %d'); }\n        ",
				name, name, *config.Max, name, *config.Max,
			))
		}
	}

	return code.String()
}

// generateObservedAttributes generates the list of observed attributes
func generateObservedAttributes(attrs map[string]AttributeConfig) string {
	var attrNames []string
	for name := range attrs {
		attrNames = append(attrNames, fmt.Sprintf("'%s'", name))
	}
	return strings.Join(attrNames, ", ")
}

// toPascalCase converts kebab-case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
