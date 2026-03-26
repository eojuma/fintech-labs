package services

import (
	"errors"
	"time"

	"fintech-labs/models"
)

var (
	accounts     = make(map[string]models.Account)
	transactions = make(map[string][]models.Transaction)
)

func WithdrawalProcess(username string, amount int64) (models.Account, error) {
	account, exists := accounts[username]

	if !exists {
		return models.Account{}, errors.New("Account not found")
	}
	if !account.Active {
		return models.Account{}, errors.New("Account inactive")
	}
	if account.Balance < amount {
		return models.Account{}, errors.New("Insufficient balance")
	}

	account.Balance -= amount
	accounts[username] = account

	history := models.Transaction{
		Username: username,
		Type:     "Withdrawal",
		Amount:   amount,
		Balance:  account.Balance,
		Time:     time.Now().UTC(),
	}
	transactions[username]=append(transactions[username],history)
	return account,nil
}
