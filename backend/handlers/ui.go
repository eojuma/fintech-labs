package handlers

import (
	"fintech-labs/backend/models"
	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"
	"html/template"
	"log"
	"net/http"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	account, err := services.GetAccountByUsername(username)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	transactions, err := services.GetTransactions(username)
	if err != nil {
		transactions = []models.Transaction{}
	}

	user, err := services.GetUserByUsername(username)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Fetch all accounts for the account switcher
	accounts, err := services.GetUserAccounts(username)
	if err != nil {
		accounts = []models.Account{}
	}

	// FIXED: Correct path to find the template from the backend directory
	tmpl, err := template.New("dashboard.html").Funcs(template.FuncMap{
		"formatKES":  utils.FormatKES,
		"formatDate": utils.FormatDate,
	}).ParseFiles("frontend/templates/dashboard.html")

	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username      string
		FullName      string
		Role          string
		AccountNumber string
		Balance       int64
		Transactions  []models.Transaction
		Accounts      []models.Account
	}{
		Username:      username,
		FullName:      user.FullName,
		Role:          user.Role,
		AccountNumber: account.Number,
		Balance:       account.Balance,
		Transactions:  transactions,
		Accounts:      accounts,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
