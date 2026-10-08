package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareWithoutGit(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	got, err := Prepare(root, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("got %q, want %q", got, root)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatal(err)
	}
}
