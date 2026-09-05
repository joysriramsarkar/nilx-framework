package router

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type HandlerFunc func(ctx *Context) (interface{}, error)
type Middleware func(next HandlerFunc) HandlerFunc

type Context struct {
	Method     string
	Path       string
	Params     map[string]string
	Query      map[string]string
	Headers    map[string]string
	Cookies    map[string]string
	Body       interface{}
	Store      map[string]interface{}
	StatusCode int
	IsAborted  bool
	TraceID    string
	UserID     string
}

func NewContext(method, path string) *Context {
	return &Context{
		Method:     strings.ToUpper(method),
		Path:       path,
		Params:     make(map[string]string),
		Query:      make(map[string]string),
		Headers:    make(map[string]string),
		Cookies:    make(map[string]string),
		Store:      make(map[string]interface{}),
		StatusCode: 200,
	}
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) ParamInt(key string) (int, error) {
	val, ok := c.Params[key]
	if !ok {
		return 0, fmt.Errorf("param %s not found", key)
	}
	return strconv.Atoi(val)
}

type RouteInfo struct {
	Method  string
	Pattern string
}

type Route struct {
	Method      string
	Pattern     string
	Handler     HandlerFunc
	Middlewares []Middleware
}

type radixNode struct {
	part        string
	children    []*radixNode
	isParam     bool
	paramName   string
	constraint  string
	isWildcard  bool
	handlers    map[string]HandlerFunc
	middlewares map[string][]Middleware
	pattern     string
}

func newRadixNode(part string) *radixNode {
	return &radixNode{
		part:        part,
		children:    make([]*radixNode, 0),
		handlers:    make(map[string]HandlerFunc),
		middlewares: make(map[string][]Middleware),
	}
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Router struct {
	root        *radixNode
	routes      []RouteInfo
	middlewares []Middleware
	mu          sync.RWMutex
}

func New() *Router {
	return &Router{
		root:        newRadixNode("/"),
		routes:      make([]RouteInfo, 0),
		middlewares: make([]Middleware, 0),
	}
}

func (r *Router) Use(mw Middleware) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, mw)
	return r
}

func (r *Router) Group(prefix string, mw ...Middleware) *RouteGroup {
	return &RouteGroup{
		router:      r,
		prefix:      strings.TrimRight(prefix, "/"),
		middlewares: mw,
	}
}

func (r *Router) AddRoute(method, pattern string, handler HandlerFunc, middlewares ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	method = strings.ToUpper(method)
	cleanPattern := "/" + strings.Trim(pattern, "/")
	parts := split(cleanPattern)

	combined := make([]Middleware, 0, len(r.middlewares)+len(middlewares))
	combined = append(combined, r.middlewares...)
	combined = append(combined, middlewares...)

	curr := r.root
	for _, part := range parts {
		var child *radixNode
		isWildcard := strings.HasPrefix(part, "*")
		isParam := false
		paramName := ""
		constraint := ""

		if !isWildcard {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				isParam = true
				raw := part[1 : len(part)-1]
				if idx := strings.Index(raw, ":"); idx != -1 {
					paramName = raw[:idx]
					constraint = raw[idx+1:]
				} else {
					paramName = raw
				}
			} else if strings.HasPrefix(part, ":") {
				isParam = true
				paramName = part[1:]
			}
		} else {
			paramName = strings.TrimPrefix(part, "*")
		}

		for _, c := range curr.children {
			if c.isWildcard == isWildcard && c.isParam == isParam && (!isParam || c.paramName == paramName) && c.part == part {
				child = c
				break
			}
		}

		if child == nil {
			child = newRadixNode(part)
			child.isParam = isParam
			child.paramName = paramName
			child.constraint = constraint
			child.isWildcard = isWildcard
			curr.children = append(curr.children, child)
		}
		curr = child
	}

	curr.handlers[method] = handler
	curr.middlewares[method] = combined
	curr.pattern = cleanPattern

	r.routes = append(r.routes, RouteInfo{Method: method, Pattern: cleanPattern})
}

func (r *Router) GET(pattern string, handler HandlerFunc, mw ...Middleware) {
	r.AddRoute("GET", pattern, handler, mw...)
}

func (r *Router) POST(pattern string, handler HandlerFunc, mw ...Middleware) {
	r.AddRoute("POST", pattern, handler, mw...)
}

func (r *Router) PUT(pattern string, handler HandlerFunc, mw ...Middleware) {
	r.AddRoute("PUT", pattern, handler, mw...)
}

func (r *Router) DELETE(pattern string, handler HandlerFunc, mw ...Middleware) {
	r.AddRoute("DELETE", pattern, handler, mw...)
}

func (r *Router) Match(method, path string) (*Route, map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	method = strings.ToUpper(method)
	cleanPath := "/" + strings.Trim(path, "/")
	parts := split(cleanPath)
	params := make(map[string]string)

	node := r.search(r.root, parts, 0, params)
	if node == nil {
		return nil, nil, false
	}

	handler, ok := node.handlers[method]
	if !ok {
		handler, ok = node.handlers["*"]
		if !ok {
			if method == "OPTIONS" && len(node.handlers) > 0 {
				for _, anyMW := range node.middlewares {
					return &Route{
						Method:      method,
						Pattern:     node.pattern,
						Handler:     func(c *Context) (interface{}, error) { return map[string]string{"status": "ok"}, nil },
						Middlewares: anyMW,
					}, params, true
				}
			}
			return nil, nil, false
		}
	}

	return &Route{
		Method:      method,
		Pattern:     node.pattern,
		Handler:     handler,
		Middlewares: node.middlewares[method],
	}, params, true
}

func (r *Router) search(node *radixNode, parts []string, index int, params map[string]string) *radixNode {
	if index == len(parts) {
		if len(node.handlers) > 0 {
			return node
		}
		return nil
	}

	part := parts[index]

	for _, child := range node.children {
		if !child.isParam && !child.isWildcard && child.part == part {
			if res := r.search(child, parts, index+1, params); res != nil {
				return res
			}
		}
	}

	for _, child := range node.children {
		if child.isParam {
			if matchesConstraint(part, child.constraint) {
				params[child.paramName] = part
				if res := r.search(child, parts, index+1, params); res != nil {
					return res
				}
				delete(params, child.paramName)
			}
		}
	}

	for _, child := range node.children {
		if child.isWildcard {
			params[child.paramName] = strings.Join(parts[index:], "/")
			return child
		}
	}

	return nil
}

func matchesConstraint(val, constraint string) bool {
	if constraint == "" {
		return true
	}
	switch constraint {
	case "int":
		_, err := strconv.Atoi(val)
		return err == nil
	case "uuid":
		return uuidRegex.MatchString(val)
	default:
		return true
	}
}

func (r *Router) Dispatch(ctx *Context) (interface{}, error) {
	route, params, ok := r.Match(ctx.Method, ctx.Path)
	if !ok {
		return nil, fmt.Errorf("404 Not Found: %s %s", ctx.Method, ctx.Path)
	}

	ctx.Params = params

	h := route.Handler
	for i := len(route.Middlewares) - 1; i >= 0; i-- {
		h = route.Middlewares[i](h)
	}

	return h(ctx)
}

func (r *Router) Routes() []RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]RouteInfo, len(r.routes))
	copy(res, r.routes)
	return res
}

func split(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

type RouteGroup struct {
	router      *Router
	prefix      string
	middlewares []Middleware
}

func (rg *RouteGroup) Group(prefix string, mw ...Middleware) *RouteGroup {
	combined := append([]Middleware{}, rg.middlewares...)
	combined = append(combined, mw...)
	return &RouteGroup{
		router:      rg.router,
		prefix:      rg.prefix + "/" + strings.Trim(prefix, "/"),
		middlewares: combined,
	}
}

func (rg *RouteGroup) GET(path string, h HandlerFunc, mw ...Middleware) {
	combined := append([]Middleware{}, rg.middlewares...)
	combined = append(combined, mw...)
	rg.router.AddRoute("GET", rg.prefix+"/"+strings.TrimPrefix(path, "/"), h, combined...)
}

func (rg *RouteGroup) POST(path string, h HandlerFunc, mw ...Middleware) {
	combined := append([]Middleware{}, rg.middlewares...)
	combined = append(combined, mw...)
	rg.router.AddRoute("POST", rg.prefix+"/"+strings.TrimPrefix(path, "/"), h, combined...)
}
