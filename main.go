package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	requestaccess "pam_postgres/services/request_access"
	"pam_postgres/services/security"
)

var (
	Port        = os.Getenv("PORT")
	Debug       = os.Getenv("DEBUG") != ""
	DatabaseURL = os.Getenv("DATABASE_URL")
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Port = os.Getenv("PORT")
	Debug = os.Getenv("DEBUG") != ""
	DatabaseURL = os.Getenv("DATABASE_URL")
}

//go:embed dashboard/dist
var dashboardFiles embed.FS

func main() {
	dashboardApp, err := fs.Sub(dashboardFiles, "dashboard/dist")
	if err != nil {
		log.Fatalf("error finding the dist folder: %v", err)
	}

	mux := http.NewServeMux()

	db := setupDB(DatabaseURL)
	defer db.Close()

	addr := ":8080"
	if Port != "" {
		addr = ":" + Port
	}

	srv := http.Server{
		Handler: loggerMiddleware(mux),
		Addr:    addr,
	}

	requestAccessService := requestaccess.NewRequestAccessService(db)
	requestAccessService.SetupRoutes(mux)

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
