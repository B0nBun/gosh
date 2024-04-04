package main

import "net/http"
import "fmt"
import "gosh/router"

func main() {
	mux := router.NewRouterMux()

	mux.Handle("GET", "/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		fmt.Fprintf(w, "Get root")
	})

	mux.Handle("POST", "/", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		fmt.Fprintf(w, "Post root")
	})

	mux.Handle("GET", "/:slug", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		fmt.Fprintf(w, "Got slug '%s'", params["slug"])
	})

	fmt.Printf("Started\n")
	http.ListenAndServe(":1234", &mux)
}
