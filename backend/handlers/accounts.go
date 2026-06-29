package handlers

import (
	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"
	"net/http"
)

func OpenSavingsAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err := services.CreateSavingsAccount(username)
	if err != nil {
		errorMsg := err.Error()
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard?success=Savings+account+opened+successfully", http.StatusSeeOther)
}