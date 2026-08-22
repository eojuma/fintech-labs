package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/models"
	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func Deposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")
	pin := r.FormValue("pin")

	if accountNumber == "" {
		http.Redirect(w, r, "/dashboard?error=Account+is+required", http.StatusSeeOther)
		return
	}

	if amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=Amount+required", http.StatusSeeOther)
		return
	}

	if pin == "" {
		http.Redirect(w, r, "/dashboard?error=Transaction+PIN+required", http.StatusSeeOther)
		return
	}

	if err := services.VerifyTransactionPin(username, pin); err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	log.Printf("💰 Processing deposit for %s to account %s: KES %d", username, accountNumber, amount)

	refNum, err := services.Deposit(username, accountNumber, amount)
	if err != nil {
		log.Printf("Deposit error for %s: %v", username, err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}
	log.Printf("Deposit successful for %s: KES %d", username, amount)
	http.Redirect(w, r, "/receipt/"+refNum, http.StatusSeeOther)
}

func Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")
	pin := r.FormValue("pin")

	if accountNumber == "" {
		http.Redirect(w, r, "/dashboard?error=Account+is+required", http.StatusSeeOther)
		return
	}

	if amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=Amount+required", http.StatusSeeOther)
		return
	}

	if pin == "" {
		http.Redirect(w, r, "/dashboard?error=Transaction+PIN+required", http.StatusSeeOther)
		return
	}

	if err := services.VerifyTransactionPin(username, pin); err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	log.Printf("💸 Processing withdrawal for %s from account %s: KES %d", username, accountNumber, amount)

	refNum, err := services.Withdraw(username, accountNumber, amount)
	if err != nil {
		log.Printf("Withdrawal error for %s: %v", username, err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	log.Printf("Withdrawal successful for %s: KES %d", username, amount)
	http.Redirect(w, r, "/receipt/"+refNum, http.StatusSeeOther)
}

func SendMoneyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Failed+to+parse+form", http.StatusSeeOther)
		return
	}

	fromAccountNumber := r.FormValue("account_number")
	toAccountNumber := r.FormValue("to_account")
	amountStr := r.FormValue("amount")
	pin := r.FormValue("pin")

	if fromAccountNumber == "" {
		http.Redirect(w, r, "/dashboard?error=Source+account+required", http.StatusSeeOther)
		return
	}

	if toAccountNumber == "" {
		http.Redirect(w, r, "/dashboard?error=Recipient+is+required", http.StatusSeeOther)
		return
	}
	if amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=Amount+required", http.StatusSeeOther)
		return
	}

	if pin == "" {
		http.Redirect(w, r, "/dashboard?error=Transaction+PIN+required", http.StatusSeeOther)
		return
	}

	if err := services.VerifyTransactionPin(username, pin); err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	log.Printf("💸 Processing transfer from %s (account %s) to account %s: KES %d", username, fromAccountNumber, toAccountNumber, amount)

	refNum, err := services.SendMoney(username, fromAccountNumber, toAccountNumber, amount)
	if err != nil {
		log.Printf("Transfer error from %s to %s: %v", username, toAccountNumber, err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	log.Printf("Transfer successful from %s to account %s: KES %d", username, toAccountNumber, amount)
	http.Redirect(w, r, "/receipt/"+refNum, http.StatusSeeOther)
}

func FilterTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read filter params from query string
	q := r.URL.Query()

	minAmount, _ := strconv.ParseInt(q.Get("min_amount"), 10, 64)
	maxAmount, _ := strconv.ParseInt(q.Get("max_amount"), 10, 64)
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := models.TransactionFilter{
		AccountNumber: q.Get("account"),
		Type:          q.Get("type"),
		From:          q.Get("from"),
		To:            q.Get("to"),
		MinAmount:     minAmount,
		MaxAmount:     maxAmount,
		SortOrder:     q.Get("sort"),
		Page:          page,
		Limit:         limit,
	}

	result, err := services.FilterTransactions(username, filter)
	if err != nil {
		log.Printf("Filter error for %s: %v", username, err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
