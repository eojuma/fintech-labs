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

func generateAccount() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("Account: %d", rand.Intn(900000)+100000)
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
			http.Error(w, "Invalid Username: Must be 3-20 characters (letters/numbers only) and not a reserved word.", http.StatusBadRequest)
			return
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

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
				Number:  generateAccount(),
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
