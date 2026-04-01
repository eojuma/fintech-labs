package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"fintech-labs/services"
	"fintech-labs/validator"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func getSessionUser(r *http.Request) string {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		return ""
	}
	return cookie.Value
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

func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func Register(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		input.Username = strings.ToLower(strings.TrimSpace(input.Username))

		if !validator.ValidUsername(input.Username) {
			http.Error(w, "Invalid username: 3-20 characters, letters/numbers only", http.StatusBadRequest)
			return
		}

		if len(input.Password) < 6 {
			http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
			return
		}

		if input.Role == "" {
			input.Role = "customer"
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error creating password", http.StatusInternalServerError)
			return
		}

		user, err := services.CreateUser(input.Username, string(hashedPassword), input.Role)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				http.Error(w, "Username already taken", http.StatusConflict)
			} else {
				http.Error(w, "Error creating user", http.StatusInternalServerError)
			}
			return
		}

		account, err := services.CreateAccountForUser(user.ID)
		if err != nil {
			log.Printf("Failed to create account for user %d: %v", user.ID, err)
			http.Error(w, "Account creation failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":        "Registration successful! Please login.",
			"account_number": account.Number,
			"username":       user.Username,
		})
	}
}

func Login(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "templates/login.html")
			return
		}

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "" || password == "" {
				http.Error(w, "Username and password required", http.StatusBadRequest)
				return
			}

			user, err := services.GetUserByUsername(username)
			if err != nil {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			_, err = services.GetAccountByUserID(user.ID)
			if err != nil {
				_, err = services.CreateAccountForUser(user.ID)
				if err != nil {
					log.Printf("Failed to create missing account: %v", err)
				}
			}

			setSessionUser(w, user.Username)

			log.Printf("User logged in: %s", username)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/register.html")
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := getSessionUser(r)
		if username == "" {
			log.Println("Unauthorized access attempt - no session")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		log.Printf("Auth middleware passed for user: %s", username)
		next(w, r)
	}
}
