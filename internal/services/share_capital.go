package services

import (
	"errors"
	"fmt"
	"strings"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecordShareContribution credits non-spendable member share capital and
// writes a contribution ledger entry in the same database transaction.
func RecordShareContribution(adminUsername, memberUsername string, amount int64, note string) (*models.ShareContribution, error) {
	adminUsername = strings.ToLower(strings.TrimSpace(adminUsername))
	memberUsername = strings.ToLower(strings.TrimSpace(memberUsername))
	note = strings.TrimSpace(note)
	if amount <= 0 {
		return nil, errors.New("share contribution must be greater than zero")
	}

	var contribution models.ShareContribution
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var member models.User
		if err := tx.Where("username = ?", memberUsername).First(&member).Error; err != nil {
			return errors.New("member not found")
		}
		if member.Role != "customer" {
			return errors.New("share capital can only be recorded for members")
		}

		var capital models.ShareCapital
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", member.ID).First(&capital).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			capital = models.ShareCapital{UserID: member.ID, Active: true}
			if err := tx.Create(&capital).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if !capital.Active {
			return errors.New("member share capital is inactive")
		}

		capital.Balance += amount
		if err := tx.Save(&capital).Error; err != nil {
			return err
		}
		var contributionCount int64
		if err := tx.Model(&models.ShareContribution{}).Count(&contributionCount).Error; err != nil {
			return err
		}
		contribution = models.ShareContribution{
			ShareCapitalID:  capital.ID,
			UserID:          member.ID,
			Amount:          amount,
			Balance:         capital.Balance,
			ReferenceNumber: fmt.Sprintf("SC-%08d", contributionCount+1),
			RecordedBy:      adminUsername,
			Note:            note,
		}
		return tx.Create(&contribution).Error
	})
	if err != nil {
		CreateAuditLog(adminUsername, "share_contribution", memberUsername, err.Error(), amount, "failed")
		return nil, err
	}
	CreateAuditLog(adminUsername, "share_contribution", memberUsername, fmt.Sprintf("Recorded share contribution %s", contribution.ReferenceNumber), amount, "success")
	return &contribution, nil
}

func GetShareCapitalByUsername(username string) (*models.ShareCapital, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var member models.User
	if err := db.DB.Where("username = ?", username).First(&member).Error; err != nil {
		return nil, errors.New("member not found")
	}
	var capital models.ShareCapital
	err := db.DB.Where("user_id = ?", member.ID).First(&capital).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.ShareCapital{UserID: member.ID, Balance: 0, Active: true}, nil
	}
	return &capital, err
}
