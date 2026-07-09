package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)

// TransactionEmailData holds the template placeholders
type TransactionEmailData struct {
	Type      string
	Amount    string
	Balance   string
	Timestamp string
}

// SendTransactionEmail builds and sends the notification email
func SendTransactionEmail(toEmail string, data TransactionEmailData) error {
	// 1. Setup SMTP Configuration (Use environment variables for safety)
	smtpHost := os.Getenv("SMTP_HOST") // e.g., "smtp.mailtrap.io" or "smtp.gmail.com"
	smtpPort := os.Getenv("SMTP_PORT") // e.g., "587"
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	// 2. Parse the HTML Template
	// Note: Path is relative to where the binary runs (usually project root)
	tmpl, err := template.ParseFiles("web/templates/email_template.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	// Write MIME headers so the email client renders it as HTML
	body.Write([]byte(fmt.Sprintf("To: %s\r\nSubject: Transaction Alert\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n", toEmail)))
	
	// Execute the template with the data
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 3. Authenticate and Send
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	err = smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, body.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}