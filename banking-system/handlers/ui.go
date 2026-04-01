package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"fintech-labs/models"
	"fintech-labs/services"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DashboardHandler called")

	username := getSessionUser(r)
	if username == "" {
		log.Println("No username in session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Printf("👤 Loading dashboard for user: %s", username)

	account, err := services.GetAccountByUsername(username)
	if err != nil {
		log.Printf("Error fetching account for %s: %v", username, err)
		http.Error(w, "Account not found. Please contact support.", http.StatusNotFound)
		return
	}

	log.Printf("Account found - Balance: KES %d, Number: %s, Active: %v",
		account.Balance, account.Number, account.Active)

	transactions, err := services.GetTransactions(username)
	if err != nil {
		log.Printf("Error fetching transactions: %v", err)
		transactions = []models.Transaction{}
	}

	log.Printf("Found %d transactions for %s", len(transactions), username)

	// Create template with custom functions
	tmpl := template.New("dashboard.html").Funcs(template.FuncMap{
		"formatKES":  formatKES,
		"formatDate": formatDate,
	})

	tmpl, err = tmpl.ParseFiles("templates/dashboard.html")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Username     string
		Account      *models.Account
		Transactions []models.Transaction
	}{
		Username:     username,
		Account:      account,
		Transactions: transactions,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func formatKES(amount int64) string {
	if amount < 0 {
		return fmt.Sprintf("-KES %d", -amount)
	}
	return fmt.Sprintf("KES %d", amount)
}

func formatDate(t time.Time) string {
	return t.Format("02 Jan 2006 15:04:05")
}
