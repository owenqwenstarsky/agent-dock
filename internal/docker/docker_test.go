package docker

import (
	"os"
	"testing"
)

func TestConfigDestination(t *testing.T) {
	if got := configDestination("/Users/test/.codex/auth.json"); got != "/home/agent/.codex/auth.json" {
		t.Fatalf("got %q", got)
	}
	if got := configDestination("/Users/test/.pi/agent/auth.json"); got != "/home/agent/.pi/agent/auth.json" {
		t.Fatalf("got %q", got)
	}
}

func TestEnvironmentFileTrimsTrailingNewline(t *testing.T) {
	file, err := environmentFile(map[string]string{"MY_ROUTER_KEY": "secret\n"})
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file)
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "MY_ROUTER_KEY=secret\n" {
		t.Fatalf("got %q", data)
	}
}

func TestEnvironmentFileRejectsEmbeddedNewline(t *testing.T) {
	if _, err := environmentFile(map[string]string{"MY_ROUTER_KEY": "sec\nret"}); err == nil {
		t.Fatal("expected embedded newline error")
	}
}

func TestOptionsCommandShape(t *testing.T) {
	o := Options{Image: "agentdock/pi:0.1.0", Workspace: "/tmp/work", Command: []string{"pi"}}
	if o.Image == "" || o.Workspace == "" || len(o.Command) != 1 {
		t.Fatal("invalid options fixture")
	}
}
