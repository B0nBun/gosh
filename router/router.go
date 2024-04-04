package router

import "net/http"
import "strings"
import "fmt"

type Handler func(http.ResponseWriter, *http.Request, map[string]string)

type route struct {
	method  string
	pattern string
	handler Handler
}

type RouterMux struct {
	routes   []route
	notFound Handler
}

func NewRouterMux() RouterMux {
	return RouterMux{
		notFound: default404Handler,
	}
}

func (mux *RouterMux) Handle(method, pattern string, handler Handler) {
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

func (mux *RouterMux) HandleNotFound(handler Handler) {
	mux.notFound = handler
}

func (mux *RouterMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	for _, route := range mux.routes {
		if route.method != "*" && route.method != r.Method {
			continue
		}
		ok, params := match(route.pattern, parts)
		if !ok {
			continue
		}
		route.handler(w, r, params)
		break
	}

	mux.notFound(w, r, nil)
}

func default404Handler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func match(pattern string, rpath []string) (ok bool, params map[string]string) {
	split := strings.Split(pattern, "/")[1:]
	if len(split) != len(rpath) {
		ok = false
		return
	}
	params = make(map[string]string)
	for i := 0; i < len(split); i++ {
		if strings.HasPrefix(split[i], ":") {
			name := strings.TrimPrefix(split[i], ":")
			params[name] = rpath[i]
			continue
		}
		if rpath[i] != split[i] {
			ok = false
			return
		}
	}
	ok = true
	return
}
