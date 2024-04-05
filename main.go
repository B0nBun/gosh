package main

import (
	"fmt"
	"gosh/router"
	"math/rand"
	"net/http"
	"strings"
)

// TODO: Minifiers + Gzip

func main() {
	mux := router.NewRouterMux()
	fs := http.FileServer(http.Dir("./static"))

	mux.Handle("GET", "/static/**", NoTrailingSlash(http.StripPrefix("/static/", fs)))

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Get root")
	})

	mux.Post("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Post root")
	})

	mux.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		slug := router.PathPart(r.URL, 0)
		fmt.Fprintf(w, "Got slug '%s'", slug)
	})

	fmt.Printf("Started\n")
	http.ListenAndServe(":1234", &mux)
}

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandAlphanumString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		idx := rand.Int63() % int64(len(alphanumeric))
		sb.WriteByte(alphanumeric[idx])
	}

	return sb.String()
}

func NoTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}