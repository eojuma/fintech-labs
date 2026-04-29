package handlers

import (
	"encoding/json"
	"fintech-labs/backend/models"
	"fintech-labs/backend/services"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// AdminAuthMiddleware checks authentication AND admin role
func AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := getSessionUser(r)
		if username == "" {
			log.Println("Unauthorized access attempt - no session")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if user has admin role
		user, err := services.GetUserByUsername(username)
		if err != nil || user.Role != "admin" {
			log.Printf("Access denied for %s (Role: %v) - Admin only", username)
			http.Error(w, "Access denied. Admin privileges required.", http.StatusForbidden)
			return
		}

		log.Printf("Admin access granted for: %s", username)
		next(w, r)
	}
}

// AdminDashboardHandler - Shows admin panel
func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := getSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verify admin role
	user, err := services.GetUserByUsername(username)
	if err != nil || user.Role != "admin" {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get all users for admin view
	users, err := services.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		users = []models.User{}
	}

	tmpl := template.New("admin.html").Funcs(template.FuncMap{
		"formatKES": formatKES,
	})

	tmpl, err = tmpl.ParseFiles("frontend/templates/admin.html")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Users    []models.User
	}{
		Username: username,
		Users:    users,
	}

	tmpl.Execute(w, data)
}

// AdminToggleAccount - API endpoint to activate/deactivate account
func AdminToggleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		AccountID uint `json:"account_id"`
		Active    bool `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := services.ToggleAccountStatus(input.AccountID, input.Active)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Account %s", map[bool]string{true: "activated", false: "deactivated"}[input.Active]),
	})
}

// AdminDepositHandler - Admin deposit to any account
func AdminDepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	err = services.AdminDeposit(accountNumber, amount)
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

	err = services.AdminWithdraw(accountNumber, amount)
	if err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/admin?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=Withdrawal+successful!+KES+"+amountStr+"+from+account+"+accountNumber, http.StatusSeeOther)
}
