package handlers

import (
	"encoding/json"
	"net/http"

	"fintech-labs/models"
	"fintech-labs/services"
	"fintech-labs/validator"
)

func Deposits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.Deposit

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = services.DepositProcess(req.Username, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message":"Deposit successful."})
}

func Withdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.Withdrawal
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = services.WithdrawalProcess(req.Username, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message":"Withdrawal successful."})
}

func Transactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")

	if !validator.ValidUsername(username) {
		http.Error(w, "A valid username is required", http.StatusBadRequest)
		return
	}
	history, err := services.GetTransactions(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.Account
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !validator.ValidUsername(req.Username) {
		http.Error(w, "A valid username is required", http.StatusBadRequest)
		return
	}

	account, err := services.CreateAccountProcess(req.Username)
	if err != nil {
		if err.Error() == "account already exists" {
			http.Error(w, err.Error(), http.StatusConflict) // 409
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func GetAccounts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
        return
    }

    names := services.GetAccountsProcess()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(names)
}

func Delete(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Only DELETE is allowed", http.StatusMethodNotAllowed)
        return
    }

    username := r.URL.Query().Get("username")
    if !validator.ValidUsername(username) {
        http.Error(w, "A valid username is required", http.StatusBadRequest)
        return
    }

    err := services.DeleteAccountProcess(username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusNoContent) // 204
}

func Balances(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
        return
    }

    username := r.URL.Query().Get("username")

    if !validator.ValidUsername(username) {
        http.Error(w, "A valid username is required", http.StatusBadRequest)
        return
    }

    account, err := services.GetBalanceProcess(username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(account)
}