package pluginhelper

type ChatTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
}

func Message(message string, data any) map[string]any {
	result := map[string]any{"ok": true, "message": message}
	if data != nil {
		result["data"] = data
	}
	return result
}

func AdvancedChatExtension(systemPrompt string, tools ...ChatTool) map[string]any {
	return map[string]any{"system_prompt": systemPrompt, "tools": tools}
}

func ToolText(text string) map[string]any {
	return map[string]any{"text": text}
}

func AllowAction() map[string]any {
	return map[string]any{"allow": true}
}

func DenyAction(message string) map[string]any {
	return map[string]any{"deny": true, "message": message}
}
