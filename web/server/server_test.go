package server

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joysriramsarkar/alap-framework/web/router"
)

func TestAlapWebServer(t *testing.T) {
	srv := New("AlapApp")

	srv.Router.GET("/api/ping", func(ctx *router.Context) (interface{}, error) {
		return map[string]string{"message": "pong"}, nil
	})

	req := httptest.NewRequest("GET", "/api/ping", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "pong") {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}

	// Verify security headers
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff header")
	}
}
