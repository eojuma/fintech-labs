package main

type Account struct {
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
}

type Deposit struct {
	Username string  `json:"username"`
	Amount   float64 `json:"amount"`
}

type Withdrawal struct {
	Username string  `json:"username"`
	Amount   float64 `json:"amount"`
}

type Transaction struct {
	Username string  `json:"username"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Balance  float64 `json:"balance"`
	Time     string  `json:"time"`
}
