package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PrepareCodex preserves provider routing while excluding host integrations.
// The returned cleanup must run after the container exits.
func PrepareCodex(files []string) (string, func(), error) {
	for _, path := range files {
		if filepath.Base(path) != "config.toml" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", nil, err
		}
		var config map[string]any
		if err := toml.Unmarshal(data, &config); err != nil {
			return "", nil, fmt.Errorf("parse Codex configuration: %w", err)
		}
		sanitizeCodex(config)
		data, err = toml.Marshal(config)
		if err != nil {
			return "", nil, err
		}
		dir, err := os.MkdirTemp("", "agentdock-config-*")
		if err != nil {
			return "", nil, err
		}
		cleanup := func() { _ = os.RemoveAll(dir) }
		// The Docker destination resolver recognizes this relative config path.
		parent := filepath.Join(dir, ".codex")
		if err := os.Mkdir(parent, 0o755); err != nil {
			cleanup()
			return "", nil, err
		}
		target := filepath.Join(parent, "config.toml")
		// Container UID differs from the host UID; the parent temp directory is
		// private on the host while the individually mounted file is readable.
		if err := os.WriteFile(target, data, 0o644); err != nil {
			cleanup()
			return "", nil, err
		}
		return target, cleanup, nil
	}
	return "", func() {}, nil
}

func sanitizeCodex(config map[string]any) {
	// Agent Dock already isolates the process in Docker. Give Codex full access
	// inside that container and avoid approval prompts for container commands.
	config["sandbox_mode"] = "danger-full-access"
	config["approval_policy"] = "never"
	// Start from no MCP servers. Disabled entries still need a valid transport
	// in Codex's schema, so remove server entries rather than creating partial
	// {enabled=false} tables that fail validation.
	delete(config, "mcp_servers")
	delete(config, "computer_use")
	// Plugins can register MCP servers, including bundled computer use/apps.
	config["plugins"] = map[string]any{
		"codex-app-tools@openai-bundled": map[string]any{"enabled": false},
	}
	features, ok := config["features"].(map[string]any)
	if !ok {
		features = make(map[string]any)
		config["features"] = features
	}
	features["computer_use"] = false
	features["memories"] = false
	features["apps"] = false
	delete(features, "memory_tool")
	config["memories"] = map[string]any{
		"use_memories":      false,
		"generate_memories": false,
		"dedicated_tools":   false,
	}
	projects, ok := config["projects"].(map[string]any)
	if !ok {
		projects = make(map[string]any)
		config["projects"] = projects
	}
	projects["/workspace"] = map[string]any{"trust_level": "trusted"}
	if profiles, ok := config["profiles"].(map[string]any); ok {
		for _, profile := range profiles {
			if profile, ok := profile.(map[string]any); ok {
				delete(profile, "mcp_servers")
				delete(profile, "computer_use")
				if profileFeatures, ok := profile["features"].(map[string]any); ok {
					profileFeatures["computer_use"] = false
					profileFeatures["memories"] = false
					profileFeatures["apps"] = false
					delete(profileFeatures, "memory_tool")
				}
			}
		}
	}
}
