package main

import (
	"log"
	"net/http"
	"os"

	"fintech-labs/backend/db"
	"fintech-labs/backend/router"
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
	db.InitDB()
	router.Setup()
	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", httpsRedirect(http.DefaultServeMux)))
}
