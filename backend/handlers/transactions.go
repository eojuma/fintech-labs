package handlers

import (
	"encoding/json"
	"fintech-labs/backend/services"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func Deposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	amountStr := r.FormValue("amount")
	if amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=Amount+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	log.Printf("💰 Processing deposit for %s: KES %d", username, amount)

	err = services.Deposit(username, amount)
	if err != nil {
		log.Printf("Deposit error for %s: %v", username, err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	log.Printf("Deposit successful for %s: KES %d", username, amount)
	http.Redirect(w, r, "/dashboard?success=Deposit+successful!+KES+"+amountStr, http.StatusSeeOther)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := getSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	amountStr := r.FormValue("amount")
	if amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=Amount+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	log.Printf("💸 Processing withdrawal for %s: KES %d", username, amount)

	err = services.Withdraw(username, amount)
	if err != nil {
		log.Printf("Withdrawal error for %s: %v", username, err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	log.Printf("Withdrawal successful for %s: KES %d", username, amount)
	http.Redirect(w, r, "/dashboard?success=Withdrawal+successful!+KES+"+amountStr, http.StatusSeeOther)
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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Failed+to+parse+form", http.StatusSeeOther)
		return
	}

	toAccountNumber := r.FormValue("to_account")
	amountStr := r.FormValue("amount")
	password := r.FormValue("password")

	if toAccountNumber == "" {
		http.Redirect(w, r, "/dashboard?error=Recipient+account+number+required", http.StatusSeeOther)
		return
	}

	if amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=Amount+required", http.StatusSeeOther)
		return
	}

	if password == "" {
		http.Redirect(w, r, "/dashboard?error=Password+required+to+confirm+transaction", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	// Verify password before proceeding
	user, err := services.GetUserByUsername(username)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=User+not+found", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+password", http.StatusSeeOther)
		return
	}

	log.Printf("💸 Processing transfer from %s to account %s: KES %d", username, toAccountNumber, amount)

	err = services.SendMoney(username, toAccountNumber, amount)
	if err != nil {
		log.Printf("Transfer error from %s to %s: %v", username, toAccountNumber, err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	log.Printf("Transfer successful from %s to account %s: KES %d", username, toAccountNumber, amount)
	http.Redirect(w, r, "/dashboard?success=Transfer+successful!+KES+"+amountStr+"+sent+to+account+"+toAccountNumber, http.StatusSeeOther)
}
