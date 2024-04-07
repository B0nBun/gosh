package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriterWrap struct {
	http.ResponseWriter
	status int
	wroteHeader bool
}

func (w responseWriterWrap) Status() int {
	if w.wroteHeader {
		return w.status
	}
	return http.StatusOK
}

func (w *responseWriterWrap) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

func Logging(logger *log.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrap := responseWriterWrap { ResponseWriter: w }
		handler.ServeHTTP(&wrap, r)
		logger.Printf("status=%d method=%s path=%s duration=%v", wrap.Status(), r.Method, r.URL.EscapedPath(), time.Since(start))
	})
}
