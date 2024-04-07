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
	mw "gosh/middleware"
	"gosh/router"
)

func main() {
	zip := flag.Bool("zip", false, "set this flag to compress static files ahead of time")
	dsName := flag.String("ds", ":memory:", "name of the datasource to use for SQLite3 database")
	addr := flag.String("addr", "0.0.0.0:1234", "TCP address to use for the servers")
	flag.Parse()

	db, err := dbservice.MakeDBService(*dsName)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to the database")

	fs := FileServer(*zip)
	log.Printf("Created a fileserver (compressed static files = %v)", *zip)

	mux := router.NewRouterMux()
	mux.Get("/static/**", mw.Logging(log.Default(), mw.NoTrailingSlash(http.StripPrefix("/static/", fs))))
	mux.Get("/", mw.Logging(log.Default(), mw.Gzip(gzip.DefaultCompression, IndexPageHandler(&db))))
	mux.Post("/", mw.Logging(log.Default(), mw.Gzip(gzip.DefaultCompression, CreateLinkHandler(&db))))
	mux.Get("/*", mw.Logging(log.Default(), RedirectHandler(&db, mux.NotFound)))

	log.Printf("Started server at address '%s'", *addr)
	http.ListenAndServe(*addr, &mux)
}

const StaticFilesPath = "./static"
const StaticZippedFilesPath = "./static-zipped"

func FileServer(zip bool) http.Handler {
	if zip {
		fs, err := ZippedFileServer(StaticFilesPath, StaticZippedFilesPath)
		if err != nil {
			log.Fatal(err)
		}
		return fs
	}
	return mw.Gzip(gzip.DefaultCompression, http.FileServer(http.Dir(StaticFilesPath)))
}

type CreatedLink struct {
	Slug, Full, Host string
}

type LinksStats struct {
	UrlsCount, RedirectsCount int
}

func indexTemplate(w http.ResponseWriter, created *CreatedLink, db *dbservice.DBService) error {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return err
	}
	urls, err := db.TotalUrls()
	if err != nil {
		return err
	}
	visits, err := db.TotalVisits()
	if err != nil {
		return err
	}
	var tmplData struct {
		Created *CreatedLink
		Stats *LinksStats
	}
	tmplData.Created = created
	// TODO: Update these stats periodically, instead of every request
	tmplData.Stats = &LinksStats{
		UrlsCount: urls,
		RedirectsCount: visits,
	}
	return tmpl.Execute(w, &tmplData)
} 

func IndexPageHandler(db *dbservice.DBService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := indexTemplate(w, nil, db); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func CreateLinkHandler(db *dbservice.DBService) http.HandlerFunc {
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
			return
		}

		created := CreatedLink {
			Slug: slug,
			Full: clientUrl,
			Host: r.Host,
		}
		if err := indexTemplate(w, &created, db); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func RedirectHandler(db *dbservice.DBService, notFoundHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := router.PathPart(r.URL, 0)
		fullUrl, err, exists := db.GetUrl(slug)
		if !exists {
			notFoundHandler.ServeHTTP(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fullUrl, http.StatusSeeOther)
	}
}
