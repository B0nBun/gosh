package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"gosh/dbservice"
	"gosh/router"
)

// TODO: Minifiers

func main() {
	mux := router.NewRouterMux()
	fs := http.FileServer(http.Dir("./static"))
	db, err := dbservice.MakeDBService()
	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("GET", "/static/**", Gzip(NoTrailingSlash(http.StripPrefix("/static/", fs))))
	mux.Get("/", Gzip(IndexPageHandler(db)))
	mux.Post("/", Gzip(CreateLinkHandler(db)))
	mux.Get("/*", RedirectHandler(db))

	log.Printf("Started server")
	http.ListenAndServe(":1234", &mux)
}

func IndexPageHandler(db dbservice.DBService) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func CreateLinkHandler(db dbservice.DBService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Couldn't parse sent form", http.StatusBadRequest)
			return
		}
		clientUrl := r.FormValue("url")
		if parsed, err := url.Parse(clientUrl); err != nil || parsed.Host == "" {
			serr := fmt.Sprintf("'%s' is not a valid absolute URL", clientUrl)
			http.Error(w, serr, http.StatusBadRequest)
			return
		}
		slug, err := db.CreateShortenedUrl(clientUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmplData := struct {
			Slug, Full, Host string
		}{
			Slug: slug,
			Full: clientUrl,
			Host: r.Host,
		}
		if err := tmpl.Execute(w, &tmplData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func RedirectHandler(db dbservice.DBService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := router.PathPart(r.URL, 0)
		fullUrl, err := db.GetUrl(slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fullUrl, http.StatusMovedPermanently)
	}
}
