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

type ShareRedemption struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
	ShareCapitalID           uint           `gorm:"index;not null" json:"share_capital_id"`
	UserID                   uint           `gorm:"index;not null" json:"user_id"`
	Amount                   int64          `gorm:"not null" json:"amount"`
	Balance                  int64          `gorm:"not null" json:"balance"`
	ReferenceNumber          string         `gorm:"uniqueIndex;not null" json:"reference_number"`
	DestinationAccountNumber string         `gorm:"not null" json:"destination_account_number"`
	RecordedBy               string         `gorm:"index;not null" json:"recorded_by"`
	Reason                   string         `gorm:"not null" json:"reason"`
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

type LoanEligibilityPolicy struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	SavingsMultiple      float64   `gorm:"not null;default:0" json:"savings_multiple"`
	ShareCapitalMultiple float64   `gorm:"not null;default:0" json:"share_capital_multiple"`
	MinimumShareCapital  int64     `gorm:"not null;default:0" json:"minimum_share_capital"`
	Active               bool      `gorm:"not null;default:true" json:"active"`
	UpdatedBy            string    `gorm:"not null" json:"updated_by"`
}

type DistributionPolicy struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	SavingsInterestRate     float64   `gorm:"not null;default:6" json:"savings_interest_rate"`
	ShareDividendRate       float64   `gorm:"not null;default:10" json:"share_dividend_rate"`
	SavingsWithholdingRate  float64   `gorm:"not null;default:15" json:"savings_withholding_rate"`
	DividendWithholdingRate float64   `gorm:"not null;default:5" json:"dividend_withholding_rate"`
	BalanceBasis            string    `gorm:"not null;default:'monthly_average'" json:"balance_basis"`
	AutoPreviewEnabled      bool      `gorm:"not null;default:false" json:"auto_preview_enabled"`
	Active                  bool      `gorm:"not null;default:true" json:"active"`
	UpdatedBy               string    `gorm:"not null" json:"updated_by"`
}

type DistributionRun struct {
	ID                uint                     `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	Type              string                   `gorm:"uniqueIndex:idx_distribution_period;not null" json:"type"`
	Period            string                   `gorm:"uniqueIndex:idx_distribution_period;not null" json:"period"`
	Rate              float64                  `gorm:"not null" json:"rate"`
	Status            string                   `gorm:"index;not null;default:'preview'" json:"status"`
	TotalAmount       int64                    `gorm:"not null;default:0" json:"total_amount"`
	CreatedBy         string                   `gorm:"not null" json:"created_by"`
	ApprovedBy        string                   `json:"approved_by"`
	ApprovedAt        *time.Time               `json:"approved_at"`
	ApprovalReference string                   `json:"approval_reference"`
	PostedBy          string                   `json:"posted_by"`
	PostedAt          *time.Time               `json:"posted_at"`
	Allocations       []DistributionAllocation `json:"allocations,omitempty"`
}

type DistributionAllocation struct {
	ID                       uint      `gorm:"primaryKey" json:"id"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	DistributionRunID        uint      `gorm:"uniqueIndex:idx_distribution_member;not null" json:"distribution_run_id"`
	UserID                   uint      `gorm:"uniqueIndex:idx_distribution_member;not null" json:"user_id"`
	BasisAmount              int64     `gorm:"not null" json:"basis_amount"`
	GrossAmount              int64     `gorm:"not null;default:0" json:"gross_amount"`
	WithholdingAmount        int64     `gorm:"not null;default:0" json:"withholding_amount"`
	Amount                   int64     `gorm:"not null" json:"amount"`
	DestinationAccountNumber string    `json:"destination_account_number"`
	ReferenceNumber          string    `gorm:"uniqueIndex;default:null" json:"reference_number"`
	Status                   string    `gorm:"index;not null;default:'pending'" json:"status"`
	User                     User      `json:"user" gorm:"foreignKey:UserID"`
}
