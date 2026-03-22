package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var accounts = make(map[string]Account)
var transactions = make(map[string][]Transaction)

const MinDeposit = 50
const MinWithdrawal = 100

func main() {
	http.HandleFunc("/account", CreateAccount)
	http.HandleFunc("/deposit", Deposits)
	http.HandleFunc("/withdraw", Withdrawals)
	http.HandleFunc("/balance", Balances)
	http.HandleFunc("/transactions", Transactions)
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

	_, exists := accounts[req.Username]
	if exists {
		http.Error(w, "Account already exists", http.StatusConflict) //409
		return
	}

	req.Balance = 0
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
	if req.Amount < MinDeposit {
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

	history:=Transaction{
		Username:req.Username,
		Type:"Deposit",
		Amount:req.Amount,
		Balance:account.Balance,
		Time:time.Now().UTC(),
	}

	transactions[req.Username]=append(transactions[req.Username],history)

	fmt.Println("Deposited Ksh.:", req.Amount, "to", req.Username)
	fmt.Println("The New Balance is: Ksh.", account.Balance)

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

	if req.Amount < MinWithdrawal {
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

	history:=Transaction{
		Username:req.Username,
		Type:"Withdrawal",
		Amount:req.Amount,
		Balance:account.Balance,
		Time:time.Now().UTC(),
	}

	transactions[req.Username]=append(transactions[req.Username],history)

	
	fmt.Println("Withdrew: Ksh.", req.Amount, "from", req.Username)
	fmt.Println("The New Balance is Ksh.:", account.Balance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func Balances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	account, exists := accounts[username]

	if !exists {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	fmt.Println("Account", username, "balance: Ksh.", account.Balance)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)

}

func Transactions(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	_, exists := accounts[username]
	if !exists {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	history := transactions[username]

	if history == nil {
		history = []Transaction{}
	}

	fmt.Println("Username", username)
	fmt.Println("Transactions", history)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}


