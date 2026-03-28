package main

import (
	"log"
	"net/http"

	"fintech-labs/db"
	"fintech-labs/handlers"
)

func main() {
	db.InitDB()
	http.HandleFunc("/account", handlers.CreateAccount)
	http.HandleFunc("/deposit", handlers.Deposits)
	http.HandleFunc("/withdraw", handlers.Withdrawals)
	http.HandleFunc("/balance", handlers.Balances)
	http.HandleFunc("/transactions", handlers.Transactions)
	http.HandleFunc("/accounts", handlers.GetAccounts)
	http.HandleFunc("/deactivate", handlers.Deactivate)
	http.HandleFunc("/reactivate", handlers.Reactivate)
	log.Println("Server running on http://8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("The server is down", err)
	}
}
