package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/models"
	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func TellerDashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	users, err := services.GetAllUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("teller_dashboard.html").Funcs(template.FuncMap{
		"formatKES":  utils.FormatKES,
		"formatDate": utils.FormatDate,
	}).ParseFiles("web/templates/teller_dashboard.html")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		TellerUsername string
		Users          []models.User
	}{
		TellerUsername: username,
		Users:          users,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func TellerDepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tellerUsername := utils.GetSessionUser(w, r)
	if tellerUsername == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")

	if accountNumber == "" {
		http.Redirect(w, r, "/teller?error=Account+number+required", http.StatusSeeOther)
		return
	}

	if amountStr == "" {
		http.Redirect(w, r, "/teller?error=Amount+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/teller?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	_, err = services.AdminDeposit(tellerUsername, accountNumber, amount)
	if err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/teller?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/teller?success=Deposit+of+KES+"+amountStr+"+successful", http.StatusSeeOther)
}

func TellerWithdrawHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tellerUsername := utils.GetSessionUser(w, r)
	if tellerUsername == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	accountNumber := r.FormValue("account_number")
	amountStr := r.FormValue("amount")

	if accountNumber == "" {
		http.Redirect(w, r, "/teller?error=Account+number+required", http.StatusSeeOther)
		return
	}

	if amountStr == "" {
		http.Redirect(w, r, "/teller?error=Amount+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Redirect(w, r, "/teller?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	_, err = services.AdminWithdraw(tellerUsername, accountNumber, amount)
	if err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/teller?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/teller?success=Withdrawal+of+KES+"+amountStr+"+successful", http.StatusSeeOther)
}
