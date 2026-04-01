package handlers

import (
	"encoding/json"
	"fintech-labs/models"
	"fintech-labs/validator"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func generateAccountNumber() string {
	return fmt.Sprintf("ACC%06d", rand.Intn(900000)+100000)
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
			http.Error(w, "Invalid Username: Must be 3-20 characters (letters/numbers only)", http.StatusBadRequest)
			return
		}

		if len(input.Password) < 6 {
			http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error creating password", http.StatusInternalServerError)
			return
		}

		newUser := models.User{
			Username: input.Username,
			Password: string(hashedPassword),
			Role:     input.Role,
		}

		if err := db.Create(&newUser).Error; err != nil {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}

		if newUser.Role == "customer" {
			newAccount := models.Account{
				UserID:  newUser.ID,
				Number:  generateAccountNumber(),
				Balance: 0,
				Active:  true,
			}
			if err := db.Create(&newAccount).Error; err != nil {
				log.Printf("Failed to create account for user %d: %v", newUser.ID, err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":        "Account Successfully Created",
				"account_number": newAccount.Number,
				"username":       newUser.Username,
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Admin user registered"})
	}
}

func LoginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle GET request - show login page
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "templates/login.html")
			return
		}

		// Handle POST request - process login
		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "" || password == "" {
				http.Error(w, "Username and password are required", http.StatusBadRequest)
				return
			}

			// Find user in database
			var user models.User
			result := db.Where("username = ?", strings.ToLower(username)).First(&user)
			if result.Error != nil {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			// Verify password
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			// Set session cookie
			cookie := &http.Cookie{
				Name:     "session_user",
				Value:    user.Username,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, cookie)

			// Redirect to dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_user")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}