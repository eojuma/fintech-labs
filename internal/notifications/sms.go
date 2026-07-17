package notifications

import (
	"fmt"
	"os"

	"github.com/AndroidStudyOpenSource/africastalking-go/sms"
)

func SendTransactionSMS(phoneNumber, message string) error {
	username := os.Getenv("AT_USERNAME")
	apiKey := os.Getenv("AT_API_KEY")

	if username == "" || apiKey == "" {
		return fmt.Errorf("Africa's Talking credentials not configured")
	}

	smsService := sms.NewService(username, apiKey, "sandbox")

	response, err := smsService.Send("", phoneNumber, message)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	fmt.Printf("SMS response: %+v\n", response)
	return nil
}

func FormatSMSMessage(transactionType, accountNumber, referenceNumber string, amount, balance int64) string {
	var action string
	switch transactionType {
	case "deposited":
		action = fmt.Sprintf("💰 KES %s has been credited to your account", formatAmount(amount))
	case "withdrawn":
		action = fmt.Sprintf("💸 KES %s has been debited from your account", formatAmount(amount))
	case "transferred out":
		action = fmt.Sprintf("📤 KES %s has been sent from your account", formatAmount(amount))
	case "received":
		action = fmt.Sprintf("📥 KES %s has been received in your account", formatAmount(amount))
	default:
		action = fmt.Sprintf("KES %s transaction on your account", formatAmount(amount))
	}

	return fmt.Sprintf(
		"[African Vault] %s %s. Avail Bal: KES %s. Ref: %s. Not you? Call support immediately. *Charges may apply*",
		action,
		accountNumber,
		formatAmount(balance),
		referenceNumber,
	)
}

func formatAmount(amount int64) string {
	return fmt.Sprintf("%d", amount)
}