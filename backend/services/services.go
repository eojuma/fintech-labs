package services

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"fintech-labs/db"
	"fintech-labs/models"

	"gorm.io/gorm"
)

const (
	MinDeposit    = 50
	MaxDeposit    = 250000
	MinWithdrawal = 100
	MaxWithdrawal = 40000

	MinTransfer = 10
	MaxTransfer = 100000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateAccountNumber() string {
	return fmt.Sprintf("%06d", rand.Intn(900000)+100000)
}

func GetUserByUsername(username string) (*models.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var user models.User
	err := db.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetAccountByUserID(userID uint) (*models.Account, error) {
	var account models.Account
	err := db.DB.Where("user_id = ?", userID).Preload("User").First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func GetAccountByUsername(username string) (*models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var account models.Account
	err := db.DB.Joins("JOIN users ON users.id = accounts.user_id").
		Where("users.username = ?", username).
		Preload("User").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func CreateUser(username, password, role string) (*models.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	user := &models.User{
		Username: username,
		Password: password,
		Role:     role,
	}

	err := db.DB.Create(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func CreateAccountForUser(userID uint) (*models.Account, error) {
	account := &models.Account{
		UserID:  userID,
		Number:  GenerateAccountNumber(),
		Balance: 0,
		Active:  true,
	}

	err := db.DB.Create(account).Error
	if err != nil {
		return nil, err
	}

	return account, nil
}

func Deposit(username string, amount int64) error {
	username = strings.ToLower(strings.TrimSpace(username))

	if amount < MinDeposit {
		return fmt.Errorf("minimum deposit is KES %d", MinDeposit)
	}
	if amount > MaxDeposit {
		return fmt.Errorf("maximum deposit is KES %d", MaxDeposit)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		var account models.Account
		if err := tx.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		if !account.Active {
			return errors.New("account is inactive")
		}

		oldBalance := account.Balance
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
			log.Printf("Failed to record transaction: %v", err)
			return err
		}

		log.Printf("💰 Deposit: %s deposited KES %d (Balance: KES %d → KES %d)",
			username, amount, oldBalance, account.Balance)
		return nil
	})
}

func Withdraw(username string, amount int64) error {
	username = strings.ToLower(strings.TrimSpace(username))

	if amount < MinWithdrawal {
		return fmt.Errorf("minimum withdrawal is KES %d", MinWithdrawal)
	}
	if amount > MaxWithdrawal {
		return fmt.Errorf("maximum withdrawal is KES %d", MaxWithdrawal)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		var account models.Account
		if err := tx.Where("user_id = ?", user.ID).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		if !account.Active {
			return errors.New("account is inactive")
		}

		if account.Balance < amount {
			return fmt.Errorf("insufficient funds. Your balance is KES %d", account.Balance)
		}

		oldBalance := account.Balance
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
			log.Printf("Failed to record transaction: %v", err)
			return err
		}

		log.Printf("💸 Withdrawal: %s withdrew KES %d (Balance: KES %d → KES %d)",
			username, amount, oldBalance, account.Balance)
		return nil
	})
}

func GetTransactions(username string) ([]models.Transaction, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	var transactions []models.Transaction
	err := db.DB.Where("username = ?", username).
		Order("created_at desc").
		Limit(50).
		Find(&transactions).Error
	if err != nil {
		log.Printf("Error fetching transactions for %s: %v", username, err)
		return nil, err
	}

	log.Printf("Retrieved %d transactions for %s", len(transactions), username)

	return transactions, nil
}

func GetAllAccounts() ([]models.Account, error) {
	var accounts []models.Account
	err := db.DB.Preload("User").Find(&accounts).Error
	return accounts, err
}

func SendMoney(fromUsername, toAccountNumber string, amount int64) error {
	fromUsername = strings.ToLower(strings.TrimSpace(fromUsername))
	toAccountNumber = strings.TrimSpace(toAccountNumber)

	if amount < MinTransfer {
		return fmt.Errorf("minimum transfer is KES %d", MinTransfer)
	}
	if amount > MaxTransfer {
		return fmt.Errorf("maximum transfer is KES %d", MaxTransfer)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		// Get sender
		var fromUser models.User
		if err := tx.Where("username = ?", fromUsername).First(&fromUser).Error; err != nil {
			return errors.New("sender not found")
		}

		// Get sender's account
		var fromAccount models.Account
		if err := tx.Where("user_id = ?", fromUser.ID).First(&fromAccount).Error; err != nil {
			return errors.New("sender account not found")
		}

		if !fromAccount.Active {
			return errors.New("your account is inactive")
		}

		if fromAccount.Balance < amount {
			return fmt.Errorf("insufficient funds. Your balance is KES %d", fromAccount.Balance)
		}

		// Get receiver's account by account number
		var toAccount models.Account
		if err := tx.Where("number = ?", toAccountNumber).First(&toAccount).Error; err != nil {
			return errors.New("recipient account not found")
		}

		if !toAccount.Active {
			return errors.New("recipient account is inactive")
		}

		// Prevent sending to self
		if fromAccount.ID == toAccount.ID {
			return errors.New("cannot send money to your own account")
		}

		// Get receiver user info for logging
		var toUser models.User
		if err := tx.Where("id = ?", toAccount.UserID).First(&toUser).Error; err != nil {
			return errors.New("recipient user not found")
		}

		// Withdraw from sender
		fromOldBalance := fromAccount.Balance
		fromAccount.Balance -= amount
		if err := tx.Save(&fromAccount).Error; err != nil {
			return err
		}

		// Deposit to receiver
		toOldBalance := toAccount.Balance
		toAccount.Balance += amount
		if err := tx.Save(&toAccount).Error; err != nil {
			return err
		}

		// Record sender transaction
		senderTx := models.Transaction{
			Username: fromUsername,
			Type:     "transfer_out",
			Amount:   amount,
			Balance:  fromAccount.Balance,
		}
		if err := tx.Create(&senderTx).Error; err != nil {
			log.Printf("Failed to record sender transaction: %v", err)
			return err
		}

		// Record receiver transaction
		receiverTx := models.Transaction{
			Username: toUser.Username,
			Type:     "transfer_in",
			Amount:   amount,
			Balance:  toAccount.Balance,
		}
		if err := tx.Create(&receiverTx).Error; err != nil {
			log.Printf("Failed to record receiver transaction: %v", err)
			return err
		}

		log.Printf("💸 Transfer: %s sent KES %d to %s (Account: %s) (Sender: KES %d → KES %d, Recipient: KES %d → KES %d)",
			fromUsername, amount, toUser.Username, toAccountNumber, fromOldBalance, fromAccount.Balance, toOldBalance, toAccount.Balance)

		return nil
	})
}

// GetAllUsers - Fetch all users with their accounts
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := db.DB.Preload("Accounts").Find(&users).Error
	return users, err
}

// GetUserByID - Fetch a single user by ID
func GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := db.DB.Preload("Accounts").First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ToggleAccountStatus - Activate/Deactivate an account
func ToggleAccountStatus(accountID uint, active bool) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.First(&account, accountID).Error; err != nil {
			return errors.New("account not found")
		}

		account.Active = active
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		status := "deactivated"
		if active {
			status = "activated"
		}
		log.Printf("Admin: Account %s (User ID: %d) has been %s", account.Number, account.UserID, status)
		return nil
	})
}

// AdminDeposit - Admin deposit to any user account
func AdminDeposit(accountNumber string, amount int64) error {
	if amount < MinDeposit {
		return fmt.Errorf("minimum deposit is KES %d", MinDeposit)
	}
	if amount > MaxDeposit {
		return fmt.Errorf("maximum deposit is KES %d", MaxDeposit)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.Where("number = ?", accountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		var user models.User
		if err := tx.First(&user, account.UserID).Error; err != nil {
			return errors.New("user not found")
		}

		oldBalance := account.Balance
		account.Balance += amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username: user.Username,
			Type:     "deposit",
			Amount:   amount,
			Balance:  account.Balance,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		log.Printf("💰 ADMIN DEPOSIT: Added KES %d to account %s (User: %s) | Balance: KES %d → KES %d",
			amount, account.Number, user.Username, oldBalance, account.Balance)
		return nil
	})
}

// AdminWithdraw - Admin withdrawal from any user account
func AdminWithdraw(accountNumber string, amount int64) error {
	if amount < MinWithdrawal {
		return fmt.Errorf("minimum withdrawal is KES %d", MinWithdrawal)
	}
	if amount > MaxWithdrawal {
		return fmt.Errorf("maximum withdrawal is KES %d", MaxWithdrawal)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.Where("number = ?", accountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		var user models.User
		if err := tx.First(&user, account.UserID).Error; err != nil {
			return errors.New("user not found")
		}

		if account.Balance < amount {
			return fmt.Errorf("insufficient funds. Account balance is KES %d", account.Balance)
		}

		oldBalance := account.Balance
		account.Balance -= amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username: user.Username,
			Type:     "withdrawal",
			Amount:   amount,
			Balance:  account.Balance,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		log.Printf("💸 ADMIN WITHDRAWAL: Removed KES %d from account %s (User: %s) | Balance: KES %d → KES %d",
			amount, account.Number, user.Username, oldBalance, account.Balance)
		return nil
	})
}

// GetAccountByNumber - Fetch account by account number
func GetAccountByNumber(accountNumber string) (*models.Account, error) {
	var account models.Account
	err := db.DB.Where("number = ?", accountNumber).Preload("User").First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
