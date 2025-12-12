package main

import (
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// loggerMiddleware logs HTTP requests with method, URL, status code, and duration
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	statusCode int
	http.ResponseWriter
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// reactFileServer creates a file server that serves index.html for client-side routing
func reactFileServer(fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// For root, empty, or dashboard paths, serve index.html
		if path == "" || path == "/" {
			serveIndexHTML(w, r, fsys)
			return
		}

		// Remove leading slash for file system lookup
		cleanPath := path
		if len(cleanPath) > 0 && cleanPath[0] == '/' {
			cleanPath = cleanPath[1:]
		}

		// Try to serve the requested file if it exists
		if cleanPath != "" {
			if file, err := fsys.Open(cleanPath); err == nil {
				defer file.Close()
				if stat, err := file.Stat(); err == nil && !stat.IsDir() {
					http.FileServerFS(fsys).ServeHTTP(w, r)
					return
				}
			}
		}

		// Check if it's a request for a static asset that doesn't exist
		ext := filepath.Ext(path)
		if ext != "" && ext != ".html" {
			http.NotFound(w, r)
			return
		}

		serveIndexHTML(w, r, fsys)
	})
}

// serveIndexHTML manually reads and serves the index.html file
func serveIndexHTML(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	indexFile, err := fsys.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer indexFile.Close()

	content, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
