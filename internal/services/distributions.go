package services

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	DefaultSavingsInterestRate = 6.0
	DefaultShareDividendRate   = 10.0
)

func GetDistributionPolicy() (*models.DistributionPolicy, error) {
	var policy models.DistributionPolicy
	err := db.DB.Where("active = ?", true).Order("id desc").First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.DistributionPolicy{SavingsInterestRate: DefaultSavingsInterestRate, ShareDividendRate: DefaultShareDividendRate, Active: true}, nil
	}
	return &policy, err
}

func SetDistributionPolicy(adminUsername string, savingsRate, dividendRate float64) error {
	if savingsRate < 0 || savingsRate > 100 || dividendRate < 0 || dividendRate > 100 {
		return errors.New("distribution rates must be between 0 and 100 percent")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DistributionPolicy{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}
		return tx.Create(&models.DistributionPolicy{SavingsInterestRate: savingsRate, ShareDividendRate: dividendRate, Active: true, UpdatedBy: adminUsername}).Error
	})
}

func PreviewDistribution(adminUsername, distributionType, period string) (*models.DistributionRun, error) {
	distributionType, period = strings.TrimSpace(distributionType), strings.TrimSpace(period)
	if distributionType != "savings_interest" && distributionType != "share_dividend" {
		return nil, errors.New("invalid distribution type")
	}
	if _, err := time.Parse("2006", period); err != nil {
		return nil, errors.New("period must be a four-digit year")
	}
	policy, err := GetDistributionPolicy()
	if err != nil {
		return nil, err
	}
	var existing models.DistributionRun
	if err := db.DB.Where("type = ? AND period = ?", distributionType, period).First(&existing).Error; err == nil {
		return nil, errors.New("a distribution run already exists for this type and period")
	}
	rate := policy.SavingsInterestRate
	if distributionType == "share_dividend" {
		rate = policy.ShareDividendRate
	}
	run := &models.DistributionRun{Type: distributionType, Period: period, Rate: rate, Status: "preview", CreatedBy: adminUsername}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(run).Error; err != nil {
			return err
		}
		var users []models.User
		if err := tx.Where("role = ?", "customer").Find(&users).Error; err != nil {
			return err
		}
		for _, user := range users {
			var basis int64
			if distributionType == "savings_interest" {
				if err := tx.Model(&models.Account{}).Where("user_id = ? AND active = ? AND account_type = ?", user.ID, true, "savings").Select("COALESCE(SUM(balance), 0)").Scan(&basis).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&models.ShareCapital{}).Where("user_id = ? AND active = ?", user.ID, true).Select("COALESCE(SUM(balance), 0)").Scan(&basis).Error; err != nil {
					return err
				}
			}
			amount := int64(math.Round(float64(basis) * rate / 100))
			if amount <= 0 {
				continue
			}
			allocation := models.DistributionAllocation{DistributionRunID: run.ID, UserID: user.ID, BasisAmount: basis, Amount: amount, Status: "pending"}
			if err := tx.Create(&allocation).Error; err != nil {
				return err
			}
			run.TotalAmount += amount
		}
		return tx.Model(run).Update("total_amount", run.TotalAmount).Error
	})
	return run, err
}

func PostDistribution(adminUsername string, runID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var run models.DistributionRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&run, runID).Error; err != nil {
			return errors.New("distribution run not found")
		}
		if run.Status != "preview" {
			return errors.New("only preview distribution runs can be posted")
		}
		var allocations []models.DistributionAllocation
		if err := tx.Where("distribution_run_id = ?", run.ID).Find(&allocations).Error; err != nil {
			return err
		}
		for i := range allocations {
			accountType := "savings"
			if run.Type == "share_dividend" {
				accountType = "current"
			}
			var account models.Account
			if err := tx.Where("user_id = ? AND account_type = ? AND active = ?", allocations[i].UserID, accountType, true).First(&account).Error; err != nil {
				return fmt.Errorf("active %s account missing for member ID %d", accountType, allocations[i].UserID)
			}
			var user models.User
			if err := tx.First(&user, allocations[i].UserID).Error; err != nil {
				return err
			}
			account.Balance += allocations[i].Amount
			if err := tx.Save(&account).Error; err != nil {
				return err
			}
			ref := GenerateReferenceNumber()
			if err := tx.Create(&models.Transaction{Username: user.Username, AccountNumber: account.Number, Type: run.Type, Amount: allocations[i].Amount, Balance: account.Balance, ReferenceNumber: ref, Status: "completed"}).Error; err != nil {
				return err
			}
			allocations[i].DestinationAccountNumber, allocations[i].ReferenceNumber, allocations[i].Status = account.Number, ref, "posted"
			if err := tx.Save(&allocations[i]).Error; err != nil {
				return err
			}
		}
		now := time.Now()
		run.Status, run.PostedBy, run.PostedAt = "posted", adminUsername, &now
		return tx.Save(&run).Error
	})
}

func GetDistributionRuns() ([]models.DistributionRun, error) {
	var runs []models.DistributionRun
	err := db.DB.Preload("Allocations.User").Order("created_at desc").Find(&runs).Error
	return runs, err
}
