package handlers

import (
	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"
	"log"
	"net/http"
	"strings"

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

		// 3. Authenticate by username, email, or phone
		user, err := services.AuthenticateUser(username, password)
		if err != nil {
			log.Printf("Login Fail for identifier '%s': %v", username, err)
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
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			// Preserve other fields when redirecting back so user doesn't retype them
			redirectURL := "/register-page?error=Passwords+do+not+match"
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		// PASS ALL 7 ARGUMENTS TO THE SERVICE
		user, err := services.CreateUser(fullname, username, email, phone, idNumber, password, "customer")
		if err != nil {
			// Preserve fields (except passwords) so user only fixes the invalid fields
			errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
			redirectURL := "/register-page?error=" + errorMsg
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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

// AdminRegister handles GET/POST for creating admin users. If there are no admins
// in the system, the page is open to create the first admin. Otherwise only an
// existing admin can create other admins.
func AdminRegister(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GET: show form
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "frontend/templates/register_admin.html")
			return
		}

		// POST: process form
		fullname := r.FormValue("fullname")
		username := r.FormValue("username")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		idNumber := r.FormValue("id_number")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			redirectURL := "/register-admin?error=Passwords+do+not+match"
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		// Allow creation if no admin exists OR session user is admin
		hasAdmin, err := services.HasAdmin()
		if err != nil {
			http.Redirect(w, r, "/register-admin?error=Server+error", http.StatusSeeOther)
			return
		}

		if hasAdmin {
			// Must be logged-in admin to create another admin
			sessionUser := utils.GetSessionUser(r)
			if sessionUser == "" {
				http.Redirect(w, r, "/login?error=Please+login+as+admin", http.StatusSeeOther)
				return
			}
			u, err := services.GetUserByUsername(sessionUser)
			if err != nil || u == nil || u.Role != "admin" {
				http.Redirect(w, r, "/dashboard?error=Admin+privileges+required", http.StatusSeeOther)
				return
			}
		}

		user, err := services.CreateUser(fullname, username, email, phone, idNumber, password, "admin")
		if err != nil {
			errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
			redirectURL := "/register-admin?error=" + errorMsg
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		_, err = services.CreateAccountForUser(user.ID)
		if err != nil {
			http.Redirect(w, r, "/register-admin?error=Failed+to+create+bank+account", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?success=Admin+account+created!+Please+login", http.StatusSeeOther)
	}
}

func setSessionUser(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
	})
}
