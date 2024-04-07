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
	msvg "github.com/tdewolff/minify/v2/svg"
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
	mimeSVG = "image/svg+xml"
)

func getMinifier() *minify.M {
	m := minify.New()
	m.AddFunc(mimeCSS, mcss.Minify)
	m.AddFunc(mimeJS, mjs.Minify)
	m.AddFunc(mimeSVG, msvg.Minify)
	return m
}

var mimeTypes = map[string]string {
	".css": mimeCSS,
	".js": mimeJS,
	".svg": mimeSVG,
}

func minifyAndGzip(m *minify.M, source, target *os.File) error {
	gzipw, err := gzip.NewWriterLevel(target, gzip.BestCompression)
	if err != nil {
		return err
	}
	extension := path.Ext(source.Name())
	mimeType, canMinify := mimeTypes[extension]
	if canMinify {
		if err := m.Minify(mimeType, gzipw, source); err != nil {
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
