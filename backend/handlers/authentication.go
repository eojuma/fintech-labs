package handlers

import (
	"log"
	"net/http"
	"strings"
	"fintech-labs/backend/utils"
	"fintech-labs/backend/services"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if utils.GetSessionUser(r) == "" {
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
		// 1. Handle the GET request to show the login page
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "frontend/templates/login.html")
			return
		}

		// 2. Capture form values
		username := r.FormValue("username")
		password := r.FormValue("password")

		// 3. Find the user (services.GetUserByUsername already handles Trim and ToLower)
		user, err := services.GetUserByUsername(username)
		if err != nil {
			// DEBUG LOG: Tells you if the user actually exists in the DB
			log.Printf("Login Fail: User '%s' not found in database", username)
			http.Redirect(w, r, "/login?error=Invalid+username+or+password", http.StatusSeeOther)
			return
		}

		// 4. Compare the provided plain-text password with the stored HASH
		// bcrypt.CompareHashAndPassword returns nil on success
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			// DEBUG LOG: Tells you if the password was typed incorrectly or wasn't hashed
			log.Printf("Login Fail: Password mismatch for user '%s'", username)
			http.Redirect(w, r, "/login?error=Invalid+username+or+password", http.StatusSeeOther)
			return
		}

		// 5. Success! Set the session and redirect based on role
		log.Printf("✅ Login Success: %s logged in as %s", user.Username, user.Role)
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

		// CAPTURE ALL NEW FIELDS FROM HTML
		fullname := r.FormValue("fullname")
		username := r.FormValue("username")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		idNumber := r.FormValue("id_number")
		password := r.FormValue("password")

		// PASS ALL 7 ARGUMENTS TO THE SERVICE
		user, err := services.CreateUser(fullname, username, email, phone, idNumber, password, "customer")
		if err != nil {
			// If validation fails (like the invalid email error), send the specific error back
			errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
			http.Redirect(w, r, "/register-page?error="+errorMsg, http.StatusSeeOther)
			return
		}

		_, err = services.CreateAccountForUser(user.ID)
		if err != nil {
			http.Redirect(w, r, "/register-page?error=Failed+to+create+bank+account", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?success=Account+created!+Please+login", http.StatusSeeOther)
	}
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/templates/register.html")
}

func setSessionUser(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
	})
}
