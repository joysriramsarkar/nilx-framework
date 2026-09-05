package ssr

import (
	"strings"
	"testing"
)

func TestSSRRenderer(t *testing.T) {
	renderer := New("Alap Dashboard")
	state := map[string]interface{}{
		"count": 42,
		"theme": "dark",
	}

	htmlOut := renderer.RenderPage("<h1>Hello SSR</h1>", state)
	if !strings.Contains(htmlOut, "<title>Alap Dashboard</title>") {
		t.Errorf("missing title in SSR output")
	}
	if !strings.Contains(htmlOut, "<h1>Hello SSR</h1>") {
		t.Errorf("missing content in SSR output")
	}
	if !strings.Contains(htmlOut, `window.__NILANG_INITIAL_STATE__`) {
		t.Errorf("missing hydration script in SSR output")
	}
}
