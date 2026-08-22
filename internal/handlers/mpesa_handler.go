package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"fintech-labs/internal/mpesa"
	"fintech-labs/internal/services"
	"fintech-labs/internal/utils"
)

func MpesaDepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := utils.GetSessionUser(w, r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	accountNumber := r.FormValue("account_number")
	phoneNumber := r.FormValue("phone_number")
	amountStr := r.FormValue("amount")

	if accountNumber == "" || phoneNumber == "" || amountStr == "" {
		http.Redirect(w, r, "/dashboard?error=All+fields+required", http.StatusSeeOther)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amount < 1 {
		http.Redirect(w, r, "/dashboard?error=Invalid+amount", http.StatusSeeOther)
		return
	}

	// Normalize phone number to 254XXXXXXXXX format
	phone := utils.FormatPhoneForSMS(phoneNumber)
	// Remove + prefix if present — M-Pesa expects 254XXXXXXXXX not +254XXXXXXXXX
	phone = strings.TrimPrefix(phone, "+")

	log.Printf("📱 Initiating M-Pesa STK Push for %s — KES %d to account %s", phone, amount, accountNumber)

	stkResp, err := mpesa.InitiateSTKPush(phone, accountNumber, amount)
	if err != nil {
		log.Printf("❌ STK Push failed: %v", err)
		errorMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, "/dashboard?error="+errorMsg, http.StatusSeeOther)
		return
	}

	// Save pending transaction to database
	if err := services.CreatePendingMpesaTransaction(username, accountNumber, phone, amount, stkResp.CheckoutRequestID, stkResp.MerchantRequestID); err != nil {
		log.Printf("⚠️ Failed to save pending M-Pesa transaction: %v", err)
	}

	log.Printf("✅ STK Push initiated — CheckoutRequestID: %s", stkResp.CheckoutRequestID)
	http.Redirect(w, r, "/dashboard?success=M-Pesa+prompt+sent!+Enter+your+PIN+on+your+phone+to+complete+the+deposit", http.StatusSeeOther)
}

func MpesaCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read M-Pesa callback body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("📨 M-Pesa callback received: %s", string(body))

	callback, err := mpesa.ParseCallback(body)
	if err != nil {
		log.Printf("❌ Failed to parse M-Pesa callback: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ResultCode 0 means success — can be int or string depending on sandbox/production
	resultCode := fmt.Sprintf("%v", callback.ResultCode)
	if resultCode != "0" {
		log.Printf("⚠️ M-Pesa payment failed — ResultCode: %v | Reason: %s", callback.ResultCode, callback.ResultDesc)
		services.FailMpesaTransaction(callback.CheckoutRequestID, callback.ResultDesc)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract payment details from callback metadata
	amount := mpesa.GetCallbackValue(callback.CallbackMetadata.Item, "Amount")
	receiptCode := mpesa.GetCallbackValue(callback.CallbackMetadata.Item, "MpesaReceiptNumber")
	phoneNumber := mpesa.GetCallbackValue(callback.CallbackMetadata.Item, "PhoneNumber")

	// Convert phone from float scientific notation to clean integer string
	phoneStr := fmt.Sprintf("%.0f", phoneNumber)

	log.Printf("✅ M-Pesa payment confirmed — Receipt: %v | Amount: %v | Phone: %s", receiptCode, amount, phoneStr)

	// Process the confirmed payment
	if err := services.ProcessMpesaDeposit(
		callback.CheckoutRequestID,
		fmt.Sprintf("%v", receiptCode),
		fmt.Sprintf("%v", amount),
	); err != nil {
		log.Printf("❌ Failed to process M-Pesa deposit: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Respond with success to Safaricom
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"ResultCode": "0",
		"ResultDesc": "Accepted",
	})
}
