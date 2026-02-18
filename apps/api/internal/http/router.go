package http

import "net/http"

type Router struct {
	routes []route
}

type route struct {
	pattern string
	handler http.Handler
}

func NewRouter() *Router { return &Router{} }

func (r *Router) Handle(pattern string, h http.Handler) {
	r.routes = append(r.routes, route{pattern, h})
}

func (r *Router) Register(mux *http.ServeMux) {
	for _, rt := range r.routes {
		if rt.handler == nil {
			continue
		}
		mux.Handle(rt.pattern, rt.handler)
	}
}

func (r *Router) HandleWithMiddleware(pattern string, h http.Handler, mw ...func(http.Handler) http.Handler) {
	for _, m := range mw {
		h = m(h)
	}
	r.Handle(pattern, h)
}
