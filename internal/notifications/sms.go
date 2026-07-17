package notifications

import (
	"fmt"
	"os"

	"github.com/AfricasTalkingLtd/africastalking-go/sms"
)

func SendTransactionSMS(phoneNumber, message string) error {
	username := os.Getenv("AT_USERNAME")
	apiKey := os.Getenv("AT_API_KEY")

	if username == "" || apiKey == "" {
		return fmt.Errorf("Africa's Talking credentials not configured")
	}

	smsService := sms.NewService(username, apiKey, "")

	response, err := smsService.Send("", []string{phoneNumber}, message)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	if len(response.SMSMessageData.Recipients) == 0 {
		return fmt.Errorf("no recepients in SMS response")
	}

	recipient := response.SMSMessageData.Recipients[0]

	if recipient.Status != "Success" {
		return fmt.Errorf("SMS failed: %s", &recipient.Status)
	}
	return nil
}

func FormatSMSMessage(transactionType, accountNumber, referenceNumber string, amount, balance int64) string {
	return fmt.Sprintf(
		"African Vault: KES %d %s on account %s. New balance: KES %d. Ref: %s. Not you? Contact support immediately.",
		amount,
		transactionType,
		accountNumber,
		balance,
		referenceNumber,
	)
}