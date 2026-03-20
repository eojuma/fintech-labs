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
	fmt.Println("Server running on http://8080/deposit")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server down", err)
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
	if req.Amount <= 0.00 {
		http.Error(w, "Amount must be greater that zero", http.StatusBadRequest)
		return
	}

	account, exists := accounts[req.Username]

	if !exists {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	account.Balance += req.Amount
	accounts[req.Username] = account

	fmt.Println("Deposited: ", req.Amount, "to", req.Username)
	fmt.Println("New Balance is: ", account.Balance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}
