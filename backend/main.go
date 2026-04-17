package main

import (
	"fintech-labs/db"
	"fintech-labs/handlers"
	"log"
	"net/http"
)

func main() {
	db.InitDB()

	// Static files: Correctly steps out of 'backend' to find 'frontend'
	fs := http.FileServer(http.Dir("../frontend/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	http.HandleFunc("/", handlers.Login(db.DB))
	http.HandleFunc("/login", handlers.Login(db.DB))
	http.HandleFunc("/register-page", handlers.RegisterPage)
	
	// Ensure this matches your <form action="/register"> in register.html
	http.HandleFunc("/register", handlers.Register(db.DB))

	// Customer routes
	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
	http.HandleFunc("/logout", handlers.AuthMiddleware(handlers.Logout))
	
	// Standardizing transaction routes to match your transactions.go exports
	http.HandleFunc("/deposit", handlers.AuthMiddleware(handlers.Deposit))
	http.HandleFunc("/withdraw", handlers.AuthMiddleware(handlers.Withdraw))
	http.HandleFunc("/transfer", handlers.AuthMiddleware(handlers.SendMoneyHandler))

	// Admin routes - Using AdminAuthMiddleware for extra security
	http.HandleFunc("/admin", handlers.AdminAuthMiddleware(handlers.AdminDashboardHandler))
	http.HandleFunc("/admin/deposit", handlers.AdminAuthMiddleware(handlers.AdminDepositHandler))
	http.HandleFunc("/admin/withdraw", handlers.AdminAuthMiddleware(handlers.AdminWithdrawHandler))

	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}