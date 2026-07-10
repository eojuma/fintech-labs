package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"

	// 1. Import your models package (adjust path to match your go.mod module name)
	"fintech-labs/internal/models"
)

// 2. Update the function signature to use models.TransactionEmailData
func SendTransactionEmail(toEmail string, data models.TransactionEmailData) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	tmpl, err := template.ParseFiles("web/templates/email.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	subject := fmt.Sprintf("African Vault — %s Alert", data.Type)
	body.Write([]byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n", toEmail, subject)))

	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	err = smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, body.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
