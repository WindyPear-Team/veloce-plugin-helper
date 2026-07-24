package pluginhelper

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type ActionHandler func(*ActionContext) (any, error)
type HookHandler func(*HookContext) (map[string]any, error)

type Plugin struct {
	Manifest Manifest
	init     func() error
	actions  map[string]ActionHandler
	hooks    map[string]HookHandler
}

type ActionRequest struct {
	UserID    uint           `json:"user_id"`
	RequestID string         `json:"request_id,omitempty"`
	Action    string         `json:"action"`
	Payload   map[string]any `json:"payload"`
	Settings  map[string]any `json:"settings"`
}

type ActionContext struct {
	context.Context
	UserID    uint
	RequestID string
	Action    string
	Payload   map[string]any
	Values    map[string]any
	Settings  Values
	Host      *Client
}

// ChannelInbound returns the inbound payload supplied to a declared custom
// channel's InboundAction.
func (c *ActionContext) ChannelInbound() (ChannelInbound, error) {
	var inbound ChannelInbound
	if err := decodeActionPayload(c.Payload, &inbound); err != nil {
		return ChannelInbound{}, err
	}
	return inbound, nil
}

// ChannelOutbound returns the delivery payload supplied to a declared custom
// channel's SendAction.
func (c *ActionContext) ChannelOutbound() (ChannelOutbound, error) {
	var outbound ChannelOutbound
	if err := decodeActionPayload(c.Payload, &outbound); err != nil {
		return ChannelOutbound{}, err
	}
	return outbound, nil
}

type ChannelConnection struct {
	Provider      string `json:"provider"`
	TypeID        string `json:"type_id"`
	IntegrationID uint   `json:"integration_id"`
	Name          string `json:"name"`
	Config        Values `json:"config"`
	BotToken      string `json:"bot_token,omitempty"`
}

type ChannelInbound struct {
	Channel ChannelConnection `json:"channel"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type ChannelOutbound struct {
	Channel              ChannelConnection `json:"channel"`
	ExternalChatID       string            `json:"external_chat_id"`
	ExternalUserID       string            `json:"external_user_id,omitempty"`
	ExternalMessageID    string            `json:"external_message_id,omitempty"`
	InboundPayload       string            `json:"inbound_payload,omitempty"`
	Content              string            `json:"content"`
}

func (c *ActionContext) Balance() (Balance, error) {
	return c.Host.Balance(c)
}

func (c *ActionContext) Settle(input WalletSettlement) (WalletSettlementResult, error) {
	if input.IdempotencyKey == "" {
		input.IdempotencyKey = c.RequestID
	}
	return c.Host.Settle(c, input)
}

func (c *ActionContext) PreviousSettlement() (WalletSettlementResult, bool, error) {
	return c.Host.Settlement(c, c.RequestID)
}

func (c *ActionContext) KVGet(key string, destination any) (bool, error) {
	return c.Host.KVGet(c, key, destination)
}

func (c *ActionContext) KVPut(key string, value any) error {
	return c.Host.KVPut(c, key, value)
}

func (c *ActionContext) KVDelete(key string) error {
	return c.Host.KVDelete(c, key)
}

func (c *ActionContext) Log(level, message string, metadata any) error {
	return c.Host.Log(c, level, message, metadata)
}

type HookRequest struct {
	Point    string         `json:"point"`
	Action   string         `json:"action,omitempty"`
	UserID   uint           `json:"user_id,omitempty"`
	Source   string         `json:"source,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
	Hook     map[string]any `json:"hook,omitempty"`
	Settings map[string]any `json:"settings,omitempty"`
}

type HookContext struct {
	context.Context
	Point    string
	Action   string
	UserID   uint
	Source   string
	Payload  map[string]any
	Settings Values
	Host     *Client
}

func (c *HookContext) KVGet(key string, destination any) (bool, error) {
	return c.Host.KVGet(c, key, destination)
}

func (c *HookContext) KVPut(key string, value any) error {
	return c.Host.KVPut(c, key, value)
}

func (c *HookContext) KVDelete(key string) error {
	return c.Host.KVDelete(c, key)
}

func (c *HookContext) Log(level, message string, metadata any) error {
	return c.Host.Log(c, level, message, metadata)
}

func New(manifest Manifest) *Plugin {
	return &Plugin{Manifest: manifest, actions: map[string]ActionHandler{}, hooks: map[string]HookHandler{}}
}

func (p *Plugin) Init(handler func() error) *Plugin {
	p.init = handler
	return p
}

func (p *Plugin) Action(name string, handler ActionHandler) *Plugin {
	p.actions[name] = handler
	return p
}

func (p *Plugin) Hook(point, action string, handler HookHandler) *Plugin {
	p.hooks[hookKey(point, action)] = handler
	return p
}

func (p *Plugin) ExportManifest() {
	writeJSON(p.Manifest)
}

func (p *Plugin) ExportInit() {
	if p.init != nil {
		if err := p.init(); err != nil {
			writeError(err)
		}
	}
}

func (p *Plugin) ExportAction() {
	var request ActionRequest
	if err := readJSON(&request); err != nil {
		writeError(err)
		return
	}
	handler := p.actions[request.Action]
	if handler == nil {
		writeError(ErrorWithCode("action_not_found", "action is not registered: "+request.Action))
		return
	}
	values := request.Payload
	if nested, ok := request.Payload["values"].(map[string]any); ok {
		values = nested
	}
	result, err := handler(&ActionContext{Context: context.Background(), UserID: request.UserID, RequestID: request.RequestID, Action: request.Action, Payload: request.Payload, Values: values, Settings: Values(request.Settings), Host: NewClient(request.UserID)})
	if err != nil {
		writeError(err)
		return
	}
	writeActionResult(result)
}

func (p *Plugin) ExportHook() {
	var request HookRequest
	if err := readJSON(&request); err != nil {
		writeError(err)
		return
	}
	handler := p.hooks[hookKey(request.Point, request.Action)]
	if handler == nil {
		handler = p.hooks[hookKey(request.Point, "*")]
	}
	if handler == nil {
		writeJSON(map[string]any{"ok": true})
		return
	}
	result, err := handler(&HookContext{Context: context.Background(), Point: request.Point, Action: request.Action, UserID: request.UserID, Source: request.Source, Payload: request.Payload, Settings: Values(request.Settings), Host: NewClient(request.UserID)})
	if err != nil {
		writeError(err)
		return
	}
	if result == nil {
		result = map[string]any{"ok": true}
	}
	writeJSON(result)
}

func hookKey(point, action string) string { return point + "\x00" + action }

func decodeActionPayload(payload map[string]any, destination any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, destination)
}

func readJSON(destination any) error {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, destination)
}

func writeJSON(value any) {
	_ = json.NewEncoder(os.Stdout).Encode(value)
}

func writeActionResult(result any) {
	if result == nil {
		writeJSON(map[string]any{"ok": true})
		return
	}
	object, ok := result.(map[string]any)
	if !ok {
		writeError(errors.New("plugin action result must be a JSON object"))
		return
	}
	if _, exists := object["ok"]; !exists {
		object["ok"] = true
	}
	writeJSON(object)
}

func writeError(err error) {
	var coded *Error
	if errors.As(err, &coded) {
		writeJSON(map[string]any{"ok": false, "code": coded.Code, "error": coded.Message})
		return
	}
	var hostErr *HostError
	if errors.As(err, &hostErr) {
		writeJSON(map[string]any{"ok": false, "code": hostErr.Code, "error": hostErr.Message})
		return
	}
	writeJSON(map[string]any{"ok": false, "code": "plugin_error", "error": err.Error()})
}
