package handlers

import (
	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := services.GetUserByUsername(username)
	if err != nil || user == nil {
		http.Redirect(w, r, "/dashboard?error=User+not+found", http.StatusSeeOther)
		return
	}

	account, err := services.GetAccountByUsername(username)
	if err != nil || account == nil {
		http.Redirect(w, r, "/dashboard?error=Account+not+found", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("frontend/templates/profile.html")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Display Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username      string
		FullName      string
		Email         string
		PhoneNumber   string
		AccountNumber string
		Role          string
		Active        bool
		CreatedAt     string
	}{
		Username:      user.Username,
		FullName:      user.FullName,
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		AccountNumber: account.Number,
		Role:          user.Role,
		Active:        account.Active,
		CreatedAt:     utils.FormatDate(user.CreatedAt),
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Execution error: %v", err)
	}
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	phone := r.FormValue("phone")
	currentPassword := r.FormValue("current_password")

	if currentPassword == "" {
		http.Redirect(w, r, "/profile?error=Current+password+is+required", http.StatusSeeOther)
		return
	}

	err := services.UpdateUserProfile(username, email, phone, currentPassword)
	if err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/profile?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?success=Profile+updated+successfully", http.StatusSeeOther)
}