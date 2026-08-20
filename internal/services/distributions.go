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
	DefaultSavingsInterestRate     = 6.0
	DefaultShareDividendRate       = 10.0
	DefaultSavingsWithholdingRate  = 15.0
	DefaultDividendWithholdingRate = 5.0
)

func GetDistributionPolicy() (*models.DistributionPolicy, error) {
	var policy models.DistributionPolicy
	err := db.DB.Where("active = ?", true).Order("id desc").First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.DistributionPolicy{SavingsInterestRate: DefaultSavingsInterestRate, ShareDividendRate: DefaultShareDividendRate, SavingsWithholdingRate: DefaultSavingsWithholdingRate, DividendWithholdingRate: DefaultDividendWithholdingRate, BalanceBasis: "monthly_average", Active: true}, nil
	}
	return &policy, err
}

func SetDistributionPolicy(adminUsername string, savingsRate, dividendRate, savingsWithholding, dividendWithholding float64, autoPreview bool) error {
	if savingsRate < 0 || savingsRate > 100 || dividendRate < 0 || dividendRate > 100 || savingsWithholding < 0 || savingsWithholding > 100 || dividendWithholding < 0 || dividendWithholding > 100 {
		return errors.New("distribution rates must be between 0 and 100 percent")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DistributionPolicy{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}
		return tx.Create(&models.DistributionPolicy{SavingsInterestRate: savingsRate, ShareDividendRate: dividendRate, SavingsWithholdingRate: savingsWithholding, DividendWithholdingRate: dividendWithholding, BalanceBasis: "monthly_average", AutoPreviewEnabled: autoPreview, Active: true, UpdatedBy: adminUsername}).Error
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
	rate, withholdingRate := policy.SavingsInterestRate, policy.SavingsWithholdingRate
	if distributionType == "share_dividend" {
		rate = policy.ShareDividendRate
		withholdingRate = policy.DividendWithholdingRate
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
			basis, err := monthlyAverageBasis(tx, user.ID, distributionType, period)
			if err != nil {
				return err
			}
			gross := int64(math.Round(float64(basis) * rate / 100))
			withholding := int64(math.Round(float64(gross) * withholdingRate / 100))
			net := gross - withholding
			if net <= 0 {
				continue
			}
			allocation := models.DistributionAllocation{DistributionRunID: run.ID, UserID: user.ID, BasisAmount: basis, GrossAmount: gross, WithholdingAmount: withholding, Amount: net, Status: "pending"}
			if err := tx.Create(&allocation).Error; err != nil {
				return err
			}
			run.TotalAmount += net
		}
		return tx.Model(run).Update("total_amount", run.TotalAmount).Error
	})
	return run, err
}

func monthlyAverageBasis(tx *gorm.DB, userID uint, distributionType, period string) (int64, error) {
	year, _ := time.Parse("2006", period)
	var total int64
	for month := 1; month <= 12; month++ {
		monthEnd := time.Date(year.Year(), time.Month(month)+1, 1, 0, 0, 0, 0, time.Local)
		var balance int64
		if distributionType == "savings_interest" {
			var account models.Account
			err := tx.Where("user_id = ? AND account_type = ? AND created_at < ?", userID, "savings", monthEnd).First(&account).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return 0, err
			}
			var last models.Transaction
			err = tx.Where("account_number = ? AND created_at < ?", account.Number, monthEnd).Order("created_at desc").First(&last).Error
			if err == nil {
				balance = last.Balance
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, err
			}
		} else {
			var contribution models.ShareContribution
			err := tx.Where("user_id = ? AND created_at < ?", userID, monthEnd).Order("created_at desc").First(&contribution).Error
			if err == nil {
				balance = contribution.Balance
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, err
			}
		}
		total += balance
	}
	return int64(math.Round(float64(total) / 12)), nil
}

func ApproveDistribution(approver string, runID uint, reference string) error {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return errors.New("board or AGM approval reference is required")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var run models.DistributionRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&run, runID).Error; err != nil {
			return errors.New("distribution run not found")
		}
		if run.Status != "preview" {
			return errors.New("only preview runs can be approved")
		}
		if run.CreatedBy == approver {
			return errors.New("preview creator cannot approve the same distribution")
		}
		now := time.Now()
		run.Status, run.ApprovedBy, run.ApprovedAt, run.ApprovalReference = "approved", approver, &now, reference
		return tx.Save(&run).Error
	})
}

func PostDistribution(adminUsername string, runID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var run models.DistributionRun
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&run, runID).Error; err != nil {
			return errors.New("distribution run not found")
		}
		if run.Status != "approved" {
			return errors.New("only approved distribution runs can be posted")
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

func CreateScheduledDistributionPreviews(now time.Time) error {
	policy, err := GetDistributionPolicy()
	if err != nil || !policy.AutoPreviewEnabled || now.Month() != time.January {
		return err
	}
	period := fmt.Sprintf("%d", now.Year()-1)
	for _, kind := range []string{"savings_interest", "share_dividend"} {
		var count int64
		if err := db.DB.Model(&models.DistributionRun{}).Where("type = ? AND period = ?", kind, period).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if _, err := PreviewDistribution("system_scheduler", kind, period); err != nil {
				return err
			}
		}
	}
	return nil
}

func StartDistributionScheduler() {
	go func() {
		_ = CreateScheduledDistributionPreviews(time.Now())
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for now := range ticker.C {
			_ = CreateScheduledDistributionPreviews(now)
		}
	}()
}

func GetDistributionRuns() ([]models.DistributionRun, error) {
	var runs []models.DistributionRun
	err := db.DB.Preload("Allocations.User").Order("created_at desc").Find(&runs).Error
	return runs, err
}
