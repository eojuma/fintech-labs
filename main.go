package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var accounts = make(map[string]Account)

func main() {
	http.HandleFunc("/account", CreateAccount)
	http.HandleFunc("/deposit", Deposits)
	http.HandleFunc("/withdraw",Withdrawals)
	fmt.Println("Server running on http://8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("The server is down", err)
	}
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Account

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if _, exists := accounts[req.Username]; exists {
		http.Error(w, "Account already exists", http.StatusBadRequest)
		return
	}

	req.Balance = 0.00
	accounts[req.Username] = req

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func Deposits(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Deposit

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if req.Amount <=49.00 {
		http.Error(w, "Amount must be greater than Ksh.50", http.StatusBadRequest)
		return
	}

	account, exists := accounts[req.Username]

	if !exists {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	account.Balance += req.Amount
	accounts[req.Username] = account

	fmt.Println("Deposited:", req.Amount, "to", req.Username)
	fmt.Println("The New Balance is:", account.Balance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func Withdrawals(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Withdrawal

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if req.Amount <= 99.00 {
		http.Error(w, "Minimum withdrawal is Ksh.100", http.StatusBadRequest)
		return
	}

	account, exist := accounts[req.Username]

	if !exist {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	if account.Balance < req.Amount {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}
	account.Balance -= req.Amount
	accounts[req.Username] = account

	fmt.Println("Withdrew:", req.Amount, "from", req.Username)
	fmt.Println("The New Balance is:", account.Balance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}
