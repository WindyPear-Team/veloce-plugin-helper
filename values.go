package pluginhelper

import (
	"encoding/json"
	"strconv"
)

type Values map[string]any

func (v Values) String(key, fallback string) string {
	value, ok := v[key]
	if !ok {
		return fallback
	}
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fallback
	}
}

func (v Values) Bool(key string, fallback bool) bool {
	value, ok := v[key]
	if !ok {
		return fallback
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	parsed, err := strconv.ParseBool(v.String(key, ""))
	if err != nil {
		return fallback
	}
	return parsed
}

func (v Values) Int(key string, fallback int) int {
	value, ok := v[key]
	if !ok {
		return fallback
	}
	if typed, ok := value.(float64); ok {
		return int(typed)
	}
	parsed, err := strconv.Atoi(v.String(key, ""))
	if err != nil {
		return fallback
	}
	return parsed
}

func (v Values) Decode(key string, destination any) error {
	raw, err := json.Marshal(v[key])
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, destination)
}
