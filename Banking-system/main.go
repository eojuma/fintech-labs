package main

import (
	"fintech-labs/db"
	"fintech-labs/handlers"
	"log"
	"net/http"
)

func main() {
	// Initialize database
	db.InitDB()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	http.HandleFunc("/", handlers.Login(db.DB))
	http.HandleFunc("/login", handlers.Login(db.DB))
	http.HandleFunc("/register-page", handlers.RegisterPage)
	http.HandleFunc("/api/register", handlers.Register(db.DB))

	// Customer routes (require authentication)
	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
	http.HandleFunc("/logout", handlers.AuthMiddleware(handlers.Logout))
	http.HandleFunc("/api/deposit", handlers.AuthMiddleware(handlers.Deposit))
	http.HandleFunc("/api/withdraw", handlers.AuthMiddleware(handlers.Withdraw))
	http.HandleFunc("/api/balance", handlers.AuthMiddleware(handlers.GetBalance))
	http.HandleFunc("/api/transactions", handlers.AuthMiddleware(handlers.GetTransactionsAPI))
	http.HandleFunc("/api/send", handlers.AuthMiddleware(handlers.SendMoneyHandler))

	// Admin routes (require admin authentication)
	http.HandleFunc("/admin", handlers.AdminAuthMiddleware(handlers.AdminDashboardHandler))
	http.HandleFunc("/admin/api/toggle-account", handlers.AdminAuthMiddleware(handlers.AdminToggleAccount))
	http.HandleFunc("/admin/api/deposit", handlers.AdminAuthMiddleware(handlers.AdminDepositHandler))
	http.HandleFunc("/admin/api/withdraw", handlers.AdminAuthMiddleware(handlers.AdminWithdrawHandler))

	log.Println("🚀 Server running on http://localhost:8080")
	log.Println("📝 Register: http://localhost:8080/register-page")
	log.Println("🔐 Login: http://localhost:8080/login")
	log.Println("👑 Admin panel: http://localhost:8080/admin (admin users only)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}