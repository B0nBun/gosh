package main

import (
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"github.com/tdewolff/minify/v2"
	mcss "github.com/tdewolff/minify/v2/css"
	mjs "github.com/tdewolff/minify/v2/js"
)

func ZippedFileServer(source string, target string) (http.Handler, error) {
	fsys := os.DirFS(source)
	err := os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return nil, err
	}
	m := getMinifier()
	err = fs.WalkDir(fsys, ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		srcPath := path.Join(source, relPath)
		newPath := path.Join(target, relPath)
		if d.IsDir() {
			err := os.MkdirAll(newPath, os.ModePerm)
			return err
		}
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		zipFile, err := os.Create(newPath)
		if err != nil {
			return err
		}
		if err := minifyAndGzip(m, srcFile, zipFile); err != nil {
			return err
		}
		if err := zipFile.Close(); err != nil {
			return err
		}
		if err := srcFile.Close(); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	zippedFS := http.FileServer(http.Dir(target))
	normalFS := http.FileServer(http.Dir(source))
	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			normalFS.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		zippedFS.ServeHTTP(w, r)
	}
	return handler, nil
}

const (
	mimeCSS = "text/css"
	mimeJS = "text/javascript"
)

func getMinifier() *minify.M {
	m := minify.New()
	m.AddFunc(mimeCSS, mcss.Minify)
	m.AddFunc(mimeJS, mjs.Minify)
	return m
}

func mediaType(filename string) (mediaTy string, canMinify bool) {
	canMinify = true
	switch path.Ext(filename) {
	case ".css":
		mediaTy = mimeCSS
	case ".js":
		mediaTy = mimeJS
	default:
		canMinify = false
	}
	return
}

func minifyAndGzip(m *minify.M, source, target *os.File) error {
	gzipw, err := gzip.NewWriterLevel(target, gzip.BestCompression)
	if err != nil {
		return err
	}
	mediatype, canMinify := mediaType(source.Name())
	if canMinify {
		if err := m.Minify(mediatype, gzipw, source); err != nil {
			gzipw.Close()
			return err
		}
	} else {
		if _, err := io.Copy(gzipw, source); err != nil {
			gzipw.Close()
			return err
		}
	}
	return gzipw.Close()
}
