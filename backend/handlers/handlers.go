package handlers

import (
	"encoding/json"
	"net/http"
	"fintech-labs/models"
	"fintech-labs/services"
	"fintech-labs/validator"
	"strconv"
	"strings"
)

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}

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
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

func Deposits(w http.ResponseWriter, r *http.Request) {
    var username string
    var amount int64

    // Check authentication from cookie
    cookie, err := r.Cookie("session_user")
    if err == nil && cookie.Value != "" {
        username = cookie.Value
    }

    // Handle HTML Form (from your dashboard)
    if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing form", http.StatusBadRequest)
            return
        }
        
        // If username not from cookie, get from form
        if username == "" {
            username = r.FormValue("username")
        }
        
        // Parse amount
        val, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
        if err != nil {
            http.Error(w, "Invalid amount format", http.StatusBadRequest)
            return
        }
        amount = val

    } else {
        // Handle JSON (from Postman or automated tests)
        var req models.Deposit
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        username = req.Username
        amount = req.Amount
    }

    if username == "" {
        http.Error(w, "Username required", http.StatusBadRequest)
        return
    }

    // Process the deposit
    err = services.DepositProcess(username, amount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Success - Redirect back to dashboard
    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func Withdrawals(w http.ResponseWriter, r *http.Request) {
    var username string
    var amount int64

    // Check authentication from cookie
    cookie, err := r.Cookie("session_user")
    if err == nil && cookie.Value != "" {
        username = cookie.Value
    }

    // Handle HTML Form (from the Dashboard UI)
    if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing form", http.StatusBadRequest)
            return
        }
        
        // If username not from cookie, get from form
        if username == "" {
            username = r.FormValue("username")
        }
        
        // Parse amount
        val, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
        if err != nil {
            http.Error(w, "Invalid amount format", http.StatusBadRequest)
            return
        }
        amount = val

    } else {
        // Handle JSON (from Postman or external API calls)
        var req models.Withdrawal
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON body", http.StatusBadRequest)
            return
        }
        username = req.Username
        amount = req.Amount
    }

    if username == "" {
        http.Error(w, "Username required", http.StatusBadRequest)
        return
    }

    // Validation: Ensure we aren't trying to withdraw 0 or negative
    if amount <= 0 {
        http.Error(w, "Withdrawal amount must be greater than zero", http.StatusBadRequest)
        return
    }

    // Call the service logic
    err = services.WithdrawalProcess(username, amount)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Success - Redirect back to the dashboard
    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func Transactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get username from cookie first, then query param
	username := ""
	cookie, err := r.Cookie("session_user")
	if err == nil && cookie.Value != "" {
		username = cookie.Value
	} else {
		username = r.URL.Query().Get("username")
	}
	
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

func GetAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	names := services.GetAccountsProcess()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func Balances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get username from cookie first, then query param
	username := ""
	cookie, err := r.Cookie("session_user")
	if err == nil && cookie.Value != "" {
		username = cookie.Value
	} else {
		username = r.URL.Query().Get("username")
	}
	
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

func Deactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE is allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if !validator.ValidUsername(username) {
		http.Error(w, "A valid username is required", http.StatusBadRequest)
		return
	}

	err := services.DeactivateAccountProcess(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account " + username + " deactivated successfully."})
}

func Reactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if !validator.ValidUsername(username) {
		http.Error(w, "A valid username is required", http.StatusBadRequest)
		return
	}

	err := services.ReactivateAccountProcess(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account " + username + " reactivated successfully."})
}