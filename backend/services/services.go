package services

import (
	"errors"
	"log"
	"strings"

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

func WithdrawalProcess(username string, amount int64) error {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	var account models.Account

	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if err := db.DB.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
		return errors.New("account not found")
	}

	if !account.Active {
		return errors.New("account is inactive")
	}
	if account.Balance < amount {
		return errors.New("insufficient funds")
	}
	if amount < MinWithdrawal || amount > MaxWithdrawal {
		return errors.New("withdrawal amount out of range")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		account.Balance -= amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username: username,
			Type:     "Withdrawal",
			Amount:   amount,
			Balance:  account.Balance,
		}
		return tx.Create(&transaction).Error
	})
}

func DepositProcess(username string, amount int64) error {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	var account models.Account

	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if err := db.DB.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
		return errors.New("account not found")
	}

	if !account.Active {
		return errors.New("account is inactive")
	}
	if amount < MinDeposit || amount > MaxDeposit {
		return errors.New("deposit amount out of range")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		account.Balance += amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username: username,
			Type:     "Deposit",
			Amount:   amount,
			Balance:  account.Balance,
		}
		return tx.Create(&transaction).Error
	})
}

func GetTransactions(username string) ([]models.Transaction, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var history []models.Transaction

	if err := db.DB.Where("username = ?", username).Order("created_at desc").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func CreateAccountProcess(username string) (models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return models.Account{}, errors.New("user not found")
	}

	var account models.Account
	err := db.DB.Where("user_id = ?", user.ID).First(&account).Error
	if err == nil {
		return models.Account{}, errors.New("account already exists")
	}

	account = models.Account{
		UserID:  user.ID,
		Balance: 0,
		Active:  true,
	}

	if err := db.DB.Create(&account).Error; err != nil {
		return models.Account{}, err
	}
	return account, nil
}

func GetAccountsProcess() []string {
	var names []string
	err := db.DB.Model(&models.User{}).Pluck("username", &names).Error
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
		return []string{}
	}
	return names
}

func DeactivateAccountProcess(username string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	result := db.DB.Model(&models.Account{}).Where("user_id = ?", user.ID).Update("active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("account not found")
	}
	return nil
}

func GetBalanceProcess(username string) (models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var user models.User
	var account models.Account

	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return models.Account{}, errors.New("user not found")
	}

	if err := db.DB.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
		return models.Account{}, errors.New("account not found")
	}
	return account, nil
}

func ReactivateAccountProcess(username string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	result := db.DB.Model(&models.Account{}).Where("user_id = ?", user.ID).Update("active", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("account not found")
	}
	return nil
}
