package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/joysriramsarkar/alap-framework/web/router"
)

func Recovery() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (res interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					_ = string(debug.Stack())
					err = fmt.Errorf("500: internal server panic: %v", r)
					ctx.IsAborted = true
					ctx.StatusCode = 500
				}
			}()
			return next(ctx)
		}
	}
}

func RequestID() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (interface{}, error) {
			reqID := ctx.Headers["X-Request-ID"]
			if reqID == "" {
				b := make([]byte, 16)
				_, _ = rand.Read(b)
				reqID = hex.EncodeToString(b)
				ctx.Headers["X-Request-ID"] = reqID
			}
			return next(ctx)
		}
	}
}

func SecurityHeaders() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (interface{}, error) {
			ctx.Headers["X-Content-Type-Options"] = "nosniff"
			ctx.Headers["X-Frame-Options"] = "SAMEORIGIN"
			ctx.Headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
			ctx.Headers["Content-Security-Policy"] = "default-src 'self'"
			ctx.Headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
			return next(ctx)
		}
	}
}

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

func CORS(cfg CORSConfig) router.Middleware {
	origins := strings.Join(cfg.AllowOrigins, ", ")
	if origins == "" {
		origins = "*"
	}
	methods := strings.Join(cfg.AllowMethods, ", ")
	if methods == "" {
		methods = "GET, POST, PUT, DELETE, OPTIONS"
	}
	headers := strings.Join(cfg.AllowHeaders, ", ")
	if headers == "" {
		headers = "Content-Type, Authorization, X-Request-ID"
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (interface{}, error) {
			ctx.Headers["Access-Control-Allow-Origin"] = origins
			ctx.Headers["Access-Control-Allow-Methods"] = methods
			ctx.Headers["Access-Control-Allow-Headers"] = headers

			if ctx.Method == "OPTIONS" {
				return map[string]string{"status": "ok"}, nil
			}
			return next(ctx)
		}
	}
}

type RateLimiter struct {
	max  int
	win  time.Duration
	hits map[string][]time.Time
	mu   sync.Mutex
}

func NewRateLimiter(max int, win time.Duration) *RateLimiter {
	return &RateLimiter{
		max:  max,
		win:  win,
		hits: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.win)

	valid := make([]time.Time, 0)
	for _, t := range rl.hits[key] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.max {
		rl.hits[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.hits[key] = valid
	return true
}

func RateLimit(requestsPerMin int) router.Middleware {
	rl := NewRateLimiter(requestsPerMin, time.Minute)
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) (interface{}, error) {
			ip := ctx.Headers["X-Forwarded-For"]
			if ip == "" {
				ip = "127.0.0.1"
			}
			if !rl.Allow(ip) {
				return nil, fmt.Errorf("429: rate limit exceeded")
			}
			return next(ctx)
		}
	}
}
