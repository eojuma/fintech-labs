package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Username            string         `gorm:"not null" json:"username"`
	Password            string         `gorm:"not null" json:"-"`
	Role                string         `gorm:"default:'customer'" json:"role"` // "customer" or "admin"
	Accounts            []Account      `json:"accounts,omitempty"`
	Email               string         `gorm:"uniqueIndex;not null" json:"email"`
	FullName            string         `gorm:"not null" json:"fullname"`
	PhoneNumber         string         `gorm:"uniqueIndex;not null" json:"phonenumber"`
	NationlID           string         `gorm:"uniqueIndex;not null" json:"national_id"`
	FailedLoginAttempts int            `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time     `gorm:"default:null" json:"locked_until"`
	TransactionPin      string         `gorm:"default:''" json:"-"`
}

type Account struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	UserID      uint           `json:"user_id"`
	Number      string         `gorm:"uniqueIndex" json:"number"`
	Balance     int64          `json:"balance"`
	Active      bool           `gorm:"default:true" json:"active"`
	AccountType string         `gorm:"default:'current'" json:"account_type"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
}

type Transaction struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Username        string         `gorm:"index" json:"username"`
	AccountNumber   string         `gorm:"index;default:''" json:"account_number"`
	Type            string         `json:"type"`
	Amount          int64          `json:"amount"` // amounts stored in whole numbers
	Balance         int64          `json:"balance"`
	ReferenceNumber string         `gorm:"uniqueIndex;default:''" json:"reference_number"`
	// Mpesa integration
	MpesaReceiptCode  string `gorm:"uniqueIndex;default:null" json:"mpesa_receipt_code,omitempty"`
	MpesaPhoneNumber  string `json:"mpesa_phone_number,omitempty"`
	MerchantRequestID string `gorm:"index;default:null" json:"merchant_request_id,omitempty"`       // For tracking STK Push
	CheckoutRequestID string `gorm:"uniqueIndex;default:null" json:"checkout_request_id,omitempty"` // For tracking STK Push
	Status            string `gorm:"default:'pending'" json:"status"`                               // "pending", "completed", "failed"
}

type DepositRequest struct {
	Amount int64 `json:"amount"`
}

type WithdrawRequest struct {
	Amount int64 `json:"amount"`
}

type TransferRecipient struct {
	AccountNumber string `json:"account_number"`
	Amount        int64  `json:"amount"`
}

type MultiTransferRequest struct {
	SenderIdentifier string              `json:"sender_identifier"`
	Recipients       []TransferRecipient `json:"recipients"`
}

type MpesaDepositRequest struct {
	Amount        int64  `json:"amount" binding:"required"`
	PhoneNumber   string `json:"phone_number" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
}

type Session struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	Token          string         `gorm:"uniqueIndex;not null" json:"token"`
	LastActivityAt time.Time      `gorm:"not null" json:"last_activity_at"`
	User           User           `json:"user" gorm:"foreignKey:UserID"`
}

type StatementData struct {
	AccountHolderName string
	AccountNumber     string
	AccountType       string
	From              time.Time
	To                time.Time
	OpeningBalance    int64
	ClosingBalance    int64
	Transactions      []Transaction
}

type TransactionFilter struct {
	AccountNumber string
	Type          string
	From          string
	To            string
	MinAmount     int64
	MaxAmount     int64
	SortOrder     string // "desc" or "asc"
	Page          int
	Limit         int
}

type FilterResult struct {
	Transactions     []Transaction
	TotalCount       int64
	TotalDeposits    int64
	TotalWithdrawals int64
	Page             int
	Limit            int
	TotalPages       int
}