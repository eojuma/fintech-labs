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

const (
	MinDeposit    = 50
	MinWithdrawal = 100
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

	if amount < MinWithdrawal {
		return models.Account{}, errors.New("Minimun withdrawal is ksh.100")
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
	transactions[username] = append(transactions[username], history)
	return account, nil
}

func DepositProcess(username string, amount int64) (models.Account, error) {
	account, exists := accounts[username]

	if !exists {
		return models.Account{}, errors.New("Account not found")
	}
	if !account.Active {
		return models.Account{}, errors.New("Account inactive")
	}
	if amount < MinDeposit {
		return models.Account{}, errors.New("Minimum withdrawal is ksh.50")
	}

	account.Balance += amount
	accounts[username] = account

	history := models.Transaction{
		Username: username,
		Type:     "Deposit",
		Amount:   amount,
		Balance:  account.Balance,
		Time:     time.Now().UTC(),
	}
	transactions[username] = append(transactions[username], history)
	return account, nil
}

func GetTransactions(username string) ([]models.Transaction, error) {
	if _, exists := accounts[username]; !exists {
		return nil, errors.New("account not found")
	}

	history := transactions[username]

	if history == nil {
		return []models.Transaction{}, nil
	}

	return history, nil
}


func CreateAccountProcess(username string) (models.Account, error) {
    if _, exists := accounts[username]; exists {
        return models.Account{}, errors.New("account already exists")
    }

    
    account := models.Account{
        Username: username,
        Balance:  0,
        Active:   true,
    }

    
    accounts[username] = account
    return account, nil
}

func GetAccountsProcess() []string {
    var names []string
    for name := range accounts {
        names = append(names, name)
    }
    return names
}

func DeleteAccountProcess(username string) error {
    account, exists := accounts[username]
    if !exists {
        return errors.New("account not found")
    }
    
    account.Active = false
    accounts[username] = account
    return nil
}

func GetBalanceProcess(username string) (models.Account, error) {
    account, exists := accounts[username]
    if !exists {
        return models.Account{}, errors.New("account not found")
    }
    return account, nil
}