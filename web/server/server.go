package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/joysriramsarkar/alap-framework/web/middleware"
	"github.com/joysriramsarkar/alap-framework/web/router"
)

type Server struct {
	Name       string
	Router     *router.Router
	httpServer *http.Server
	mu         sync.Mutex
}

func New(name string) *Server {
	r := router.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())

	return &Server{
		Name:   name,
		Router: r,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := router.NewContext(req.Method, req.URL.Path)

	for k, vals := range req.Header {
		if len(vals) > 0 {
			ctx.Headers[k] = vals[0]
		}
	}

	for k, vals := range req.URL.Query() {
		if len(vals) > 0 {
			ctx.Query[k] = vals[0]
		}
	}

	if req.Body != nil && req.ContentLength > 0 {
		b, _ := io.ReadAll(req.Body)
		_ = req.Body.Close()
		var jsonData interface{}
		if err := json.Unmarshal(b, &jsonData); err == nil {
			ctx.Body = jsonData
		} else {
			ctx.Body = string(b)
		}
	}

	res, err := s.Router.Dispatch(ctx)

	for k, v := range ctx.Headers {
		if strings.HasPrefix(k, "X-") || strings.HasPrefix(k, "Access-Control-") || k == "Content-Security-Policy" || k == "Strict-Transport-Security" {
			w.Header().Set(k, v)
		}
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		code := 500
		if strings.HasPrefix(err.Error(), "404") {
			code = 404
		} else if strings.HasPrefix(err.Error(), "401") {
			code = 401
		} else if strings.HasPrefix(err.Error(), "403") {
			code = 403
		} else if strings.HasPrefix(err.Error(), "429") {
			code = 429
		}
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(res)
}

func (s *Server) Listen(addr string) error {
	s.mu.Lock()
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s,
	}
	s.mu.Unlock()

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
