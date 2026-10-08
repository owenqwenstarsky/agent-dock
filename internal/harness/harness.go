package harness

import (
	"os"
	"path/filepath"
)

type Harness struct {
	Name             string
	Image            string
	Dockerfile       string
	ContainerCommand []string
	ConfigCandidates []string
	MountCandidates  []string
}

// ConfigFiles returns only authentication/configuration files, never session stores.
func (h Harness) ConfigFiles(home string) []string {
	paths := make([]string, 0, len(h.MountCandidates))
	for _, relative := range h.MountCandidates {
		path := filepath.Join(home, relative)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			paths = append(paths, path)
		}
	}
	return paths
}

func (h Harness) LocalConfigFiles(home string) []string {
	paths := make([]string, 0, len(h.ConfigCandidates))
	for _, relative := range h.ConfigCandidates {
		path := filepath.Join(home, relative)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			paths = append(paths, path)
		}
	}
	return paths
}

var registry = map[string]Harness{
	"codex": {Name: "codex", Image: "agentdock/codex:0.161.2", Dockerfile: "docker/codex/Dockerfile", ContainerCommand: []string{}, ConfigCandidates: []string{".codex/auth.json", ".codex/config.toml"}, MountCandidates: []string{".codex/auth.json"}},
	"pi":    {Name: "pi", Image: "agentdock/pi:0.73.1", Dockerfile: "docker/pi/Dockerfile", ConfigCandidates: []string{".pi/agent/auth.json", ".pi/agent/models.json", ".pi/agent/settings.json"}, MountCandidates: []string{".pi/agent/auth.json", ".pi/agent/models.json", ".pi/agent/settings.json"}},
}

func Lookup(name string) (Harness, bool) { h, ok := registry[name]; return h, ok }
