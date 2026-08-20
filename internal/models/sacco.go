package models

import (
	"time"

	"gorm.io/gorm"
)

// ShareCapital stores a member's non-spendable ownership balance.
type ShareCapital struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance   int64          `gorm:"not null;default:0" json:"balance"`
	Active    bool           `gorm:"not null;default:true" json:"active"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
}

// ShareContribution is the immutable ledger of administrator-recorded capital.
type ShareContribution struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	ShareCapitalID  uint           `gorm:"index;not null" json:"share_capital_id"`
	UserID          uint           `gorm:"index;not null" json:"user_id"`
	Amount          int64          `gorm:"not null" json:"amount"`
	Balance         int64          `gorm:"not null" json:"balance"`
	ReferenceNumber string         `gorm:"uniqueIndex;not null" json:"reference_number"`
	RecordedBy      string         `gorm:"index;not null" json:"recorded_by"`
	Note            string         `json:"note"`
}

type Loan struct {
	ID                   uint              `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	DeletedAt            gorm.DeletedAt    `gorm:"index" json:"-"`
	UserID               uint              `gorm:"index;not null" json:"user_id"`
	ReferenceNumber      string            `gorm:"uniqueIndex;not null" json:"reference_number"`
	Purpose              string            `gorm:"not null" json:"purpose"`
	RequestedPrincipal   int64             `gorm:"not null" json:"requested_principal"`
	ApprovedPrincipal    int64             `gorm:"not null;default:0" json:"approved_principal"`
	TotalRepayable       int64             `gorm:"not null;default:0" json:"total_repayable"`
	OutstandingPrincipal int64             `gorm:"not null;default:0" json:"outstanding_principal"`
	OutstandingInterest  int64             `gorm:"not null;default:0" json:"outstanding_interest"`
	TermMonths           int               `gorm:"not null" json:"term_months"`
	Status               string            `gorm:"index;not null;default:'pending'" json:"status"`
	DecisionNote         string            `json:"decision_note"`
	DecidedBy            string            `json:"decided_by"`
	DecidedAt            *time.Time        `json:"decided_at"`
	DisbursedBy          string            `json:"disbursed_by"`
	DisbursedAt          *time.Time        `json:"disbursed_at"`
	User                 User              `json:"user" gorm:"foreignKey:UserID"`
	Installments         []LoanInstallment `json:"installments,omitempty"`
}

type LoanInstallment struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	LoanID        uint           `gorm:"index;not null" json:"loan_id"`
	InstallmentNo int            `gorm:"not null" json:"installment_no"`
	DueDate       time.Time      `gorm:"index;not null" json:"due_date"`
	PrincipalDue  int64          `gorm:"not null" json:"principal_due"`
	InterestDue   int64          `gorm:"not null" json:"interest_due"`
	AmountPaid    int64          `gorm:"not null;default:0" json:"amount_paid"`
	Status        string         `gorm:"index;not null;default:'pending'" json:"status"`
}

type LoanRepayment struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	LoanID          uint           `gorm:"index;not null" json:"loan_id"`
	Amount          int64          `gorm:"not null" json:"amount"`
	PrincipalPaid   int64          `gorm:"not null" json:"principal_paid"`
	InterestPaid    int64          `gorm:"not null" json:"interest_paid"`
	ReferenceNumber string         `gorm:"uniqueIndex;not null" json:"reference_number"`
	RecordedBy      string         `gorm:"index;not null" json:"recorded_by"`
}
