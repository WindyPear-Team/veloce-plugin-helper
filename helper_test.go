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
	manifest := Manifest{ID: "test-plugin", Name: "Test", Version: "1.0.0", Permissions: []string{PermissionPluginKVRead}}
	var decoded map[string]any
	if err := json.Unmarshal(manifest.JSON(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["id"] != "test-plugin" {
		t.Fatalf("manifest = %#v", decoded)
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
