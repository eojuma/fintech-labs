package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"
	"fintech-labs/internal/notifications"
	"fintech-labs/internal/utils"

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

	DailyDepositLimit    = 500000
	DailyWithdrawalLimit = 100000
	DailyTransferLimit   = 250000
)

var dummyHash []byte

func init() {
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
}

func enforceDailyLimit(tx *gorm.DB, accountNumber, transactionType string, amount, limit int64) error {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var total int64
	if err := tx.Model(&models.Transaction{}).
		Where("account_number = ? AND type = ? AND created_at >= ? AND status != ?", accountNumber, transactionType, start, "failed").
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return err
	}
	if total+amount > limit {
		return fmt.Errorf("daily %s limit of KES %d would be exceeded", strings.ReplaceAll(transactionType, "_", " "), limit)
	}
	return nil
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

	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
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
		UserID:      userID,
		Number:      num,
		Balance:     0,
		Active:      true,
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

func VerifyDevice(userID uint, fingerprint, userAgent string) error {
	if fingerprint == "" {
		return errors.New("device fingerprint is required")
	}
	var device models.Device
	result := db.DB.Where("user_id = ? AND fingerprint = ?", userID, fingerprint).First(&device)
	if result.Error == nil {
		if !device.Approved {
			return errors.New("this device is pending administrator approval")
		}
		db.DB.Model(&device).Updates(map[string]interface{}{"last_seen_at": time.Now(), "user_agent": userAgent})
		return nil
	}
	device = models.Device{UserID: userID, Fingerprint: fingerprint, UserAgent: userAgent, LastSeenAt: time.Now(), Approved: false}
	if err := db.DB.Create(&device).Error; err != nil {
		return err
	}
	return errors.New("new device recorded and is pending administrator approval")
}

func Deposit(username, accountNumber string, amount int64) (string, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	accountNumber = strings.TrimSpace(accountNumber)

	var refNum string
	var transactionID uint

	if amount < MinDeposit {
		return "", fmt.Errorf("minimum deposit is KES %d", MinDeposit)
	}
	if amount > MaxDeposit {
		return "", fmt.Errorf("maximum deposit is KES %d", MaxDeposit)
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		var account models.Account
		if err := tx.Where("number = ?", accountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		if account.UserID != user.ID {
			return errors.New("you can only operate on your own account")
		}

		if !account.Active {
			return errors.New("account is inactive")
		}
		if err := enforceDailyLimit(tx, account.Number, "deposit", amount, DailyDepositLimit); err != nil {
			return err
		}

		oldBalance := account.Balance
		account.Balance += amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username:        user.Username,
			AccountNumber:   account.Number,
			ReferenceNumber: GenerateReferenceNumber(),
			Type:            "deposit",
			Amount:          amount,
			Balance:         account.Balance,
			Status:          "completed",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}
		refNum = transaction.ReferenceNumber
		transactionID = transaction.ID
		log.Printf("💰 Deposit: %s deposited KES %d to account %s (Balance: KES %d → KES %d)",
			user.Username, amount, account.Number, oldBalance, account.Balance)

		// Send email notification in background
		go func() {
			emailData := models.TransactionEmailData{
				FullName:        user.FullName,
				Email:           user.Email,
				Type:            "Deposit",
				Amount:          amount,
				Balance:         account.Balance,
				AccountNumber:   account.Number,
				ReferenceNumber: transaction.ReferenceNumber,
				Timestamp:       utils.FormatDate(transaction.CreatedAt),
			}
			if err := notifications.SendTransactionEmail(user.Email, emailData); err != nil {
				log.Printf("⚠️ Failed to send deposit email to %s: %v", user.Email, err)
			}
		}()
		// Send SMS notification in background
		go func() {
			if user.SMSOptOut {
				return
			}
			phone := utils.FormatPhoneForSMS(user.PhoneNumber)
			message := notifications.FormatSMSMessage("deposited", account.Number, transaction.ReferenceNumber, amount, account.Balance)
			if err := notifications.SendTransactionSMS(phone, message); err != nil {
				log.Printf("⚠️ Failed to send deposit SMS to %s: %v", phone, err)
			}
		}()
		return nil
	})
	if err != nil {
		return "", err
	}
	go CheckSuspiciousActivity(transactionID, accountNumber, amount, "deposit")
	return refNum, err
}

// AdminDeposit - Admin deposit to any user account
func AdminDeposit(adminUsername, accountNumber string, amount int64) error {
	accountNumber = strings.TrimSpace(accountNumber)

	if amount <= 0 {
		CreateAuditLog(adminUsername, "deposit", accountNumber, "Invalid amount", amount, "failed")
		return errors.New("amount must be greater than zero")
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.Where("number = ?", accountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		if !account.Active {
			return errors.New("account is inactive")
		}
		if err := enforceDailyLimit(tx, account.Number, "deposit", amount, DailyDepositLimit); err != nil {
			return err
		}

		var user models.User
		if err := tx.First(&user, account.UserID).Error; err != nil {
			return errors.New("user not found")
		}

		account.Balance += amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username:        user.Username,
			AccountNumber:   account.Number,
			ReferenceNumber: GenerateReferenceNumber(),
			Type:            "deposit",
			Amount:          amount,
			Balance:         account.Balance,
			Status:          "completed",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		log.Printf("Admin deposit: KES %d to account %s by admin %s", amount, account.Number, adminUsername)

		// Email notification
		go func() {
			emailData := models.TransactionEmailData{
				FullName:        user.FullName,
				Email:           user.Email,
				Type:            "Deposit",
				Amount:          amount,
				Balance:         account.Balance,
				AccountNumber:   account.Number,
				ReferenceNumber: transaction.ReferenceNumber,
				Timestamp:       utils.FormatDate(transaction.CreatedAt),
			}
			if err := notifications.SendTransactionEmail(user.Email, emailData); err != nil {
				log.Printf("⚠️ Failed to send admin deposit email to %s: %v", user.Email, err)
			}
		}()

		// SMS notification
		go func() {
			if user.SMSOptOut {
				return
			}
			phone := utils.FormatPhoneForSMS(user.PhoneNumber)
			message := notifications.FormatSMSMessage("deposited", account.Number, transaction.ReferenceNumber, amount, account.Balance)
			if err := notifications.SendTransactionSMS(phone, message); err != nil {
				log.Printf("⚠️ Failed to send admin deposit SMS to %s: %v", phone, err)
			}
		}()

		return nil
	})
	if err != nil {
		CreateAuditLog(adminUsername, "deposit", accountNumber, err.Error(), amount, "failed")
		return err
	}

	CreateAuditLog(adminUsername, "deposit", accountNumber, fmt.Sprintf("Admin deposited KES %d successfully", amount), amount, "success")
	return nil
}

func Withdraw(username, accountNumber string, amount int64) (string, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	accountNumber = strings.TrimSpace(accountNumber)

	var refNum string
	var transactionID uint

	if amount < MinWithdrawal {
		return "", fmt.Errorf("minimum withdrawal is KES %d", MinWithdrawal)
	}
	if amount > MaxWithdrawal {
		return "", fmt.Errorf("maximum withdrawal is KES %d", MaxWithdrawal)
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}

		var account models.Account
		if err := tx.Where("number = ?", accountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		if account.UserID != user.ID {
			return errors.New("you can only operate on your own account")
		}

		if !account.Active {
			return errors.New("account is inactive")
		}
		if err := enforceDailyLimit(tx, account.Number, "withdrawal", amount, DailyWithdrawalLimit); err != nil {
			return err
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
			Username:        user.Username,
			AccountNumber:   account.Number,
			ReferenceNumber: GenerateReferenceNumber(),
			Type:            "withdrawal",
			Amount:          amount,
			Balance:         account.Balance,
			Status:          "completed",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}
		refNum = transaction.ReferenceNumber
		transactionID = transaction.ID
		log.Printf("💸 Withdrawal: %s withdrew KES %d from account %s (Balance: KES %d → KES %d)",
			user.Username, amount, account.Number, oldBalance, account.Balance)

		go func() {
			emailData := models.TransactionEmailData{
				FullName:        user.FullName,
				Email:           user.Email,
				Type:            "Withdrawal",
				Amount:          amount,
				Balance:         account.Balance,
				AccountNumber:   account.Number,
				ReferenceNumber: transaction.ReferenceNumber,
				Timestamp:       utils.FormatDate(transaction.CreatedAt),
			}
			if err := notifications.SendTransactionEmail(user.Email, emailData); err != nil {
				log.Printf("⚠️ Failed to send withdrawal email to %s: %v", user.Email, err)
			}
		}()
		go func() {
			if user.SMSOptOut {
				return
			}
			phone := utils.FormatPhoneForSMS(user.PhoneNumber)
			message := notifications.FormatSMSMessage("withdrawn", account.Number, transaction.ReferenceNumber, amount, account.Balance)
			if err := notifications.SendTransactionSMS(phone, message); err != nil {
				log.Printf("⚠️ Failed to send withdrawal SMS to %s: %v", phone, err)
			}
		}()
		return nil
	})
	if err != nil {
		return "", err
	}
	go CheckSuspiciousActivity(transactionID, accountNumber, amount, "withdrawal")
	return refNum, err
}

// AdminWithdraw - Admin withdrawal from any user account
func AdminWithdraw(adminUsername, accountNumber string, amount int64) error {
	accountNumber = strings.TrimSpace(accountNumber)

	if amount <= 0 {
		CreateAuditLog(adminUsername, "withdrawal", accountNumber, "Invalid amount", amount, "failed")
		return errors.New("amount must be greater than zero")
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.Where("number = ?", accountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		if !account.Active {
			return errors.New("account is inactive")
		}
		if err := enforceDailyLimit(tx, account.Number, "withdrawal", amount, DailyWithdrawalLimit); err != nil {
			return err
		}

		var user models.User
		if err := tx.First(&user, account.UserID).Error; err != nil {
			return errors.New("user not found")
		}

		if account.Balance < amount {
			return fmt.Errorf("insufficient funds. Balance is KES %d", account.Balance)
		}

		account.Balance -= amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			Username:        user.Username,
			AccountNumber:   account.Number,
			ReferenceNumber: GenerateReferenceNumber(),
			Type:            "withdrawal",
			Amount:          amount,
			Balance:         account.Balance,
			Status:          "completed",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		log.Printf("Admin withdrawal: KES %d from account %s by admin %s", amount, account.Number, adminUsername)

		// Email notification
		go func() {
			emailData := models.TransactionEmailData{
				FullName:        user.FullName,
				Email:           user.Email,
				Type:            "Withdrawal",
				Amount:          amount,
				Balance:         account.Balance,
				AccountNumber:   account.Number,
				ReferenceNumber: transaction.ReferenceNumber,
				Timestamp:       utils.FormatDate(transaction.CreatedAt),
			}
			if err := notifications.SendTransactionEmail(user.Email, emailData); err != nil {
				log.Printf("⚠️ Failed to send admin withdrawal email to %s: %v", user.Email, err)
			}
		}()

		// SMS notification
		go func() {
			if user.SMSOptOut {
				return
			}
			phone := utils.FormatPhoneForSMS(user.PhoneNumber)
			message := notifications.FormatSMSMessage("withdrawn", account.Number, transaction.ReferenceNumber, amount, account.Balance)
			if err := notifications.SendTransactionSMS(phone, message); err != nil {
				log.Printf("⚠️ Failed to send admin withdrawal SMS to %s: %v", phone, err)
			}
		}()

		return nil
	})
	if err != nil {
		CreateAuditLog(adminUsername, "withdrawal", accountNumber, err.Error(), amount, "failed")
		return err
	}

	CreateAuditLog(adminUsername, "withdrawal", accountNumber, fmt.Sprintf("Admin withdrew KES %d successfully", amount), amount, "success")
	return nil
}

// ResolveRecipientAccount finds a recipient's account using either an account
// number or a phone number. Account numbers follow the "AV" prefix convention
// (e.g. AV202600000001); anything else is treated as a phone number and
// normalized the same way as registration (leading 0 -> 254).
func ResolveRecipientAccount(tx *gorm.DB, identifier string) (*models.Account, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, errors.New("recipient identifier is required")
	}

	// Looks like an account number
	if strings.HasPrefix(strings.ToUpper(identifier), "AV") {
		var account models.Account
		if err := tx.Where("number = ?", identifier).First(&account).Error; err != nil {
			return nil, errors.New("recipient account not found")
		}
		return &account, nil
	}

	// Otherwise, treat it as a phone number
	cleanPhone := identifier
	if strings.HasPrefix(cleanPhone, "0") {
		cleanPhone = "254" + cleanPhone[1:]
	} else if strings.HasPrefix(cleanPhone, "+") {
		cleanPhone = cleanPhone[1:]
	}
	var user models.User
	if err := tx.Where("phone_number = ?", cleanPhone).First(&user).Error; err != nil {
		return nil, errors.New("recipient not found")
	}

	var account models.Account
	if err := tx.Where("user_id = ? AND account_type = ?", user.ID, "current").First(&account).Error; err != nil {
		return nil, errors.New("recipient account not found")
	}

	return &account, nil
}

func SendMoney(username, fromAccountNumber, toIdentifier string, amount int64) (string, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	fromAccountNumber = strings.TrimSpace(fromAccountNumber)
	toIdentifier = strings.TrimSpace(toIdentifier)
	var refNum string
	var transactionID uint
	if amount < MinTransfer || amount > MaxTransfer {
		return "", fmt.Errorf("transfer must be between KES %d and KES %d", MinTransfer, MaxTransfer)
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var fromUser models.User
		if err := tx.Where("username = ?", username).First(&fromUser).Error; err != nil {
			return errors.New("sender not found")
		}

		var fromAccount models.Account
		if err := tx.Where("number = ?", fromAccountNumber).First(&fromAccount).Error; err != nil {
			return errors.New("sender account not found")
		}

		if fromAccount.UserID != fromUser.ID {
			return errors.New("you can only transfer from your own account")
		}

		if !fromAccount.Active {
			return errors.New("transaction denied: account inactive")
		}
		if err := enforceDailyLimit(tx, fromAccount.Number, "transfer_out", amount, DailyTransferLimit); err != nil {
			return err
		}

		if fromAccount.Balance < amount {
			return fmt.Errorf("insufficient funds. Your balance is KES %d", fromAccount.Balance)
		}

		toAccountPtr, err := ResolveRecipientAccount(tx, toIdentifier)
		if err != nil {
			return err
		}
		toAccount := *toAccountPtr

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
		outRef := GenerateReferenceNumber()
		inRef := GenerateReferenceNumber()
		outTransaction := models.Transaction{
			Username:        fromUser.Username,
			AccountNumber:   fromAccount.Number,
			ReferenceNumber: outRef,
			Type:            "transfer_out",
			Amount:          amount,
			Balance:         fromAccount.Balance,
			Status:          "completed",
		}
		if err := tx.Create(&outTransaction).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.Transaction{
			Username:        toUser.Username,
			AccountNumber:   toAccount.Number,
			ReferenceNumber: inRef,
			Type:            "transfer_in",
			Amount:          amount,
			Balance:         toAccount.Balance,
			Status:          "completed",
		}).Error; err != nil {
			return err
		}
		refNum = outRef
		transactionID = outTransaction.ID
		log.Printf("💸 Transfer: %s sent KES %d to %s | Sender: %d -> %d | Recipient: %d -> %d",
			fromUser.Username, amount, toUser.Username, fromOldBalance, fromAccount.Balance, toOldBalance, toAccount.Balance)

		// Notify sender
		go func() {
			emailData := models.TransactionEmailData{
				FullName:        fromUser.FullName,
				Email:           fromUser.Email,
				Type:            "Transfer Out",
				Amount:          amount,
				Balance:         fromAccount.Balance,
				AccountNumber:   fromAccount.Number,
				ReferenceNumber: outRef,
				Timestamp:       utils.FormatDate(time.Now()),
			}
			if err := notifications.SendTransactionEmail(fromUser.Email, emailData); err != nil {
				log.Printf("⚠️ Failed to send transfer email to sender %s: %v", fromUser.Email, err)
			}
		}()

		// Notify recipient
		go func() {
			emailData := models.TransactionEmailData{
				FullName:        toUser.FullName,
				Email:           toUser.Email,
				Type:            "Transfer In",
				Amount:          amount,
				Balance:         toAccount.Balance,
				AccountNumber:   toAccount.Number,
				ReferenceNumber: inRef,
				Timestamp:       utils.FormatDate(time.Now()),
			}
			if err := notifications.SendTransactionEmail(toUser.Email, emailData); err != nil {
				log.Printf("⚠️ Failed to send transfer email to recipient %s: %v", toUser.Email, err)
			}
		}()

		// SMS — sender
		fromUserCopy2 := fromUser
		fromAccountCopy2 := fromAccount
		go func() {
			if fromUserCopy2.SMSOptOut {
				return
			}
			phone := utils.FormatPhoneForSMS(fromUserCopy2.PhoneNumber)
			message := notifications.FormatSMSMessage("transferred out", fromAccountCopy2.Number, outRef, amount, fromAccountCopy2.Balance)
			if err := notifications.SendTransactionSMS(phone, message); err != nil {
				log.Printf("⚠️ Failed to send transfer SMS to sender %s: %v", phone, err)
			}
		}()

		// SMS — recipient
		toUserCopy2 := toUser
		toAccountCopy2 := toAccount
		go func() {
			if toUserCopy2.SMSOptOut {
				return
			}
			phone := utils.FormatPhoneForSMS(toUserCopy2.PhoneNumber)
			message := notifications.FormatSMSMessage("received", toAccountCopy2.Number, inRef, amount, toAccountCopy2.Balance)
			if err := notifications.SendTransactionSMS(phone, message); err != nil {
				log.Printf("⚠️ Failed to send transfer SMS to recipient %s: %v", phone, err)
			}
		}()
		return nil
	})
	if err != nil {
		return "", err
	}
	go CheckSuspiciousActivity(transactionID, fromAccountNumber, amount, "transfer_out")
	return refNum, err
}

func CreateRecurringTransfer(username, senderAccount, recipient string, amount int64, frequency string, nextRunAt time.Time) (*models.RecurringTransfer, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	frequency = strings.ToLower(strings.TrimSpace(frequency))
	if amount < MinTransfer || amount > MaxTransfer {
		return nil, fmt.Errorf("amount must be between KES %d and KES %d", MinTransfer, MaxTransfer)
	}
	if frequency != "daily" && frequency != "weekly" && frequency != "monthly" {
		return nil, errors.New("frequency must be daily, weekly, or monthly")
	}
	if nextRunAt.Before(time.Now()) {
		return nil, errors.New("next run must be in the future")
	}
	account, err := GetAccountByNumber(senderAccount)
	if err != nil || account.User.Username != username {
		return nil, errors.New("sender account not found")
	}
	recurring := &models.RecurringTransfer{Username: username, SenderAccount: senderAccount, Recipient: strings.TrimSpace(recipient), Amount: amount, Frequency: frequency, NextRunAt: nextRunAt, Active: true}
	if recurring.Recipient == "" {
		return nil, errors.New("recipient is required")
	}
	if err := db.DB.Create(recurring).Error; err != nil {
		return nil, err
	}
	return recurring, nil
}

func ProcessDueRecurringTransfers(now time.Time) (int, error) {
	var jobs []models.RecurringTransfer
	if err := db.DB.Where("active = ? AND next_run_at <= ?", true, now).Find(&jobs).Error; err != nil {
		return 0, err
	}
	processed := 0
	for _, job := range jobs {
		_, err := SendMoney(job.Username, job.SenderAccount, job.Recipient, job.Amount)
		updates := map[string]interface{}{}
		if err != nil {
			updates["last_error"] = err.Error()
		} else {
			processed++
			updates["last_error"] = ""
		}
		next := job.NextRunAt
		switch job.Frequency {
		case "daily":
			next = next.AddDate(0, 0, 1)
		case "weekly":
			next = next.AddDate(0, 0, 7)
		case "monthly":
			next = next.AddDate(0, 1, 0)
		}
		updates["next_run_at"] = next
		if err := db.DB.Model(&job).Updates(updates).Error; err != nil {
			return processed, err
		}
	}
	return processed, nil
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
func ToggleAccountStatus(adminUsername string, accountID uint, active bool) error {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var account models.Account
		if err := tx.First(&account, accountID).Error; err != nil {
			return errors.New("account not found")
		}

		account.Active = active
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

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
	if err != nil {
		action := "block"
		if active {
			action = "unblock"
		}
		CreateAuditLog(adminUsername, action, fmt.Sprintf("ID:%d", accountID), err.Error(), 0, "failed")
		return err
	}

	action := "block"
	if active {
		action = "unblock"
	}
	CreateAuditLog(adminUsername, action, fmt.Sprintf("ID:%d", accountID), fmt.Sprintf("Account %s successfully", action+"ed"), 0, "success")
	return nil
}

// GetAllUsers - Fetches all users and their associated accounts for the Admin Dashboard
func GetAllUsers() ([]models.User, error) {
	var users []models.User
	// Preload("Accounts") tells GORM to fetch the bank account for each user
	err := db.DB.Preload("Accounts").Find(&users).Error
	return users, err
}

// HasAdmin checks if there is any administrator in the system.
func HasAdmin() (bool, error) {
	var count int64
	err := db.DB.Model(&models.User{}).Where("role IN ?", []string{"admin", "super_admin"}).Count(&count).Error
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
	username = strings.ToLower(strings.TrimSpace(username))
	pin = strings.TrimSpace(pin)

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if user.PinLockedUntil != nil && time.Now().Before(*user.PinLockedUntil) {
		remaining := time.Until(*user.PinLockedUntil).Round(time.Second)
		return fmt.Errorf("PIN entry locked. Try again in %v", remaining)
	}

	if user.TransactionPin == "" {
		return errors.New("transaction PIN not set. Please set your PIN first")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.TransactionPin), []byte(pin)); err != nil {
		user.PinFailedAttempts++
		if user.PinFailedAttempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			user.PinLockedUntil = &lockUntil
			user.PinFailedAttempts = 0
			db.DB.Save(&user)
			return errors.New("PIN locked for 15 minutes due to too many failed attempts")
		}
		db.DB.Save(&user)
		attemptsLeft := 5 - user.PinFailedAttempts
		return fmt.Errorf("incorrect PIN. %d attempt(s) remaining before lockout", attemptsLeft)
	}

	user.PinFailedAttempts = 0
	user.PinLockedUntil = nil
	db.DB.Save(&user)

	return nil
}

// UpdateUserProfile — updates a user's email and phone number
func UpdateUserProfile(username, email, phone, currentPassword string, smsOptOut bool) error {
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
	updates["sms_opt_out"] = smsOptOut

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

// CloseUserAccount permanently disables a member's access while retaining
// financial records through GORM soft deletes.
func CloseUserAccount(username, currentPassword string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("username = ?", username).First(&user).Error; err != nil {
			return errors.New("user not found")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
			return errors.New("current password is incorrect")
		}
		var accounts []models.Account
		if err := tx.Where("user_id = ?", user.ID).Find(&accounts).Error; err != nil {
			return err
		}
		for _, account := range accounts {
			if account.Balance != 0 {
				return errors.New("all account balances must be zero before closure")
			}
		}
		var shareCapital models.ShareCapital
		err := tx.Where("user_id = ?", user.ID).First(&shareCapital).Error
		if err == nil && shareCapital.Balance != 0 {
			return errors.New("share capital must be redeemed by an administrator before closure")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var openLoans int64
		if err := tx.Model(&models.Loan{}).Where("user_id = ? AND status IN ?", user.ID, []string{"pending", "approved", "disbursed"}).Count(&openLoans).Error; err != nil {
			return err
		}
		if openLoans > 0 {
			return errors.New("open loans must be resolved before closure")
		}
		if err := tx.Model(&models.Account{}).Where("user_id = ?", user.ID).Update("active", false).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Session{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user).Error
	})
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
	err := db.DB.Where("account_number = ? AND created_at < ?", accountNumber, from).
		Order("created_at desc").First(&lastBefore).Error
	if err == nil {
		openingBalance = lastBefore.Balance
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var transactions []models.Transaction
	if err := db.DB.Where("account_number = ? AND created_at >= ? AND created_at <= ?", accountNumber, from, toInclusive).
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
		AccountType:       account.AccountType,
		From:              from,
		To:                to,
		OpeningBalance:    openingBalance,
		ClosingBalance:    closingBalance,
		Transactions:      transactions,
	}, nil
}

// GenerateReferenceNumber — creates a unique transaction reference number
func GenerateReferenceNumber() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Practically unreachable; fall back to a time-based value.
		return fmt.Sprintf("AV-%d-%d", time.Now().Year(), time.Now().UnixNano())
	}
	return fmt.Sprintf("AV-%d-%X", time.Now().Year(), b)
}

// GetTransactionByReference — fetches a transaction by its reference number
func GetTransactionByReference(reference string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := db.DB.Where("reference_number = ?", reference).First(&transaction).Error; err != nil {
		return nil, errors.New("transaction not found")
	}
	return &transaction, nil
}

// FilterTransactions — fetches paginated and filtered transactions for a user
func FilterTransactions(username string, f models.TransactionFilter) (*models.FilterResult, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	// Default values
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.Page == 0 {
		f.Page = 1
	}
	if f.SortOrder != "asc" {
		f.SortOrder = "desc"
	}

	query := db.DB.Model(&models.Transaction{}).Where("username = ?", username)

	// Apply account number filter
	if f.AccountNumber != "" {
		query = query.Where("account_number = ?", f.AccountNumber)
	}

	// Apply type filter
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
	}

	// Apply date range filter
	if f.From != "" {
		from, err := time.Parse("2006-01-02", f.From)
		if err == nil {
			query = query.Where("created_at >= ?", from)
		}
	}
	if f.To != "" {
		to, err := time.Parse("2006-01-02", f.To)
		if err == nil {
			query = query.Where("created_at <= ?", to.Add(24*time.Hour))
		}
	}

	// Apply amount range filter
	if f.MinAmount > 0 {
		query = query.Where("amount >= ?", f.MinAmount)
	}
	if f.MaxAmount > 0 {
		query = query.Where("amount <= ?", f.MaxAmount)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Get aggregate amounts
	var deposits, withdrawals struct{ Total int64 }
	db.DB.Model(&models.Transaction{}).
		Where("username = ? AND type = ?", username, "deposit").
		Select("SUM(amount) as total").Scan(&deposits)
	db.DB.Model(&models.Transaction{}).
		Where("username = ? AND (type = ? OR type = ?)", username, "withdrawal", "transfer_out").
		Select("SUM(amount) as total").Scan(&withdrawals)

	// Apply pagination and sort
	offset := (f.Page - 1) * f.Limit
	var transactions []models.Transaction
	if err := query.
		Order("created_at " + f.SortOrder).
		Limit(f.Limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	totalPages := int(totalCount) / f.Limit
	if int(totalCount)%f.Limit != 0 {
		totalPages++
	}

	return &models.FilterResult{
		Transactions:     transactions,
		TotalCount:       totalCount,
		TotalDeposits:    deposits.Total,
		TotalWithdrawals: withdrawals.Total,
		Page:             f.Page,
		Limit:            f.Limit,
		TotalPages:       totalPages,
	}, nil
}

func CreateAuditLog(adminUsername, action, targetAccount, details string, amount int64, result string) {
	entry := models.AuditLog{
		AdminUsername: adminUsername,
		Action:        action,
		TargetAccount: targetAccount,
		Amount:        amount,
		Result:        result,
		Details:       details,
	}

	if err := db.DB.Create(&entry).Error; err != nil {
		fmt.Printf("⚠️Failed to create audit log: %v\n", err)
	}
}

// GetAuditLogs — fetches all audit logs ordered by most recent first
func GetAuditLogs() ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := db.DB.Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetDashboardStats — returns aggregated stats for the admin dashboard
func GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Today's date range
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Total deposits today
	var depositResult struct{ Total int64 }
	db.DB.Model(&models.Transaction{}).
		Where("type = ? AND created_at >= ? AND created_at < ? AND status = ?", "deposit", startOfDay, endOfDay, "completed").
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&depositResult)
	stats.TotalDepositsToday = depositResult.Total

	// Total withdrawals today
	var withdrawResult struct{ Total int64 }
	db.DB.Model(&models.Transaction{}).
		Where("type = ? AND created_at >= ? AND created_at < ? AND status = ?", "withdrawal", startOfDay, endOfDay, "completed").
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&withdrawResult)
	stats.TotalWithdrawalsToday = withdrawResult.Total

	// Total active users
	db.DB.Model(&models.User{}).Count(&stats.ActiveUsers)
	// Total funds in system
	var fundsResult struct{ Total int64 }
	db.DB.Model(&models.Account{}).
		Where("active = ?", true).
		Select("COALESCE(SUM(balance), 0) as total").
		Scan(&fundsResult)
	stats.TotalFunds = fundsResult.Total

	// Weekly volume — last 7 days
	weeklyVolume := []models.DailyVolume{}
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		dayEnd := dayStart.Add(24 * time.Hour)

		var result struct {
			Count int64
			Total int64
		}
		db.DB.Model(&models.Transaction{}).
			Where("created_at >= ? AND created_at < ? AND status = ?", dayStart, dayEnd, "completed").
			Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
			Scan(&result)

		weeklyVolume = append(weeklyVolume, models.DailyVolume{
			Day:   day.Format("Mon 02"),
			Count: result.Count,
			Total: result.Total,
		})
	}
	stats.WeeklyVolume = weeklyVolume

	return stats, nil
}

// CheckSuspiciousActivity — checks a transaction for suspicious patterns and flags it if needed
func CheckSuspiciousActivity(transactionID uint, accountNumber string, amount int64, transactionType string) {
	var reasons []string

	// Rule 1 — Large single transaction above KES 100,000
	if amount >= 100000 {
		reasons = append(reasons, fmt.Sprintf("Large %s of KES %d exceeds threshold", transactionType, amount))
	}

	// Rule 2 — Rapid successive transactions — 3 or more in 5 minutes
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	var recentCount int64
	db.DB.Model(&models.Transaction{}).
		Where("account_number = ? AND created_at >= ?", accountNumber, fiveMinutesAgo).
		Count(&recentCount)
	if recentCount >= 3 {
		reasons = append(reasons, fmt.Sprintf("Rapid transactions: %d transactions in the last 5 minutes", recentCount))
	}

	// Rule 3 — Large withdrawal relative to balance (more than 80% of balance)
	if transactionType == "withdrawal" {
		var account models.Account
		if err := db.DB.Where("number = ?", accountNumber).First(&account).Error; err == nil {
			if account.Balance > 0 {
				percentage := (amount * 100) / (account.Balance + amount)
				if percentage >= 80 {
					reasons = append(reasons, fmt.Sprintf("Withdrawal of KES %d is %d%% of account balance", amount, percentage))
				}
			}
		}
	}

	// If any rules triggered — flag the transaction
	if len(reasons) > 0 {
		flagReason := strings.Join(reasons, " | ")
		db.DB.Model(&models.Transaction{}).
			Where("id = ?", transactionID).
			Updates(map[string]interface{}{
				"flagged":     true,
				"flag_reason": flagReason,
			})
		log.Printf("🚨 Suspicious transaction flagged — ID: %d | Reason: %s", transactionID, flagReason)
	}
}

// GetFlaggedTransactions — fetches all flagged transactions ordered by most recent first
func GetFlaggedTransactions() ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := db.DB.Where("flagged = ?", true).
		Order("created_at desc").
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// AssignTellerRole — promotes a customer to teller role
func AssignTellerRole(adminUsername, targetUsername string) error {
	targetUsername = strings.ToLower(strings.TrimSpace(targetUsername))

	var user models.User
	if err := db.DB.Where("username = ?", targetUsername).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if user.Role == "admin" {
		return errors.New("cannot assign teller role to an admin")
	}

	if user.Role == "teller" {
		return errors.New("user is already a teller")
	}

	if err := db.DB.Model(&user).Update("role", "teller").Error; err != nil {
		return err
	}

	// Audit log
	CreateAuditLog(adminUsername, "assign_teller", user.Username, fmt.Sprintf("Admin assigned teller role to %s", targetUsername), 0, "success")
	log.Printf("✅ Admin %s assigned teller role to %s", adminUsername, targetUsername)
	return nil
}

// RevokeTellerRole — demotes a teller back to customer
func RevokeTellerRole(adminUsername, targetUsername string) error {
	targetUsername = strings.ToLower(strings.TrimSpace(targetUsername))

	var user models.User
	if err := db.DB.Where("username = ?", targetUsername).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if user.Role != "teller" {
		return errors.New("user is not a teller")
	}

	if err := db.DB.Model(&user).Update("role", "customer").Error; err != nil {
		return err
	}

	// Audit log
	CreateAuditLog(adminUsername, "revoke_teller", user.Username, fmt.Sprintf("Admin revoked teller role from %s", targetUsername), 0, "success")
	log.Printf("✅ Admin %s revoked teller role from %s", adminUsername, targetUsername)
	return nil
}

// CreatePendingMpesaTransaction — saves a pending M-Pesa transaction while waiting for callback
func CreatePendingMpesaTransaction(username, accountNumber, phoneNumber string, amount int64, checkoutRequestID, merchantRequestID string) error {
	username = strings.ToLower(strings.TrimSpace(username))

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	var account models.Account
	if err := db.DB.Where("number = ?", accountNumber).First(&account).Error; err != nil {
		return errors.New("account not found")
	}

	if account.UserID != user.ID {
		return errors.New("you can only deposit into your own account")
	}

	transaction := models.Transaction{
		Username:          username,
		AccountNumber:     accountNumber,
		ReferenceNumber:   GenerateReferenceNumber(),
		Type:              "mpesa_deposit",
		Amount:            amount,
		Balance:           account.Balance,
		Status:            "pending",
		MpesaPhoneNumber:  phoneNumber,
		CheckoutRequestID: checkoutRequestID,
		MerchantRequestID: merchantRequestID,
	}

	return db.DB.Create(&transaction).Error
}

// ProcessMpesaDeposit — credits account after successful M-Pesa callback
func ProcessMpesaDeposit(checkoutRequestID, receiptCode, amountStr string) error {
	callbackAmount, err := strconv.ParseInt(strings.TrimSpace(amountStr), 10, 64)
	if err != nil || callbackAmount <= 0 {
		return errors.New("invalid M-Pesa callback amount")
	}
	var transaction models.Transaction
	if err := db.DB.Where("checkout_request_id = ?", checkoutRequestID).First(&transaction).Error; err != nil {
		return fmt.Errorf("pending transaction not found for CheckoutRequestID: %s", checkoutRequestID)
	}
	if transaction.Amount != callbackAmount {
		return fmt.Errorf("M-Pesa amount mismatch: expected KES %d, received KES %d", transaction.Amount, callbackAmount)
	}
	if transaction.Status == "completed" {
		if transaction.MpesaReceiptCode == receiptCode {
			return nil
		}
		return errors.New("M-Pesa transaction has already been completed")
	}
	if transaction.Status != "pending" {
		return fmt.Errorf("M-Pesa transaction is not pending (status: %s)", transaction.Status)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		updated := tx.Model(&models.Transaction{}).Where("id = ? AND status = ?", transaction.ID, "pending").Updates(map[string]interface{}{
			"status": "completed", "mpesa_receipt_code": receiptCode,
		})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return errors.New("M-Pesa transaction was already processed")
		}
		var account models.Account
		if err := tx.Where("number = ?", transaction.AccountNumber).First(&account).Error; err != nil {
			return errors.New("account not found")
		}

		account.Balance += transaction.Amount
		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		transaction.Balance = account.Balance
		if err := tx.Save(&transaction).Error; err != nil {
			return err
		}

		log.Printf("✅ M-Pesa deposit processed — Account: %s | Amount: KES %d | Receipt: %s",
			account.Number, transaction.Amount, receiptCode)
		return nil
	})
}

// FailMpesaTransaction — marks a pending M-Pesa transaction as failed
func FailMpesaTransaction(checkoutRequestID, reason string) {
	db.DB.Model(&models.Transaction{}).
		Where("checkout_request_id = ?", checkoutRequestID).
		Updates(map[string]interface{}{
			"status": "failed",
		})
	log.Printf("❌ M-Pesa transaction failed — CheckoutRequestID: %s | Reason: %s", checkoutRequestID, reason)
}
