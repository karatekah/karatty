package security

import "testing"

func TestTerminalInputAndBounds(t *testing.T) {
	for _, s := range []string{"", "   ", "hello\x1b[2J", "osc\x1b]52;c;SECRET\a", "line\nnext", "bidirectional\u202e", "\x9b31m", string([]byte{0xff})} {
		if Body(s) {
			t.Errorf("accepted unsafe input %q", s)
		}
	}
	if !Body("Hola 👋 — texto válido") {
		t.Fatal("rejected safe text")
	}
	if Display("A\x1b\u202eB") != "A��B" {
		t.Fatal("display sanitizer")
	}
	if Body(string(make([]byte, 4097))) {
		t.Fatal("oversize")
	}
}
func TestLimiterBounded(t *testing.T) {
	l := NewLimiter(0, 2)
	l.Capacity = 1
	if !l.Allow("a") || !l.Allow("a") || l.Allow("a") || l.Allow("b") {
		t.Fatal("limiter failed")
	}
	if len(l.entries) != 1 {
		t.Fatal("unbounded map")
	}
}
