package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"fintech-labs/backend/db"
	"fintech-labs/backend/models"
	"fintech-labs/backend/utils"

	"golang.org/x/crypto/bcrypt"
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

var dummyHash []byte

func init() {
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
}

func CreateUser(fullname, username, email, phone, Id, password, role string) (*models.User, error) {
	cleanfullname := strings.TrimSpace(fullname)
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanUsername := strings.ToLower(strings.TrimSpace(username))
	cleanPhoneNumber := strings.TrimSpace(phone)
	cleanId := strings.TrimSpace(Id)
	if !utils.ValidEmail(cleanEmail) {
		return nil, fmt.Errorf("invalid email address")
	}

	if !utils.ValidFullName(cleanfullname) {
		return nil, fmt.Errorf("invalid full name")
	}
	if !utils.ValidUsername(cleanUsername) {
		return nil, fmt.Errorf("invalid username: must be 3-30 characters and contain only letters, numbers or .-_")
	}
	if strings.HasPrefix(cleanPhoneNumber, "0") {
		cleanPhoneNumber = "254" + cleanPhoneNumber[1:]
	}

	if !utils.ValidPhoneNumber(cleanPhoneNumber) {
		return nil, fmt.Errorf("invalid phone number")
	}

	if !utils.ValidNationalID(cleanId) {
		return nil, fmt.Errorf("invalid National ID Number")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, err
	}

	user := &models.User{
		Email:       cleanEmail,
		FullName:    cleanfullname,
		Username:    cleanUsername,
		Password:    string(hashedPassword),
		NationlID:   cleanId,
		Role:        role,
		PhoneNumber: cleanPhoneNumber,
	}

	result := db.DB.Create(user)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("a user with this email, username, or phone number already exists")
		}
		log.Printf("Error creating user: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("✅ User created successfully: %s (Role: %s)", cleanUsername, role)
	return user, nil
}

func GenerateAccountNumber() (string, error) {
	var count int64
	if err := db.DB.Model(&models.Account{}).Count(&count).Error; err != nil {
		return "", err
	}
	year := time.Now().Year()
	sequence := count + 1
	return fmt.Sprintf("AV%d%08d", year, sequence), nil
}

func CreateAccountForUser(userID uint) (*models.Account, error) {
	num, err := GenerateAccountNumber()
	if err != nil {
		return nil, err
	}

	// Check for collision just in case
	var existing models.Account
	if err := db.DB.Where("number = ?", num).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("account number collision, please try again")
	}

	account := &models.Account{
		UserID:  userID,
		Number:  num,
		Balance: 0,
		Active:  true,
		AccountType: "current",
	}
	if err := db.DB.Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func AuthenticateUser(identifier, password string) (*models.User, error) {
	cleanIdentifier := strings.ToLower(strings.TrimSpace(identifier))

	var user models.User

	err := db.DB.Where("email = ? OR phone_number = ? OR username = ?",
		cleanIdentifier, cleanIdentifier, cleanIdentifier).First(&user).Error
	if err != nil {
		// User not found — run bcrypt anyway to prevent timing attacks
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, errors.New("invalid credentials")
	}

	// Check if account is currently locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remaining := time.Until(*user.LockedUntil).Round(time.Second)
		return nil, fmt.Errorf("account locked. Try again in %v", remaining)
	}
	// Check if account is suspended
	var account models.Account
	if err := db.DB.Where("user_id = ?", user.ID).First(&account).Error; err == nil {
		if !account.Active {
			return nil, errors.New("your account has been suspended. Please contact support")
		}
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		user.FailedLoginAttempts++

		if user.FailedLoginAttempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &lockUntil
			user.FailedLoginAttempts = 0
			db.DB.Save(&user)
			return nil, errors.New("account locked for 15 minutes due to too many failed attempts")
		}

		db.DB.Save(&user)
		attemptsLeft := 5 - user.FailedLoginAttempts
		return nil, fmt.Errorf("invalid credentials. %d attempt(s) remaining before lockout", attemptsLeft)
	}

	// Success — reset everything
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	db.DB.Save(&user)

	return &user, nil
}
func Deposit(accountNumber string, amount int64) error {
	accountNumber = strings.TrimSpace(accountNumber)

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

		if !account.Active {
			return errors.New("account is inactive")
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
			log.Printf("Failed to record transaction: %v", err)
			return err
		}

		log.Printf("💰 Deposit: %s deposited KES %d to account %s (Balance: KES %d → KES %d)",
			user.Username, amount, account.Number, oldBalance, account.Balance)
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

func Withdraw(accountNumber string, amount int64) error {
	accountNumber = strings.TrimSpace(accountNumber)

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

		if !account.Active {
			return errors.New("account is inactive")
		}

		var user models.User
		if err := tx.First(&user, account.UserID).Error; err != nil {
			return errors.New("user not found")
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
			Username: user.Username,
			Type:     "withdrawal",
			Amount:   amount,
			Balance:  account.Balance,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			log.Printf("Failed to record transaction: %v", err)
			return err
		}

		log.Printf("💸 Withdrawal: %s withdrew KES %d from account %s (Balance: KES %d → KES %d)",
			user.Username, amount, account.Number, oldBalance, account.Balance)
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

func SendMoney(fromAccountNumber, toIdentifier string, amount int64) error {
	fromAccountNumber = strings.TrimSpace(fromAccountNumber)
	toIdentifier = strings.TrimSpace(toIdentifier)

	if amount < MinTransfer || amount > MaxTransfer {
		return fmt.Errorf("transfer must be between KES %d and KES %d", MinTransfer, MaxTransfer)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var fromAccount models.Account
		if err := tx.Where("number = ?", fromAccountNumber).First(&fromAccount).Error; err != nil {
			return errors.New("sender account not found")
		}

		if !fromAccount.Active {
			return errors.New("transaction denied: account inactive")
		}

		var fromUser models.User
		if err := tx.First(&fromUser, fromAccount.UserID).Error; err != nil {
			return errors.New("sender not found")
		}

		if fromAccount.Balance < amount {
			return fmt.Errorf("insufficient funds. Your balance is KES %d", fromAccount.Balance)
		}

		var toAccount models.Account
		if err := tx.Where("number = ?", toIdentifier).First(&toAccount).Error; err != nil {
			return errors.New("recipient account not found")
		}

		if !toAccount.Active || fromAccount.ID == toAccount.ID {
			return errors.New("invalid recipient account")
		}

		var toUser models.User
		if err := tx.Where("id = ?", toAccount.UserID).First(&toUser).Error; err != nil {
			return errors.New("recipient user not found")
		}

		fromOldBalance := fromAccount.Balance
		fromAccount.Balance -= amount
		toOldBalance := toAccount.Balance
		toAccount.Balance += amount

		if err := tx.Save(&fromAccount).Error; err != nil {
			return err
		}
		if err := tx.Save(&toAccount).Error; err != nil {
			return err
		}

		tx.Create(&models.Transaction{Username: fromUser.Username, Type: "transfer_out", Amount: amount, Balance: fromAccount.Balance})
		tx.Create(&models.Transaction{Username: toUser.Username, Type: "transfer_in", Amount: amount, Balance: toAccount.Balance})

		log.Printf("💸 Transfer: %s sent KES %d to %s (Acc: %s) | Sender: %d -> %d | Recipient: %d -> %d",
			fromUser.Username, amount, toUser.Username, toAccount.Number, fromOldBalance, fromAccount.Balance, toOldBalance, toAccount.Balance)

		return nil
	})
}

func GetTransactions(Identifier string) ([]models.Transaction, error) {
	Identifier = strings.ToLower(strings.TrimSpace(Identifier))

	var user models.User
	var transactions []models.Transaction
	query := "username = ? OR email = ?"
	if err := db.DB.Where(query, Identifier, Identifier).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	err := db.DB.Where("username = ?", user.Username).
		Order("created_at desc").
		Limit(50).
		Find(&transactions).Error
	if err != nil {
		log.Printf("Error fetching transactions for %s: %v", user.Username, err)
		return nil, err
	}

	log.Printf("Retrieved %d transactions for %s", len(transactions), user.Username)

	return transactions, nil
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

// GetUserByID - Fetch a single user by ID
func GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := db.DB.Preload("Accounts").First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAccountByUserID - Fetch primary account by user id
func GetAccountByUserID(userID uint) (*models.Account, error) {
	var account models.Account
	err := db.DB.Where("user_id = ? AND account_type = ?", userID, "current").
		Preload("User").First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
// GetAccountByUsername - Fetch primary account by username

func GetAccountByUsername(username string) (*models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var account models.Account
	err := db.DB.Joins("JOIN users ON users.id = accounts.user_id").
		Where("users.username = ? AND accounts.account_type = ?", username, "current").
		Preload("User").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
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

func GetAllAccounts() ([]models.Account, error) {
	var accounts []models.Account
	err := db.DB.Preload("User").Find(&accounts).Error
	return accounts, err
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
		// Invalidate all active sessions when blocking
		if !active {
			db.DB.Where("user_id = ?", account.UserID).Delete(&models.Session{})
			log.Printf("Admin: All sessions invalidated for user ID %d", account.UserID)
		}
		status := "deactivated"
		if active {
			status = "activated"
		}
		log.Printf("Admin: Account %s (User ID: %d) has been %s", account.Number, account.UserID, status)
		return nil
	})
}

func MultiTransfer(senderIdentifier string, recipients []models.TransferRecipient) error {
	senderIdentifier = strings.ToLower(strings.TrimSpace(senderIdentifier))

	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Validate Sender
		var sender models.User
		query := "email = ? OR phone_number = ? OR username = ?"
		if err := tx.Where(query, senderIdentifier, senderIdentifier, senderIdentifier).First(&sender).Error; err != nil {
			return errors.New("sender not found")
		}

		var senderAcc models.Account
		if err := tx.Where("user_id = ?", sender.ID).First(&senderAcc).Error; err != nil {
			return errors.New("sender account not found")
		}

		// 2. Calculate Total needed
		var totalAmount int64
		for _, r := range recipients {
			if r.Amount < MinTransfer {
				return fmt.Errorf("transfer to %s is below minimum", r.AccountNumber)
			}
			totalAmount += r.Amount
		}

		if senderAcc.Balance < totalAmount {
			return fmt.Errorf("insufficient funds for batch: need KES %d", totalAmount)
		}

		// 3. Process each recipient
		for _, r := range recipients {
			var recAcc models.Account
			if err := tx.Where("number = ?", r.AccountNumber).First(&recAcc).Error; err != nil {
				return fmt.Errorf("recipient %s not found", r.AccountNumber)
			}

			if !recAcc.Active || recAcc.ID == senderAcc.ID {
				return fmt.Errorf("invalid recipient: %s", r.AccountNumber)
			}

			// Update Balances
			senderAcc.Balance -= r.Amount
			recAcc.Balance += r.Amount

			if err := tx.Save(&senderAcc).Error; err != nil {
				return err
			}
			if err := tx.Save(&recAcc).Error; err != nil {
				return err
			}

			// Record Log for Recipient
			var recUser models.User
			tx.Where("id = ?", recAcc.UserID).First(&recUser)
			tx.Create(&models.Transaction{Username: recUser.Username, Type: "transfer_in", Amount: r.Amount, Balance: recAcc.Balance})
		}

		// 4. Record one final log for Sender's total exit
		tx.Create(&models.Transaction{Username: sender.Username, Type: "batch_transfer_out", Amount: totalAmount, Balance: senderAcc.Balance})

		return nil
	})
}

// GetAllUsers - Fetches all users and their associated accounts for the Admin Dashboard
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	// Preload("Accounts") tells GORM to fetch the bank account for each user
	err := db.DB.Preload("Accounts").Find(&users).Error
	return users, err
}

// HasAdmin checks if there is any user with role 'admin' in the system
func HasAdmin() (bool, error) {
	var count int64
	err := db.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SetTransactionPin — hashes and saves the user's transaction PIN
func SetTransactionPin(username, pin string) error {
	pin = strings.TrimSpace(pin)

	if len(pin) != 4 {
		return errors.New("PIN must be exactly 4 digits")
	}

	for _, ch := range pin {
		if ch < '0' || ch > '9' {
			return errors.New("PIN must contain digits only")
		}
	}

	hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.DB.Model(&models.User{}).
		Where("username = ?", username).
		Update("transaction_pin", string(hashedPin)).Error
}

// VerifyTransactionPin — checks the provided PIN against the stored hash
func VerifyTransactionPin(username, pin string) error {
	pin = strings.TrimSpace(pin)

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if user.TransactionPin == "" {
		return errors.New("transaction PIN not set. Please set your PIN first")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.TransactionPin), []byte(pin)); err != nil {
		return errors.New("incorrect PIN")
	}

	return nil
}

// UpdateUserProfile — updates a user's email and phone number
func UpdateUserProfile(username, email, phone, currentPassword string) error {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify current password before allowing changes
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("incorrect password")
	}

	// Validate and clean new values
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanPhone := strings.TrimSpace(phone)

	if cleanEmail != "" && !utils.ValidEmail(cleanEmail) {
		return errors.New("invalid email address")
	}

	if cleanPhone != "" {
		if strings.HasPrefix(cleanPhone, "0") {
			cleanPhone = "254" + cleanPhone[1:]
		}
		if !utils.ValidPhoneNumber(cleanPhone) {
			return errors.New("invalid phone number")
		}
	}

	// Apply updates
	updates := map[string]interface{}{}
	if cleanEmail != "" {
		updates["email"] = cleanEmail
	}
	if cleanPhone != "" {
		updates["phone_number"] = cleanPhone
	}

	if len(updates) == 0 {
		return errors.New("no changes provided")
	}

	result := db.DB.Model(&user).Updates(updates)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return errors.New("email or phone number already in use")
		}
		return result.Error
	}

	log.Printf("✅ Profile updated for %s", username)
	return nil
}

// ChangeTransactionPin — verifies current PIN then sets a new one
func ChangeTransactionPin(username, currentPin, newPin string) error {
	// Verify current PIN first
	if err := VerifyTransactionPin(username, currentPin); err != nil {
		return errors.New("current PIN is incorrect")
	}

	if len(newPin) != 4 {
		return errors.New("new PIN must be exactly 4 digits")
	}

	for _, ch := range newPin {
		if ch < '0' || ch > '9' {
			return errors.New("PIN must contain digits only")
		}
	}

	return SetTransactionPin(username, newPin)
}

// ChangePassword — verifies current password then sets a new one
func ChangePassword(username, currentPassword, newPassword string) error {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.DB.Model(&user).Update("password", string(hashedPassword)).Error
}

// GetUserAccounts — fetches all accounts belonging to a user
func GetUserAccounts(username string) ([]models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var accounts []models.Account
	if err := db.DB.Where("user_id = ?", user.ID).Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

// CreateSavingsAccount — opens a savings account for a user who already has a current account
func CreateSavingsAccount(username string) (*models.Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}

	// Check if user already has a savings account
	var existing models.Account
	if err := db.DB.Where("user_id = ? AND account_type = ?", user.ID, "savings").First(&existing).Error; err == nil {
		return nil, errors.New("you already have a savings account")
	}

	num, err := GenerateAccountNumber()
	if err != nil {
		return nil, err
	}

	account := &models.Account{
		UserID:      user.ID,
		Number:      num,
		Balance:     0,
		Active:      true,
		AccountType: "savings",
	}

	if err := db.DB.Create(account).Error; err != nil {
		return nil, err
	}

	log.Printf("✅ Savings account created for %s: %s", username, num)
	return account, nil
}

// GenerateStatement builds a statement for a user's account over a date range.
// NOTE: Transactions are not currently tied to a specific account number in the
// schema, so this reflects all of the user's transactions across every account,
// labeled under the account number passed in.
func GenerateStatement(username, accountNumber string, from, to time.Time) (*models.StatementData, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}
	var account models.Account
	if err := db.DB.Where("number = ? AND user_id = ?", accountNumber, user.ID).First(&account).Error; err != nil {
		return nil, errors.New("account not found or does not belong to you")
	}

	toInclusive := to.Add(24*time.Hour - time.Second)

	var openingBalance int64 = 0
	var lastBefore models.Transaction
	err := db.DB.Where("username = ? AND created_at < ?", user.Username, from).
		Order("created_at desc").First(&lastBefore).Error
	if err == nil {
		openingBalance = lastBefore.Balance
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var transactions []models.Transaction
	if err := db.DB.Where("username = ? AND created_at >= ? AND created_at <= ?", user.Username, from, toInclusive).
		Order("created_at asc").
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	closingBalance := openingBalance
	if len(transactions) > 0 {
		closingBalance = transactions[len(transactions)-1].Balance
	}

	return &models.StatementData{
		AccountHolderName: user.FullName,
		AccountNumber:     account.Number,
		From:              from,
		To:                to,
		OpeningBalance:    openingBalance,
		ClosingBalance:    closingBalance,
		Transactions:      transactions,
	}, nil
}