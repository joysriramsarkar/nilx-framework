package middleware

import (
	"testing"
	"time"

	"github.com/joysriramsarkar/alap-framework/web/router"
)

func TestMiddlewares(t *testing.T) {
	r := router.New()
	r.Use(Recovery())
	r.Use(RequestID())
	r.Use(SecurityHeaders())
	r.Use(RateLimit(10))

	r.GET("/panic", func(ctx *router.Context) (interface{}, error) {
		panic("crash")
	})

	r.GET("/ok", func(ctx *router.Context) (interface{}, error) {
		return "hello", nil
	})

	// Test Recovery
	ctxPanic := router.NewContext("GET", "/panic")
	_, err := r.Dispatch(ctxPanic)
	if err == nil || !ctxPanic.IsAborted || ctxPanic.StatusCode != 500 {
		t.Errorf("expected recovery from panic, got err=%v code=%d", err, ctxPanic.StatusCode)
	}

	// Test RequestID & SecurityHeaders
	ctxOK := router.NewContext("GET", "/ok")
	res, err := r.Dispatch(ctxOK)
	if err != nil || res != "hello" {
		t.Errorf("expected hello, got %v, err=%v", res, err)
	}
	if ctxOK.Headers["X-Request-ID"] == "" {
		t.Errorf("expected X-Request-ID header to be populated")
	}
	if ctxOK.Headers["X-Frame-Options"] != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options: SAMEORIGIN")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	if !rl.Allow("ip1") {
		t.Error("1st request allowed")
	}
	if !rl.Allow("ip1") {
		t.Error("2nd request allowed")
	}
	if rl.Allow("ip1") {
		t.Error("3rd request should be blocked")
	}
}
