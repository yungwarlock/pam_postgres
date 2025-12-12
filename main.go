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
	dbHost      = os.Getenv("DB_HOST")
	dbPort      = os.Getenv("DB_PORT")
	Debug       = os.Getenv("DEBUG") != ""
	rootUser    = os.Getenv("DB_ROOT_USER")
	rootPass    = os.Getenv("DB_ROOT_PASSWORD")
	dbAdminName = os.Getenv("DB_ADMIN_DATABASE")
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Port = os.Getenv("PORT")
	Debug = os.Getenv("DEBUG") != ""
}

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

	model := requestaccess.NewRequestAccessModel(
		rootUser,
		rootPass,
		dbHost,
		dbPort,
		dbAdminName,
	)
	if model == nil {
		log.Fatal("Failed to create RequestAccessModel")
	}

	requestAccessService := requestaccess.NewRequestAccessService(model)
	requestAccessService.SetupRoutes(mux)

	var apiHandler http.Handler
	if Debug {
		apiHandler = mux
	} else {
		apiHandler = security.CSRFMiddleware(security.CSRFValidationMiddleware(mux))
	}

	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))
	dashboardHandler := security.CSRFMiddleware(reactFileServer(dashboardApp))
	mux.Handle("/", dashboardHandler)

	log.Printf("Starting pam_postgres on %s", addr)
	srv.ListenAndServe()
}
