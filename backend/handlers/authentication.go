package handlers

import (
	"net/http"
	"time"

	"fintech-labs/services"
	"fintech-labs/validator"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthMiddleware protects routes from unauthenticated users
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getSessionUser(r) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// Logout clears the session cookie
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
			http.Redirect(w, r, "/login?error=Invalid+credentials", http.StatusSeeOther)
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

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../frontend/templates/register.html")
}

func Register(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if !validator.ValidUsername(username) {
			http.Redirect(w, r, "/register-page?error=Invalid+username", http.StatusSeeOther)
			return
		}

		// 1. Create the User
		user, err := services.CreateUser(username, password, "customer")
		if err != nil {
			http.Redirect(w, r, "/register-page?error=Registration+failed", http.StatusSeeOther)
			return
		}

		// 2. NEW: Create the Account for the user so they can actually use the bank
		_, err = services.CreateAccountForUser(user.ID)
		if err != nil {
			http.Redirect(w, r, "/register-page?error=Failed+to+initialize+account", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?success=Account+created", http.StatusSeeOther)
	}
}

func setSessionUser(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}