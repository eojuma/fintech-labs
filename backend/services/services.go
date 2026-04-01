package services

import (
	"errors"
	"log"
	"strings"
	"fmt"
	"math/rand"
	"time"
	"fintech-labs/db"
	"fintech-labs/models"
	"gorm.io/gorm"
)

const (
	MinDeposit    = 50
	MinWithdrawal = 100
	MaxWithdrawal = 40000
	MaxDeposit    = 250000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func WithdrawalProcess(username string, amount int64) error {
	username = strings.ToLower(strings.TrimSpace(username))

	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}
	if amount < MinWithdrawal {
		return fmt.Errorf("minimum withdrawal amount is %d", MinWithdrawal)
	}
	if amount > MaxWithdrawal {
		return fmt.Errorf("maximum withdrawal amount is %d", MaxWithdrawal)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("user not found")
			}
			return err
		}

		var account models.Account
		if err := tx.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("account not found")
			}
			return err
		}

		if !account.Active {
			return errors.New("account is inactive")
		}
		
		if account.Balance < amount {
			return fmt.Errorf("insufficient funds. Balance: %d, Requested: %d", account.Balance, amount)
		}

		account.Balance -= amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username: username,
			Type:     "withdrawal",
			Amount:   amount,
			Balance:  account.Balance,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		log.Printf("Withdrawal: %s withdrew %d, new balance: %d", username, amount, account.Balance)
		return nil
	})
}

func DepositProcess(username string, amount int64) error {
	username = strings.ToLower(strings.TrimSpace(username))

	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}
	if amount < MinDeposit {
		return fmt.Errorf("minimum deposit amount is %d", MinDeposit)
	}
	if amount > MaxDeposit {
		return fmt.Errorf("maximum deposit amount is %d", MaxDeposit)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("user not found")
			}
			return err
		}

		var account models.Account
		if err := tx.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("account not found")
			}
			return err
		}

		if !account.Active {
			return errors.New("account is inactive")
		}

		account.Balance += amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username: username,
			Type:     "deposit",
			Amount:   amount,
			Balance:  account.Balance,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		log.Printf("Deposit: %s deposited %d, new balance: %d", username, amount, account.Balance)
		return nil
	})
}

func GetTransactions(username string) ([]models.Transaction, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	
	var history []models.Transaction
	if err := db.DB.Where("username = ?", username).Order("created_at desc").Find(&history).Error; err != nil {
		return nil, err
	}
	
	return history, nil
}

func CreateAccountProcess(username string) (models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	
	var account models.Account
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		var existingAccount models.Account
		err := tx.Where("user_id = ?", user.ID).First(&existingAccount).Error
		if err == nil {
			return errors.New("account already exists")
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		account = models.Account{
			UserID:  user.ID,
			Number:  fmt.Sprintf("ACC%06d", rand.Intn(900000)+100000),
			Balance: 0,
			Active:  true,
		}

		return tx.Create(&account).Error
	})
	
	return account, err
}

func GetAccountsProcess() []string {
	var names []string
	db.DB.Model(&models.User{}).Where("role = ?", "customer").Pluck("username", &names)
	return names
}

func DeactivateAccountProcess(username string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		result := tx.Model(&models.Account{}).Where("user_id = ?", user.ID).Update("active", false)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("account not found")
		}
		return nil
	})
}

func GetBalanceProcess(username string) (models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	
	var account models.Account
	err := db.DB.Joins("JOIN users ON users.id = accounts.user_id").
		Where("users.username = ?", username).
		Preload("User").
		First(&account).Error

	return account, err
}

func ReactivateAccountProcess(username string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		result := tx.Model(&models.Account{}).Where("user_id = ?", user.ID).Update("active", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("account not found")
		}
		return nil
	})
}