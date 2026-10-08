package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	if err := Run([]string{"--help"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "agentdock run") {
		t.Fatalf("unexpected help: %s", out.String())
	}
}

func TestInvalidRun(t *testing.T) {
	if err := Run([]string{"run", "wat"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected unsupported harness error")
	}
}
