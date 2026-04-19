package main

import (
	"fintech-labs/db"
	"fintech-labs/handlers"
	"log"
	"net/http"
)

func main() {
	db.InitDB()

	// Static files
	fs := http.FileServer(http.Dir("../frontend/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", handlers.Login(db.DB))
	http.HandleFunc("/login", handlers.Login(db.DB))
	http.HandleFunc("/register-page", handlers.RegisterPage)
	http.HandleFunc("/register", handlers.Register(db.DB)) // Standardized route

	// Dashboard & Transactions
	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
	http.HandleFunc("/logout", handlers.AuthMiddleware(handlers.Logout))
	http.HandleFunc("/transfer", handlers.AuthMiddleware(handlers.SendMoneyHandler))

	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}