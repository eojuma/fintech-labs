package handlers

import (
	"encoding/json"
	"fintech-labs/services"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"golang.org/x/crypto/bcrypt"
)

func Deposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	amountStr := r.FormValue("amount")
	if amountStr == "" {
		http.Error(w, "Amount required", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	log.Printf("💰 Processing deposit for %s: KES %d", username, amount)

	err = services.Deposit(username, amount)
	if err != nil {
		log.Printf("Deposit error for %s: %v", username, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Deposit successful for %s: KES %d", username, amount)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	amountStr := r.FormValue("amount")
	if amountStr == "" {
		http.Error(w, "Amount required", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	log.Printf("💸 Processing withdrawal for %s: KES %d", username, amount)

	err = services.Withdraw(username, amount)
	if err != nil {
		log.Printf("Withdrawal error for %s: %v", username, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Withdrawal successful for %s: KES %d", username, amount)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	account, err := services.GetAccountByUsername(username)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"balance":   account.Balance,
		"formatted": fmt.Sprintf("KES %d", account.Balance),
		"number":    account.Number,
		"active":    account.Active,
	})
}

func GetTransactionsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	transactions, err := services.GetTransactions(username)
	if err != nil {
		log.Printf("Error fetching transactions for %s: %v", username, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("📊 Returning %d transactions for %s", len(transactions), username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}


func SendMoneyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	toAccountNumber := r.FormValue("to_account")
	amountStr := r.FormValue("amount")
	password := r.FormValue("password")

	if toAccountNumber == "" {
		http.Error(w, "Recipient account number required", http.StatusBadRequest)
		return
	}

	if amountStr == "" {
		http.Error(w, "Amount required", http.StatusBadRequest)
		return
	}

	if password == "" {
		http.Error(w, "Password required to confirm transaction", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Verify password before proceeding
	user, err := services.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	log.Printf("💸 Processing transfer from %s to account %s: KES %d", username, toAccountNumber, amount)

	err = services.SendMoney(username, toAccountNumber, amount)
	if err != nil {
		log.Printf("Transfer error from %s to %s: %v", username, toAccountNumber, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Transfer successful from %s to account %s: KES %d", username, toAccountNumber, amount)
	
	// Set success message in session or redirect with query param
	http.Redirect(w, r, "/dashboard?success=transfer", http.StatusSeeOther)
}

