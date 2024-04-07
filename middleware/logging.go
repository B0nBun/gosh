package middleware

import (
	"log"
	"net/http"
)

func Logging(l *log.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Adequate formatting
		l.Printf("Request %+v", r)
		handler.ServeHTTP(w, r)
	})
}
