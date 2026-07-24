package main

import (
	"encoding/json"
	"strings"

	plugin "github.com/WindyPear-Team/veloce-plugin-helper"
)

var app = plugin.New(plugin.Manifest{
	ID:      "example-custom-channel",
	Name:    "Custom Channel Example",
	Version: "0.1.0",
	Permissions: []string{
		plugin.PermissionPluginChannelHTTP,
	},
	Channels: []plugin.ChannelType{{
		ID:            "acme",
		Name:          "Acme Chat",
		Description:   "Example webhook-backed message channel.",
		InboundAction: "channel.inbound",
		SendAction:    "channel.send",
		Config: plugin.SettingsSchema{Fields: []plugin.SettingsField{
			{Type: "input", Name: "base_url", Label: "API Base URL", Required: true},
			{Type: "secret", Name: "token", Label: "API Token", Required: true},
		}},
	}},
})

func init() {
	app.Action("channel.inbound", parseInbound)
	app.Action("channel.send", sendReply)
}

func parseInbound(ctx *plugin.ActionContext) (any, error) {
	inbound, err := ctx.ChannelInbound()
	if err != nil {
		return nil, err
	}
	var event struct {
		ChatID    string `json:"chat_id"`
		UserID    string `json:"user_id"`
		UserName  string `json:"user_name"`
		MessageID string `json:"message_id"`
		Text      string `json:"text"`
	}
	if err := json.Unmarshal([]byte(inbound.Body), &event); err != nil {
		return nil, plugin.ErrorWithCode("invalid_webhook", "invalid Acme webhook JSON")
	}
	return map[string]any{
		"external_chat_id":    event.ChatID,
		"external_user_id":    event.UserID,
		"external_user_name":  event.UserName,
		"external_message_id": event.MessageID,
		"content":             event.Text,
	}, nil
}

func sendReply(ctx *plugin.ActionContext) (any, error) {
	outbound, err := ctx.ChannelOutbound()
	if err != nil {
		return nil, err
	}
	endpoint := strings.TrimRight(outbound.Channel.Config.String("base_url", ""), "/") + "/messages"
	body, err := json.Marshal(map[string]any{"chat_id": outbound.ExternalChatID, "text": outbound.Content})
	if err != nil {
		return nil, err
	}
	_, err = ctx.Host.HTTP(ctx, plugin.HTTPRequest{
		Method: "POST", URL: endpoint,
		Headers: map[string]string{"Authorization": "Bearer " + outbound.Channel.Config.String("token", ""), "Content-Type": "application/json"},
		Body:    string(body),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func main() {}

//export plugin_manifest
func manifest() { app.ExportManifest() }

//export plugin_init
func initPlugin() { app.ExportInit() }

//export plugin_handle_action
func action() { app.ExportAction() }

//export plugin_handle_hook
func hook() { app.ExportHook() }
