package handlers

import (
	"fintech-labs/backend/db"
	"fintech-labs/backend/models"
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
func ChangePinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentPin := r.FormValue("current_pin")
	newPin := r.FormValue("new_pin")
	confirmPin := r.FormValue("confirm_new_pin")

	if newPin != confirmPin {
		http.Redirect(w, r, "/profile?error=New+PINs+do+not+match", http.StatusSeeOther)
		return
	}

	if err := services.ChangeTransactionPin(username, currentPin, newPin); err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/profile?error="+errorMsg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?success=Transaction+PIN+changed+successfully", http.StatusSeeOther)
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_new_password")

	if newPassword != confirmPassword {
		http.Redirect(w, r, "/profile?error=New+passwords+do+not+match", http.StatusSeeOther)
		return
	}

	if err := services.ChangePassword(username, currentPassword, newPassword); err != nil {
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/profile?error="+errorMsg, http.StatusSeeOther)
		return
	}

	// Log out after password change for security
	cookie, err := r.Cookie("session_user")
	if err == nil {
		db.DB.Where("token = ?", cookie.Value).Delete(&models.Session{})
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/login?success=Password+changed!+Please+login+again", http.StatusSeeOther)
}
