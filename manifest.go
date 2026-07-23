package pluginhelper

import "encoding/json"

const (
	PermissionWalletBalanceRead    = "wallet.balance.read"
	PermissionWalletBalanceDebit   = "wallet.balance.debit"
	PermissionWalletBalanceCredit  = "wallet.balance.credit"
	PermissionPluginKVRead         = "plugin.kv.read"
	PermissionPluginKVWrite        = "plugin.kv.write"
	PermissionPluginSettingsGlobal = "plugin.settings.global"

	HookAppBoot                      = "app.boot"
	HookAdvancedChatRuntimeExtension = "advanced_chat.runtime_extension"
	HookAdvancedChatToolCall         = "advanced_chat.tool_call"
	HookPluginActionBefore           = "plugin.action.before"
	HookPluginActionAfter            = "plugin.action.after"
	HookPluginActionError            = "plugin.action.error"
	HookPluginSettingsUpdated        = "plugin.settings.updated"
	HookPluginEnabled                = "plugin.enabled"
	HookPluginDisabled               = "plugin.disabled"
	HookPluginInstalled              = "plugin.installed"

	HookModeSync  = "sync"
	HookModeAsync = "async"
)

type Manifest struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description,omitempty"`
	Author      string         `json:"author,omitempty"`
	Permissions []string       `json:"permissions,omitempty"`
	Hooks       []Hook         `json:"hooks,omitempty"`
	Frontend    Frontend       `json:"frontend,omitempty"`
	Settings    SettingsSchema `json:"settings,omitempty"`
}

type Hook struct {
	Point    string `json:"point"`
	Mode     string `json:"mode,omitempty"`
	Action   string `json:"action,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Config   any    `json:"config,omitempty"`
}

type Frontend struct {
	Sidebar []SidebarItem `json:"sidebar,omitempty"`
	Routes  []Route       `json:"routes,omitempty"`
}

type SidebarItem struct {
	Label string `json:"label"`
	Path  string `json:"path,omitempty"`
}

type Route struct {
	Path        string `json:"path"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Page        any    `json:"page,omitempty"`
}

type SettingsSchema struct {
	Type   string          `json:"type,omitempty"`
	Tabs   []SettingsTab   `json:"tabs,omitempty"`
	Fields []SettingsField `json:"fields,omitempty"`
}

type SettingsTab struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type SettingsField struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Label       string         `json:"label"`
	Description string         `json:"description,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Default     any            `json:"default,omitempty"`
	Options     []SelectOption `json:"options,omitempty"`
	Min         any            `json:"min,omitempty"`
	Max         any            `json:"max,omitempty"`
	Step        any            `json:"step,omitempty"`
	Tab         string         `json:"tab,omitempty"`
}

type SelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (m Manifest) JSON() []byte {
	raw, _ := json.Marshal(m)
	return raw
}
