package services

import (
	"errors"
	"log"

	"fintech-labs/db"
	"fintech-labs/models"

	"gorm.io/gorm"
)

var (
	accounts     = make(map[string]models.Account)
	transactions = make(map[string][]models.Transaction)
)

const (
	MinDeposit    = 50
	MinWithdrawal = 100
)

func WithdrawalProcess(username string, amount int64) (error) {
	var account models.Account
	if err := db.DB.Where("username=?", username).First(&account).Error; err != nil {
		log.Printf("Withdrawal failed: User %s not found", username)
		return errors.New("Account not found")
	}

	if !account.Active{
		return errors.New("Account is inactivate")
	}
	if account.Balance < amount {
		return errors.New("Insufficient funds")
	}

	if amount < MinWithdrawal {
		return  errors.New("Minimun withdrawal is ksh.100")
	}


	return db.DB.Transaction(func (tx *gorm.DB)error{
		account.Balance-=amount
		if err:=tx.Save(&account).Error;err !=nil{
			return err
		}
		transaction:=models.Transaction{
			Username: username,
			Type: "withdrawal",
			Amount: amount,
			Balance: account.Balance,
		}
		if err :=tx.Create(&transaction).Error;err !=nil{
			return err
		}
		log.Printf("Successfully withdrew %d from %s",amount,username)
		return nil
	})
}

func DepositProcess(username string, amount int64) (error) {
var account models.Account
	
if err :=db.DB.Where("username=?",username).First(&account).Error;err !=nil{
	log.Printf("Deposit failed: User %s not found",username)
	return errors.New("Account not found")
}
	if !account.Active {
		return errors.New("Account is inactive")
	}
	if amount < MinDeposit {
		return errors.New("Minimum deposit is ksh.50")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error{
account.Balance+=amount
if err:=tx.Save(&account).Error;err !=nil{
return err
}

	transaction := models.Transaction{
		Username: username,
		Type:     "Deposit",
		Amount:   amount,
		Balance:  account.Balance,
	}
	if err :=tx.Create(&transaction).Error;err !=nil{
return err
	}
	log.Printf("Successfully deposited %d to %s",amount,username)
	return nil
})
}

func GetTransactions(username string) ([]models.Transaction, error) {
	var history []models.Transaction
	if err:=db.DB.Where("username=?",username).Order("created_at desc").Find(&history).Error;err !=nil{
		return nil,err
	}

	return history, nil
}

func CreateAccountProcess(username string) (models.Account, error) {
var account models.Account

err:=db.DB.Where("username=?",username).First(&account).Error

if err ==nil{
	return models.Account{},errors.New("Account already exists")
}

	account = models.Account{
		Username: username,
		Balance:  0,
		Active:   true,
	}
if err :=db.DB.Create(&account).Error;err !=nil{
	return models.Account{},err
}
return account,nil
}

func GetAccountsProcess() []string {
	var names []string
	err:=db.DB.Model(&models.Account{}).Pluck("username",&names).Error

	if err !=nil{
		log.Printf("Error fetching accounts: %v",err)
		return []string{}
	}
	return names
}

func DeleteAccountProcess(username string) error {
result:=db.DB.Model(&models.Account{}).Where("username=?",username).Update("active",false)

if result.Error !=nil{
	return result.Error
}


if result.RowsAffected==0{
	return errors.New("Account not found")
}
	log.Printf("Account %s has been deactivated",username)
	return nil
}

func GetBalanceProcess(username string) (models.Account, error) {
var account models.Account
if err:=db.DB.Where("username=?",username).First(&account).Error;err !=nil{
	return models.Account{},errors.New("Acocunt not found")
}
	return account, nil
}
