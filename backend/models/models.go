package models

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	UserID  uint   `json:"user_id"`
	Number  string `json:"number" gorm:"uniqueIndex"`
	Balance int64  `json:"balance"   gorm:"default:0"`
	Active  bool   `json:"active"   gorm:"default:true"`
}

type Deposit struct {
	Username string `json:"username"`
	Amount   int64  `json:"amount"`
}

type Withdrawal struct {
	Username string `json:"username"`
	Amount   int64  `json:"amount"`
}

type Transaction struct {
	gorm.Model
	Username string `json:"username"  gorm:"index"`
	Type     string `json:"type"`
	Amount   int64  `json:"amount"`
	Balance  int64  `json:"balance"`
}

type User struct {
	gorm.Model
	Username string    `json:"username" gorm:"uniqueIndex;not null"`
	Password string    `json:"-" gorm:"not null"`
	Role     string    `json:"role" gorm:"default:'customer'"`
	Accounts []Account `json:"accounts" gorm:"foreignKey:UserID"`
}
