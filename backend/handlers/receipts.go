package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"fintech-labs/backend/services"
	"fintech-labs/backend/utils"
)

func ReceiptHandler(w http.ResponseWriter, r *http.Request) {
	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract reference number from URL path /receipt/AV-2026-00000001
	path := strings.TrimPrefix(r.URL.Path, "/receipt/")
	path = strings.TrimSpace(path)
	log.Printf("DEBUG receipt path: '%s'", path)

	if path == "" {
		http.Redirect(w, r, "/dashboard?error=Invalid+receipt+reference", http.StatusSeeOther)
		return
	}

	transaction, err := services.GetTransactionByReference(path)
	if err != nil {
		http.Redirect(w, r, "/dashboard?error=Receipt+not+found", http.StatusSeeOther)
		return
	}

	// Security check — users can only view their own receipts
	if transaction.Username != username {
		http.Redirect(w, r, "/dashboard?error=Unauthorized", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("receipt.html").Funcs(template.FuncMap{
		"formatKES":  utils.FormatKES,
		"formatDate": utils.FormatDate,
	}).ParseFiles("frontend/templates/receipt.html")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		ReferenceNumber string
		Type            string
		Amount          int64
		Balance         int64
		AccountNumber   string
		Date            string
		Status          string
	}{
		ReferenceNumber: transaction.ReferenceNumber,
		Type:            strings.Title(transaction.Type),
		Amount:          transaction.Amount,
		Balance:         transaction.Balance,
		AccountNumber:   transaction.AccountNumber,
		Date:            utils.FormatDate(transaction.CreatedAt),
		Status:          strings.Title(transaction.Status),
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
