package http

import "net/http"

// Router stores route registrations before binding them to a ServeMux.
type Router struct {
	routes []route
}

type route struct {
	pattern string
	handler http.Handler
}

// NewRouter constructs an empty Router.
func NewRouter() *Router { return &Router{} }

// Handle registers a pattern and handler without middleware decoration.
func (r *Router) Handle(pattern string, h http.Handler) {
	r.routes = append(r.routes, route{pattern, h})
}

// Register binds all registered routes to the provided mux.
func (r *Router) Register(mux *http.ServeMux) {
	for _, rt := range r.routes {
		if rt.handler == nil {
			continue
		}
		mux.Handle(rt.pattern, rt.handler)
	}
}

// HandleWithMiddleware wraps a handler with middleware and registers it.
func (r *Router) HandleWithMiddleware(pattern string, h http.Handler, mw ...func(http.Handler) http.Handler) {
	for _, m := range mw {
		h = m(h)
	}
	r.Handle(pattern, h)
}
