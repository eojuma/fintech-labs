package handlers

import (
	"html/template"
	"log"
	"net/http"
	"fintech-labs/models"
	"fintech-labs/services"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := getSessionUser(r)
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

	// FIXED: Correct path to find the template from the backend directory
	tmpl, err := template.New("dashboard.html").Funcs(template.FuncMap{
		"formatKES": formatKES,
		"formatDate": formatDate,
	}).ParseFiles("../frontend/templates/dashboard.html")
	
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username     string
		AccountNumber string
		Balance      int64
		Transactions []models.Transaction
	}{
		Username:      username,
		AccountNumber: account.Number,
		Balance:       account.Balance,
		Transactions:  transactions,
	}

	tmpl.Execute(w, data)
}