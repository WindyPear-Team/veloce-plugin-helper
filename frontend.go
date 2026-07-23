package pluginhelper

// The frontend helpers intentionally return JSON-compatible values. This
// keeps the SDK independent from React while avoiding hand-written schema maps.
func Page(children ...any) map[string]any {
	return map[string]any{"type": "page", "children": children}
}

func Card(title string, children ...any) map[string]any {
	return map[string]any{"type": "card", "title": title, "children": children}
}

func Text(value string) map[string]any {
	return map[string]any{"type": "text", "text": value}
}

func Alert(value string) map[string]any {
	return map[string]any{"type": "alert", "text": value}
}

func JSONValue(value any) map[string]any {
	return map[string]any{"type": "json", "value": value}
}

func Button(action, label string) map[string]any {
	return map[string]any{"type": "button", "action": action, "label": label}
}

func Form(action, label string, fields ...map[string]any) map[string]any {
	return map[string]any{"type": "form", "action": action, "submit_label": label, "fields": fields}
}

func Input(name, label, defaultValue string) map[string]any {
	return map[string]any{"type": "input", "name": name, "label": label, "default": defaultValue}
}

func Textarea(name, label, defaultValue string) map[string]any {
	return map[string]any{"type": "textarea", "name": name, "label": label, "default": defaultValue}
}

func Switch(name, label string, defaultValue bool) map[string]any {
	return map[string]any{"type": "switch", "name": name, "label": label, "default": defaultValue}
}

func Select(name, label string, options ...SelectOption) map[string]any {
	return map[string]any{"type": "select", "name": name, "label": label, "options": options}
}
