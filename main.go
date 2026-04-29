package main

import (
	"fintech-labs/backend/db"
	"fintech-labs/backend/handlers"
	"log"
	"net/http"
)

func main() {
    db.InitDB()

    // 1. STATIC ASSETS
    fs := http.FileServer(http.Dir("frontend/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // 2. PUBLIC & AUTH ROUTES
    http.HandleFunc("/", handlers.Login(db.DB))           // Default landing
    http.HandleFunc("/login", handlers.Login(db.DB))
    http.HandleFunc("/register-page", handlers.RegisterPage)
    http.HandleFunc("/register", handlers.Register(db.DB))
    http.HandleFunc("/logout", handlers.Logout)           // Middleware optional for logout

    // 3. PROTECTED USER ROUTES (Dashboard & Money)
    http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
    http.HandleFunc("/deposit", handlers.AuthMiddleware(handlers.Deposit))
    http.HandleFunc("/transfer", handlers.AuthMiddleware(handlers.SendMoneyHandler))
 http.HandleFunc("/withdraw", handlers.AuthMiddleware(handlers.Withdraw))

    // 4. PROTECTED ADMIN ROUTES
    http.HandleFunc("/admin", handlers.AdminAuthMiddleware(handlers.AdminDashboardHandler))
    http.HandleFunc("/admin/deposit", handlers.AdminAuthMiddleware(handlers.AdminDepositHandler))
    http.HandleFunc("/admin/withdraw", handlers.AdminAuthMiddleware(handlers.AdminWithdrawHandler))

    log.Println("🚀 Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}