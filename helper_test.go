package pluginhelper

import (
	"encoding/json"
	"testing"
)

func TestValuesAndManifest(t *testing.T) {
	values := Values{"name": "Veloce", "enabled": true, "count": float64(3)}
	if values.String("name", "") != "Veloce" || !values.Bool("enabled", false) || values.Int("count", 0) != 3 {
		t.Fatalf("unexpected values: %#v", values)
	}
	manifest := Manifest{ID: "test-plugin", Name: "Test", Version: "1.0.0", Permissions: []string{PermissionPluginKVRead}, Channels: []ChannelType{{ID: "acme", Name: "Acme", InboundAction: "channel.inbound", SendAction: "channel.send"}}}
	var decoded map[string]any
	if err := json.Unmarshal(manifest.JSON(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["id"] != "test-plugin" {
		t.Fatalf("manifest = %#v", decoded)
	}
	channels, ok := decoded["channels"].([]any)
	if !ok || len(channels) != 1 {
		t.Fatalf("channels = %#v", decoded["channels"])
	}
}

func TestSecureIndex(t *testing.T) {
	for i := 0; i < 100; i++ {
		index, err := SecureIndex(3)
		if err != nil {
			t.Fatal(err)
		}
		if index < 0 || index >= 3 {
			t.Fatalf("index = %d", index)
		}
	}
	if _, err := SecureIndex(0); err == nil {
		t.Fatal("expected invalid length error")
	}
}

func TestUpstreamHelpers(t *testing.T) {
	manifest := Manifest{ID: "test-plugin", Name: "Test", Version: "1.0.0", Upstreams: []UpstreamType{{
		ID: "test", Name: "Test upstream", Protocol: UpstreamProtocolResponses, PrepareAction: "upstream.prepare",
		Config: SettingsSchema{Fields: []SettingsField{{Type: "select", Name: "pool_id", Label: "Pool", Required: true, OptionsFrom: "pools", OptionLabel: "name", OptionValue: "id"}}},
	}}}
	var document map[string]any
	if err := json.Unmarshal(manifest.JSON(), &document); err != nil {
		t.Fatal(err)
	}
	upstreams, ok := document["upstreams"].([]any)
	if !ok || len(upstreams) != 1 {
		t.Fatalf("upstreams = %#v", document["upstreams"])
	}

	ctx := &ActionContext{Values: map[string]any{"channel": map[string]any{"id": float64(2), "base_url": "https://example.com", "config": map[string]any{"pool_id": "shared"}}, "request": map[string]any{"payload": map[string]any{"model": "test"}, "stream": true}}}
	invocation, err := ctx.Upstream()
	if err != nil {
		t.Fatal(err)
	}
	if invocation.Channel.ID != 2 || invocation.Channel.Config.String("pool_id", "") != "shared" || !invocation.Request.Stream {
		t.Fatalf("invocation = %#v", invocation)
	}
	request, err := JSONPostRequest(" https://example.com/v1/responses ", map[string]any{"model": "test"}, map[string]string{"Accept": "application/json"})
	if err != nil {
		t.Fatal(err)
	}
	result := UpstreamRequestResult(request).WithSettingsPatch(map[string]any{"pools": []any{}})
	if result.Request == nil || result.Request.Method != "POST" || result.Request.Headers["Content-Type"] != "application/json" || result.SettingsPatch == nil {
		t.Fatalf("result = %#v", result)
	}
}
