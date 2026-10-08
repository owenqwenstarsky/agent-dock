package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestPrepareCodexPreservesRouting(t *testing.T) {
	original := []byte(`model_provider = "router"
model = "custom-model"
[model_providers.router]
base_url = "https://example.com/v1"
env_key = "MY_ROUTER_KEY"
[mcp_servers.host]
command = "host-only"
[features]
computer_use = true
[plugins.example]
enabled = true
[profiles.work]
model_provider = "router"
[profiles.work.mcp_servers.host]
command = "host-only"
`)
	source := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(source, original, 0o600); err != nil {
		t.Fatal(err)
	}
	path, cleanup, err := PrepareCodex([]string{source})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := toml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["model_provider"] != "router" || parsed["model"] != "custom-model" {
		t.Fatal("routing lost")
	}
	if parsed["sandbox_mode"] != "danger-full-access" || parsed["approval_policy"] != "never" {
		t.Fatal("full container access not configured")
	}
	if parsed["computer_use"] != nil {
		t.Fatal("retained top-level computer use")
	}
	if _, ok := parsed["mcp_servers"]; ok {
		t.Fatal("MCP config retained")
	}
	if parsed["plugins"].(map[string]any)["codex-app-tools@openai-bundled"].(map[string]any)["enabled"] != false {
		t.Fatal("apps plugin enabled")
	}
	features := parsed["features"].(map[string]any)
	if features["computer_use"] != false || features["memories"] != false || features["apps"] != false {
		t.Fatal("disabled features enabled")
	}
	if _, ok := features["memory_tool"]; ok {
		t.Fatal("deprecated memory_tool retained")
	}
	projects := parsed["projects"].(map[string]any)
	if projects["/workspace"].(map[string]any)["trust_level"] != "trusted" {
		t.Fatal("workspace not trusted")
	}
	profile := parsed["profiles"].(map[string]any)["work"].(map[string]any)
	if _, ok := profile["mcp_servers"]; ok {
		t.Fatal("profile MCP retained")
	}
	unchanged, _ := os.ReadFile(source)
	if string(unchanged) != string(original) {
		t.Fatal("host config modified")
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("temporary config not removed")
	}
}
