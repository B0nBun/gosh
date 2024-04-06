package main

import (
	"compress/gzip"
	"flag"
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
	db, err := dbservice.MakeDBService()
	if err != nil {
		log.Fatal(err)
	}

	zip := flag.Bool("zip", false, "set this flag to compress static files ahead of time")
	flag.Parse()

	var fs http.Handler
	if *zip {
		log.Println("Compressing static files ahead of time")
		fs, err = ZippedFileServer("./static", "./static-zipped")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Using uncompressed static files")
		fs = Gzip(gzip.DefaultCompression, http.FileServer(http.Dir("./static")))
	}
	mux.Handle("GET", "/static/**", NoTrailingSlash(http.StripPrefix("/static/", fs)))
	mux.Get("/", Gzip(gzip.DefaultCompression, IndexPageHandler(db)))
	mux.Post("/", Gzip(gzip.DefaultCompression, CreateLinkHandler(db)))
	mux.Get("/*", RedirectHandler(db))

	log.Printf("Started server")
	http.ListenAndServe(":1234", &mux)
}

func IndexPageHandler(db dbservice.DBService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
