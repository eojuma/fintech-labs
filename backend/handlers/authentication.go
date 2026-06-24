package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fintech-labs/backend/db"
	"fintech-labs/backend/models"
	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"

	"gorm.io/gorm"
)

func isProduction() bool {
	return os.Getenv("RENDER") == "true"
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if utils.GetSessionUser(w, r) == "" {
			http.Redirect(w, r, "/login?error=Please+login+first", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func Login(gormDB *gorm.DB) http.HandlerFunc {
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
			errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
			http.Redirect(w, r, "/login?error="+errorMsg, http.StatusSeeOther)
			return
		}

		// 5. Success! Set the session and redirect based on role
		log.Printf("✅ Login Success: %s logged in as %s", user.Username, user.Role)
		// Pass the userID so we can link the session to the correct user
		if err := setSessionUser(w, user.ID); err != nil {
			http.Redirect(w, r, "/login?error=Failed+to+create+session", http.StatusSeeOther)
			return
		}
		if user.Role == "admin" {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

func Register(gormDB *gorm.DB) http.HandlerFunc {
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
		transactionPin := r.FormValue("transaction_pin")
		confirmPin := r.FormValue("confirm_transaction_pin")

		if password != confirmPassword {
			// Preserve other fields when redirecting back so user doesn't retype them
			redirectURL := "/register-page?error=Passwords+do+not+match"
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		// Validate PIN
		if len(transactionPin) != 4 {
			redirectURL := "/register-page?error=PIN+must+be+exactly+4+digits"
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		if transactionPin != confirmPin {
			redirectURL := "/register-page?error=PINs+do+not+match"
			redirectURL += "&fullname=" + fullname + "&username=" + username + "&email=" + email + "&phone=" + phone + "&id_number=" + idNumber
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		// PASS ALL 8 ARGUMENTS TO THE SERVICE
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

		// Save the transaction PIN
		if err := services.SetTransactionPin(user.Username, transactionPin); err != nil {
			http.Redirect(w, r, "/register-page?error=Failed+to+set+transaction+PIN", http.StatusSeeOther)
			return
		}
		
		http.Redirect(w, r, "/login?success=Account+created!+Please+login", http.StatusSeeOther)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Get the token from the cookie and delete the session from the database
	cookie, err := r.Cookie("session_user")
	if err == nil {
		db.DB.Where("token = ?", cookie.Value).Delete(&models.Session{})
	}
	// Clear the cookie on the browser

	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/login?success=Logged+out+successfully", http.StatusSeeOther)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/templates/register.html")
}

// AdminRegister handles GET/POST for creating admin users. If there are no admins
// in the system, the page is open to create the first admin. Otherwise only an
// existing admin can create other admins.
func AdminRegister(gormDB *gorm.DB) http.HandlerFunc {
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
			sessionUser := utils.GetSessionUser(w, r)
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

func setSessionUser(w http.ResponseWriter, userID uint) error {
	// Generate a random 32-byte token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	token := hex.EncodeToString(bytes)

	// Save the session to the database
	session := models.Session{
		UserID:         userID,
		Token:          token,
		LastActivityAt: time.Now(),
	}
	if err := db.DB.Create(&session).Error; err != nil {
		return err
	}

	// Store only the token in the cookie — never the username
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   600, // 10 minutes in seconds
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
	})
	return nil
}

func RefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}
