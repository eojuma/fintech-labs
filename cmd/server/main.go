package main

import (
	"log"
	"net/http"
	"os"

	"fintech-labs/internal/db"
	"fintech-labs/internal/router"
	"fintech-labs/internal/services"

	"github.com/joho/godotenv"
)

func httpsRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("RENDER") == "true" && r.Header.Get("X-Forwarded-Proto") == "http" {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load .env for local development. In production (Render, etc.),
	// real environment variables are injected by the platform, so a
	// missing .env file there is expected and not fatal.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — relying on system environment variables")
	}

	db.InitDB()
	services.StartDistributionScheduler()
	services.StartLoanCollectionScheduler()
	router.Setup()
	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", httpsRedirect(http.DefaultServeMux)))
}