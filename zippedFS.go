package main

import (
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

func ZippedFileServer(source string, target string) (http.Handler, error) {
	fsys := os.DirFS(source)
	err := os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return nil, err
	}
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
		if err := gzipFile(srcFile, zipFile); err != nil {
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

func gzipFile(source, target *os.File) error {
	w, err := gzip.NewWriterLevel(target, gzip.BestCompression)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, source)
	if err != nil {
		w.Close()
		return err
	}
	return w.Close()
}
