package sse

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSSEStream(t *testing.T) {
	rec := httptest.NewRecorder()
	stream, err := NewStream(rec)
	if err != nil {
		t.Fatalf("failed to create SSE stream: %v", err)
	}

	err = stream.Emit("notification", map[string]string{"title": "Deploy Succeeded"})
	if err != nil {
		t.Fatalf("failed to emit SSE event: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: notification\n") {
		t.Errorf("missing event header in SSE: %s", body)
	}
	if !strings.Contains(body, `"title":"Deploy Succeeded"`) {
		t.Errorf("missing data body in SSE: %s", body)
	}
}
