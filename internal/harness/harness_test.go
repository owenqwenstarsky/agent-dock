package harness

import "testing"

func TestLookup(t *testing.T) {
	for _, name := range []string{"codex", "pi"} {
		h, ok := Lookup(name)
		if !ok || h.Name != name {
			t.Fatalf("bad harness %s: %#v", name, h)
		}
	}
	if _, ok := Lookup("unknown"); ok {
		t.Fatal("unknown harness was accepted")
	}
}
