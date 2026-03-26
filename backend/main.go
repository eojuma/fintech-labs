package main

import (
	"fmt"
	"net/http"

	"fintech-labs/handlers"
)

func main() {
	http.HandleFunc("/account", handlers.CreateAccount)
	http.HandleFunc("/deposit", handlers.Deposits)
	http.HandleFunc("/withdraw", handlers.Withdrawals)
	http.HandleFunc("/balance", handlers.Balances)
	http.HandleFunc("/transactions", handlers.Transactions)
	http.HandleFunc("/accounts", handlers.GetAccounts)
	http.HandleFunc("/delete", handlers.Delete)
	fmt.Println("Server running on http://8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("The server is down", err)
	}
}
