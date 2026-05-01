package services

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"fintech-labs/backend/db"
	"fintech-labs/backend/models"
	"fintech-labs/backend/validator"

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

func init() {
	rand.Seed(time.Now().UnixNano())
}
func CreateUser(fullname, username, email, phone,Id, password, role string) (*models.User, error) {
	cleanfullname := strings.TrimSpace(fullname)
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanUsername := strings.ToLower(strings.TrimSpace(username))
	cleanPhoneNumber := strings.TrimSpace(phone)
	cleanId:=strings.TrimSpace(Id)
	if !validator.ValidEmail(cleanEmail) {
		return nil, fmt.Errorf("invalid email address")
	}

	if !validator.ValidFullName(cleanfullname) {
		return nil, fmt.Errorf("invalid full  name")
	}
	if !validator.ValidUsername(cleanUsername) {
		return nil, fmt.Errorf("invalid username:must be 3-30 characters and contains only letters,numbers or . - _")
	}
	if strings.HasPrefix(cleanPhoneNumber, "0") {
		cleanPhoneNumber = "254" + cleanPhoneNumber[1:]
	}

	if !validator.ValidPhoneNumber(cleanPhoneNumber) {
		return nil, fmt.Errorf("invalid phone number")
	}

	if !validator.ValidNationalID(cleanId){

		return nil,fmt.Errorf("invalid National ID Number")
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
		NationlID: cleanId,
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

func GenerateAccountNumber() string {
	return fmt.Sprintf("%06d", rand.Intn(900000)+100000)
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

func AuthenticateUser(Identifier,password string) (*models.User, error) {
	cleanIdentifier := strings.ToLower(strings.TrimSpace(Identifier))

 var user models.User

	if err := db.DB.Where("email = ? OR phone_number = ?" , cleanIdentifier).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}

func Deposit(Identifier string, amount int64) error {
	cleanIdentifier:= strings.ToLower(strings.TrimSpace(Identifier))

	if amount < MinDeposit {
		return fmt.Errorf("minimum deposit is KES %d", MinDeposit)
	}
	if amount > MaxDeposit {
		return fmt.Errorf("maximum deposit is KES %d", MaxDeposit)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {

		var user models.User
		query:="email = ? OR phone_number = ? OR username = ?"
		if err := tx.Where(query,cleanIdentifier,cleanIdentifier,cleanIdentifier).First(&user).Error; err != nil {
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
			Username: user.Username,
			Type:     "deposit",
			Amount:   amount,
			Balance:  account.Balance,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			log.Printf("Failed to record transaction: %v", err)
			return err
		}

		log.Printf("💰 Deposit: %s deposited KES %d (Balance: KES %d → KES %d)",
			user.Username, amount, oldBalance, account.Balance)
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


func Withdraw(Identifier string, amount int64) error {
	cleanIdentifier := strings.ToLower(strings.TrimSpace(Identifier))

	if amount < MinWithdrawal {
		return fmt.Errorf("minimum withdrawal is KES %d", MinWithdrawal)
	}
	if amount > MaxWithdrawal {
		return fmt.Errorf("maximum withdrawal is KES %d", MaxWithdrawal)
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		query:=("email = ? OR username = ? OR phone_number = ?")
		if err := tx.Where(query,cleanIdentifier,cleanIdentifier,cleanIdentifier).First(&user).Error; err != nil {
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
			Username: user.Username,
			Type:     "withdrawal",
			Amount:   amount,
			Balance:  account.Balance,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			log.Printf("Failed to record transaction: %v", err)
			return err
		}

		log.Printf("💸 Withdrawal: %s withdrew KES %d (Balance: KES %d → KES %d)",
			user.Username, amount, oldBalance, account.Balance)
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


func SendMoney(fromIdentifier, toIdentifier string, amount int64) error {
    fromIdentifier = strings.ToLower(strings.TrimSpace(fromIdentifier))
    toIdentifier = strings.TrimSpace(toIdentifier)

    if amount < MinTransfer || amount > MaxTransfer {
        return fmt.Errorf("transfer must be between KES %d and KES %d", MinTransfer, MaxTransfer)
    }

    return db.DB.Transaction(func(tx *gorm.DB) error {
        // 1. Get sender with corrected query
        var fromUser models.User
        query := "email = ? OR phone_number = ? OR username = ?"
        if err := tx.Where(query, fromIdentifier, fromIdentifier, fromIdentifier).First(&fromUser).Error; err != nil {
            return errors.New("sender not found")
        }

        // 2. Get sender's account
        var fromAccount models.Account
        if err := tx.Where("user_id = ?", fromUser.ID).First(&fromAccount).Error; err != nil {
            return errors.New("sender account not found")
        }

        if !fromAccount.Active || fromAccount.Balance < amount {
            return errors.New("transaction denied: check status or balance")
        }

        // 3. Get receiver's account by account number
        var toAccount models.Account
        if err := tx.Where("number = ?", toIdentifier).First(&toAccount).Error; err != nil {
            return errors.New("recipient account not found")
        }

        if !toAccount.Active || fromAccount.ID == toAccount.ID {
            return errors.New("invalid recipient account")
        }

        // 4. Get receiver user info for logging
        var toUser models.User
        if err := tx.Where("id = ?", toAccount.UserID).First(&toUser).Error; err != nil {
            return errors.New("recipient user not found")
        }

        // 5. ATOMIC SWAP: Execute the transfer[cite: 1]
        fromOldBalance := fromAccount.Balance
        fromAccount.Balance -= amount
        toOldBalance := toAccount.Balance
        toAccount.Balance += amount

        if err := tx.Save(&fromAccount).Error; err != nil { return err }
        if err := tx.Save(&toAccount).Error; err != nil { return err }

        // 6. RECORD LOGS: Both sides see the history[cite: 1]
        tx.Create(&models.Transaction{Username: fromUser.Username, Type: "transfer_out", Amount: amount, Balance: fromAccount.Balance})
        tx.Create(&models.Transaction{Username: toUser.Username, Type: "transfer_in", Amount: amount, Balance: toAccount.Balance})

        // FIXED LOGGING: Using correct variable types[cite: 1]
        log.Printf("💸 Transfer: %s sent KES %d to %s (Acc: %s) | Sender: %d -> %d | Recipient: %d -> %d",
            fromUser.Username, amount, toUser.Username, toAccount.Number, fromOldBalance, fromAccount.Balance, toOldBalance, toAccount.Balance)

        return nil
    })
}

func GetTransactions(Identifier string) ([]models.Transaction, error) {
	Identifier = strings.ToLower(strings.TrimSpace(Identifier))

	var user models.User
	var transactions []models.Transaction
	query:="username = ? OR email = ?"
	if err := db.DB.Where(query,Identifier,Identifier).First(&user).Error; err !=nil{
return nil,errors.New("user not found")
	}

	err := db.DB.Where("username = ?",user.Username).
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

            if err := tx.Save(&senderAcc).Error; err != nil { return err }
            if err := tx.Save(&recAcc).Error; err != nil { return err }

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