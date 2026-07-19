package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/models"
	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

// AdminAuthMiddleware checks authentication AND admin role
func AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := utils.GetSessionUser(w, r)
		if username == "" {
			log.Println("Unauthorized access attempt - no session")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if user has admin role
		user, err := services.GetUserByUsername(username)
		roleStr := "unknown"
		if user != nil {
			roleStr = user.Role
		}
		if err != nil || roleStr != "admin" {
			log.Printf("Access denied for %s (Role: %s) - Admin only", username, roleStr)
			// Redirect non-admins back to dashboard with an error message instead of returning 403
			http.Redirect(w, r, "/dashboard?error=Admin+privileges+required", http.StatusSeeOther)
			return
		}

		log.Printf("Admin access granted for: %s", username)
		next(w, r)
	}
}

// AdminDashboardHandler - Shows admin panel
func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	adminUsername := utils.GetSessionUser(w, r)
	if adminUsername == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := services.GetUserByUsername(username)
	if err != nil || user == nil || user.Role != "admin" {
		http.Redirect(w, r, "/dashboard?error=Admin+privileges+required", http.StatusSeeOther)
		return
	}

	users, err := services.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		// If DB fails, it's better to show an error than an empty list
		http.Error(w, "Internal Database Error", http.StatusInternalServerError)
		return
	}

	// Adding more functions to the template map makes the dashboard "fancier"
	tmpl, err := template.New("admin.html").Funcs(template.FuncMap{
		"formatKES":  utils.FormatKES,
		"formatDate": utils.FormatDate, // Essential for admin auditing
	}).ParseFiles("web/templates/admin.html")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Display Error", http.StatusInternalServerError)
		return
	}

	// Fetch recent audit logs
	auditLogs, _ := services.GetAuditLogs()

	data := struct {
		AdminUsername string
		Users         []models.User
		AuditLogs     []models.AuditLog
	}{
		AdminUsername: username,
		Users:         users,
		AuditLogs:     auditLogs,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Execution error: %v", err)
	}
}

// AdminToggleAccount - API endpoint to activate/deactivate account
func AdminToggleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	adminUsername := utils.GetSessionUser(w, r)
	if adminUsername == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	accountNumber := r.FormValue("account_number")
	if accountNumber == "" {
		http.Redirect(w, r, "/admin?error=Account+number+required", http.StatusSeeOther)
		return
	}

	// Prevent admin from blocking their own account
	sessionUsername := utils.GetSessionUser(w, r)
	acc, err := services.GetAccountByNumber(accountNumber)
	if err != nil || acc == nil {
		http.Redirect(w, r, "/admin?error=Account+not+found", http.StatusSeeOther)
		return
	}

	if acc.User.Username == sessionUsername {
		http.Redirect(w, r, "/admin?error=You+cannot+block+your+own+account", http.StatusSeeOther)
		return
	}

	account := *acc
	newActive := !account.Active
	if err := services.ToggleAccountStatus(adminUsername, account.ID, newActive); err != nil {
		http.Redirect(w, r, "/admin?error=Failed+to+toggle+status", http.StatusSeeOther)
		return
	}

	action := "activated"
	if !newActive {
		action = "deactivated"
	}
	http.Redirect(w, r, "/admin?success=Account+"+action+"+successfully", http.StatusSeeOther)
}

// AdminDepositHandler - Admin deposit to any account
func AdminDepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	adminUsername := utils.GetSessionUser(w, r)
	if adminUsername == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")

	if accountNumber == "" {
		http.Redirect(w, r, "/admin?error=Account+number+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	err = services.AdminDeposit(adminUsername, accountNumber, amount)
	if err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/admin?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=Deposit+successful!+KES+"+amountStr+"+to+account+"+accountNumber, http.StatusSeeOther)
}

// AdminWithdrawHandler - Admin withdrawal from any account
func AdminWithdrawHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	adminUsername := utils.GetSessionUser(w, r)
	if adminUsername == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")

	if accountNumber == "" {
		http.Redirect(w, r, "/admin?error=Account+number+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	err = services.AdminWithdraw(adminUsername, accountNumber, amount)
	if err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/admin?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=Withdrawal+successful!+KES+"+amountStr+"+from+account+"+accountNumber, http.StatusSeeOther)
}

func AuditLogHandler(w http.ResponseWriter, r *http.Request) {
	logs, err := services.GetAuditLogs()
	if err != nil {
		http.Redirect(w, r, "/admin?error=Failed+to+fetch+audit+logs", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("web/templates/admin.html")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	adminUsername := utils.GetSessionUser(w, r)

	users, _ := services.GetAllUsers()
	adminUsername = utils.GetSessionUser(w, r)

	data := struct {
		AdminUsername string
		Users         []models.User
		AuditLogs     []models.AuditLog
	}{
		AdminUsername: adminUsername,
		Users:         users,
		AuditLogs:     logs,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
