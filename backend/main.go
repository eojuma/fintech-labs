package main

import (
	"fintech-labs/db"
	"fintech-labs/handlers"
	"log"
	"net/http"
)

func main() {
	db.InitDB()

	// Static files: Assumes you run from the 'backend' folder
	fs := http.FileServer(http.Dir("../frontend/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	http.HandleFunc("/", handlers.Login(db.DB))
	http.HandleFunc("/login", handlers.Login(db.DB))
	http.HandleFunc("/register-page", handlers.RegisterPage)
	http.HandleFunc("/api/register", handlers.Register(db.DB))

	// Customer routes
	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
	http.HandleFunc("/logout", handlers.AuthMiddleware(handlers.Logout))
	http.HandleFunc("/transfer", handlers.AuthMiddleware(handlers.SendMoneyHandler))

	// Admin routes
	http.HandleFunc("/admin", handlers.AdminAuthMiddleware(handlers.AdminDashboardHandler))
	http.HandleFunc("/admin/api/deposit", handlers.AdminAuthMiddleware(handlers.AdminDepositHandler))
	http.HandleFunc("/admin/api/withdraw", handlers.AdminAuthMiddleware(handlers.AdminWithdrawHandler))

	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
