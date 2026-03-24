package models
import "time"

type Account struct {
	Username string `json:"username"`
	Balance  int64  `json:"balance"`
	Active   bool   `json:"active"`
}

type Deposit struct {
	Username string `json:"username"`
	Amount   int64  `json:"amount"`
}

type Withdrawal struct {
	Username string `json:"username"`
	Amount   int64  `json:"amount"`
}

type Transaction struct {
	Username string    `json:"username"`
	Type     string    `json:"type"`
	Amount   int64     `json:"amount"`
	Balance  int64     `json:"balance"`
	Time     time.Time `json:"time"`
}
