package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironmentReadsPiReferences(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "from-host")
	t.Setenv("CUSTOM_KEY", "custom-secret")
	dir := t.TempDir()
	auth := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(auth, []byte(`{"openai":{"type":"api_key","key":"OPENAI_API_KEY"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	models := filepath.Join(dir, "models.json")
	if err := os.WriteFile(models, []byte(`{"providers":{"custom":{"apiKey":"$CUSTOM_KEY"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Environment("pi", []string{auth, models})
	if err != nil {
		t.Fatal(err)
	}
	if got["OPENAI_API_KEY"] != "from-host" || got["CUSTOM_KEY"] != "custom-secret" {
		t.Fatalf("got %v", got)
	}
}

func TestEnvironmentDoesNotTreatLiteralAsHostEnvironment(t *testing.T) {
	t.Setenv("literal-key", "should-not-be-forwarded")
	dir := t.TempDir()
	file := filepath.Join(dir, "models.json")
	if err := os.WriteFile(file, []byte(`{"apiKey":"literal-key"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Environment("pi", []string{file})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
