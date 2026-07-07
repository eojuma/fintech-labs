package router

import (
	"net/http"

	"fintech-labs/internal/db"
	"fintech-labs/internal/handlers"
)

func Setup() {
	// 1. STATIC ASSETS
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. PUBLIC & AUTH ROUTES
	http.HandleFunc("/", handlers.Login(db.DB)) // Default landing
	http.HandleFunc("/login", handlers.Login(db.DB))
	http.HandleFunc("/register-page", handlers.RegisterPage)
	http.HandleFunc("/register", handlers.Register(db.DB))
	http.HandleFunc("/register-admin", handlers.AdminRegister(db.DB))
	http.HandleFunc("/logout", handlers.Logout) // Middleware optional for logout

	// 3. PROTECTED USER ROUTES (Dashboard & Money)
	http.HandleFunc("/dashboard", handlers.AuthMiddleware(handlers.DashboardHandler))
	http.HandleFunc("/deposit", handlers.AuthMiddleware(handlers.Deposit))
	http.HandleFunc("/transfer", handlers.AuthMiddleware(handlers.SendMoneyHandler))
	http.HandleFunc("/withdraw", handlers.AuthMiddleware(handlers.Withdraw))

	// 4. PROTECTED ADMIN ROUTES
	http.HandleFunc("/admin", handlers.AdminAuthMiddleware(handlers.AdminDashboardHandler))
	http.HandleFunc("/admin/deposit", handlers.AdminAuthMiddleware(handlers.AdminDepositHandler))
	http.HandleFunc("/admin/withdraw", handlers.AdminAuthMiddleware(handlers.AdminWithdrawHandler))
	http.HandleFunc("/session/refresh", handlers.AuthMiddleware(handlers.RefreshSession))
	http.HandleFunc("/admin/toggle", handlers.AdminAuthMiddleware(handlers.AdminToggleAccount))

	// 5. PROFILE ROUTE
	http.HandleFunc("/profile", handlers.AuthMiddleware(handlers.ProfileHandler))
	http.HandleFunc("/profile/update", handlers.AuthMiddleware(handlers.UpdateProfileHandler))
	http.HandleFunc("/profile/change-pin", handlers.AuthMiddleware(handlers.ChangePinHandler))
	http.HandleFunc("/profile/change-password", handlers.AuthMiddleware(handlers.ChangePasswordHandler))

	// other accounts
	http.HandleFunc("/accounts/open", handlers.AuthMiddleware(handlers.OpenSavingsAccountHandler))

	// download statement,history and reciepts
	http.HandleFunc("/statement/download", handlers.AuthMiddleware(handlers.DownloadStatementHandler))
	http.HandleFunc("/receipt/", handlers.AuthMiddleware(handlers.ReceiptHandler))
	http.HandleFunc("/transactions/filter", handlers.AuthMiddleware(handlers.FilterTransactionsHandler))
}
