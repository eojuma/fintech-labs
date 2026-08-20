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
