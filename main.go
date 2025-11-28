package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"pam_postgres/services/security"
)

var (
	Port  = os.Getenv("PORT")
	Debug = os.Getenv("DEBUG") != ""
)

//go:embed dashboard/dist
var dashboardFiles embed.FS

func main() {
	dashboardApp, err := fs.Sub(dashboardFiles, "dashboard/dist")
	if err != nil {
		log.Fatalf("error finding the dist folder: %v", err)
	}

	mux := http.NewServeMux()

	addr := ":8080"
	if Port != "" {
		addr = ":" + Port
	}

	srv := http.Server{
		Handler: loggerMiddleware(mux),
		Addr:    addr,
	}

	var apiHandler http.Handler
	if Debug {
		apiHandler = mux
	} else {
		apiHandler = security.CSRFMiddleware(security.CSRFValidationMiddleware(mux))
	}

	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))
	// Handle dashboard - register both patterns
	dashboardHandler := security.CSRFMiddleware(reactFileServer(dashboardApp))
	// mux.Handle("/", http.StripPrefix("/dashboard", dashboardHandler))
	mux.Handle("/", dashboardHandler)

	log.Printf("Starting Simplipay server on %s", addr)
	srv.ListenAndServe()
}
