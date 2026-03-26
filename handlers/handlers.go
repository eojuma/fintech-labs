package handlers

import (
	"encoding/json"
	"net/http"

	"fintech-labs/models"
	"fintech-labs/services"
)

func Deposits(w http.ResponseWriter,r *http.Request){
if r.Method !=http.MethodPost{
	http.Error(w,"Only POST is allowed",http.StatusMethodNotAllowed)
	return
}

var req models.Deposit

err:=json.NewDecoder(r.Body).Decode(&req)

if err !=nil{
	http.Error(w,"Invalid request body",http.StatusBadRequest)
	return
}

account,err:=services.DepositProcess(req.Username,req.Amount)

if err !=nil{
	http.Error(w,err.Error(),http.StatusBadRequest)
	return
}
json.NewEncoder(w).Encode(account)
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

	account, err := services.WithdrawalProcess(req.Username, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(account)
}

func Transactions(w http.ResponseWriter,r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w,"Only GET is allowed",http.StatusMethodNotAllowed)
		return 
	}

	username:=r.URL.Query().Get("username")

	if !ValidUsername(username){
		http.Error(w,"A valid username is required",http.StatusBadRequest)
		return
	}
hisotry 
	
}