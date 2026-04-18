package handlers

import (
	"fintech-labs/services"
	"fintech-labs/validator"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getSessionUser(r) == "" {
			http.Redirect(w, r, "/login?error=Please+login+first", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login?success=Logged+out+successfully", http.StatusSeeOther)
}

func Login(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "../frontend/templates/login.html")
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := services.GetUserByUsername(username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
			http.Redirect(w, r, "/login?error=Invalid+username+or+password", http.StatusSeeOther)
			return
		}

		setSessionUser(w, user.Username)
		if user.Role == "admin" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

func Register(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if !validator.ValidUsername(username) {
			http.Redirect(w, r, "/register-page?error=Invalid+username+format", http.StatusSeeOther)
			return
		}

		user, err := services.CreateUser(username, password, "customer")
		if err != nil {
			http.Redirect(w, r, "/register-page?error=User+already+exists", http.StatusSeeOther)
			return
		}

		// Create the actual bank account so the user can see their dashboard
		_, err = services.CreateAccountForUser(user.ID)
		if err != nil {
			http.Redirect(w, r, "/register-page?error=Failed+to+create+bank+account", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?success=Account+created!+Please+login", http.StatusSeeOther)
	}
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../frontend/templates/register.html")
}

func setSessionUser(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
	})
}