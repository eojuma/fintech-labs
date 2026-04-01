package main

import (
    "fintech-labs/db"
    "fintech-labs/handlers"
    "log"
    "net/http"
)

func main() {
    // Initialize SQLite Database
    db.InitDB()

    // Serve Static Files (CSS)
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))

    // UI Routes (Browser)
    http.HandleFunc("/login", handlers.LoginHandler(db.DB))
    http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
    http.HandleFunc("/logout", handlers.LogoutHandler)
    
    // Serve registration page
    http.HandleFunc("/register-page", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "templates/register.html")
    })

    // API Routes (Functional)
    http.HandleFunc("/deposit", handlers.Deposits)
    http.HandleFunc("/withdraw", handlers.Withdrawals)
    http.HandleFunc("/balance", handlers.Balances)
    http.HandleFunc("/transactions", handlers.Transactions)
    http.HandleFunc("/accounts", handlers.GetAccounts)
    
    // Registration API
    http.HandleFunc("/register", handlers.Register(db.DB))

    log.Println("========================================")
    log.Println("🚀 FinTech Banking Application Started")
    log.Println("📍 Server running on http://localhost:8080")
    log.Println("📝 Login page: http://localhost:8080/login")
    log.Println("📝 Register page: http://localhost:8080/register-page")
    log.Println("========================================")
    log.Fatal(http.ListenAndServe(":8080", nil))
}