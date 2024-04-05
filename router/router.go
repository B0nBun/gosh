package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type route struct {
	method  string
	pattern string
	handler http.Handler
}

type RouterMux struct {
	routes   []route
	notFound http.Handler
}

func NewRouterMux() RouterMux {
	return RouterMux{
		notFound: http.HandlerFunc(http.NotFound),
	}
}

func (mux *RouterMux) Handle(method, pattern string, handler http.Handler) {
	if len(pattern) == 0 || !strings.HasPrefix(pattern, "/") {
		panic(fmt.Sprintf("Invalid pattern argument '%s': Pattern has to start with '/'", pattern))
	}
	if strings.HasSuffix(pattern[1:], "/") {
		panic(fmt.Sprintf("Invalid pattern argument '%s': Pattern can't end with '/'", pattern))
	}
	mux.routes = append(mux.routes, route{
		method:  method,
		pattern: pattern,
		handler: handler,
	})
}

func (mux *RouterMux) HandleFunc(method, pattern string, handler http.HandlerFunc) {
	mux.Handle(method, pattern, handler)
}

func (mux *RouterMux) Get(pattern string, handler http.HandlerFunc) {
	mux.Handle("GET", pattern, handler)
}

func (mux *RouterMux) Post(pattern string, handler http.HandlerFunc) {
	mux.Handle("POST", pattern, handler)
}

func (mux *RouterMux) HandleNotFound(handler http.Handler) {
	mux.notFound = handler
}

func (mux *RouterMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	for _, route := range mux.routes {
		if route.method != "*" && route.method != r.Method {
			continue
		}
		ok := match(route.pattern, parts)
		if !ok {
			continue
		}
		route.handler.ServeHTTP(w, r)
		return
	}

	mux.notFound.ServeHTTP(w, r)
}

func PathPart(url *url.URL, n int) string {
	return strings.Split(url.Path, "/")[1:][n]
}

func match(pattern string, rpath []string) (ok bool) {
	split := strings.Split(pattern, "/")[1:]
	if len(split) != len(rpath) {
		ok = false
		return
	}
	for i := 0; i < len(split); i++ {
		if split[i] == "*" {
			continue
		}
		if split[i] == "**" {
			ok = true
			return
		}
		if rpath[i] != split[i] {
			ok = false
			return
		}
	}
	ok = true
	return
}
