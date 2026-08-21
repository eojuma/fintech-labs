package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ApplyForLoan(username, purpose string, principal int64, termMonths int) (*models.Loan, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	purpose = strings.TrimSpace(purpose)
	if principal <= 0 || termMonths < 1 || termMonths > 60 || purpose == "" {
		return nil, errors.New("principal, purpose, and a term of 1 to 60 months are required")
	}
	var member models.User
	if err := db.DB.Where("username = ?", username).First(&member).Error; err != nil || member.Role != "customer" {
		return nil, errors.New("member not found")
	}
	eligibility, err := EvaluateLoanEligibility(member.ID, principal)
	if err != nil {
		return nil, err
	}
	if !eligibility.Eligible {
		return nil, errors.New(eligibility.Reason)
	}
	var open int64
	db.DB.Model(&models.Loan{}).Where("user_id = ? AND status IN ?", member.ID, []string{"pending", "approved", "disbursed"}).Count(&open)
	if open > 0 {
		return nil, errors.New("you already have an open loan application or loan")
	}
	var count int64
	db.DB.Model(&models.Loan{}).Count(&count)
	loan := &models.Loan{UserID: member.ID, ReferenceNumber: fmt.Sprintf("LN-%08d", count+1), Purpose: purpose, RequestedPrincipal: principal, TermMonths: termMonths, Status: "pending"}
	return loan, db.DB.Create(loan).Error
}

type LoanEligibilityResult struct {
	Eligible     bool
	Limit        int64
	Savings      int64
	ShareCapital int64
	Reason       string
}

func EvaluateLoanEligibility(userID uint, requestedPrincipal int64) (*LoanEligibilityResult, error) {
	policy, err := GetLoanEligibilityPolicy()
	if err != nil {
		return nil, err
	}
	var savings int64
	if err := db.DB.Model(&models.Account{}).Where("user_id = ? AND active = ? AND account_type = ?", userID, true, "savings").Select("COALESCE(SUM(balance), 0)").Scan(&savings).Error; err != nil {
		return nil, err
	}
	var capital models.ShareCapital
	err = db.DB.Where("user_id = ? AND active = ?", userID, true).First(&capital).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		capital.Balance = 0
	} else if err != nil {
		return nil, err
	}
	result := &LoanEligibilityResult{Savings: savings, ShareCapital: capital.Balance}
	result.Limit = int64(float64(savings)*policy.SavingsMultiple + float64(capital.Balance)*policy.ShareCapitalMultiple)
	if capital.Balance < policy.MinimumShareCapital {
		result.Reason = fmt.Sprintf("minimum share capital of KES %d is required; current share capital is KES %d", policy.MinimumShareCapital, capital.Balance)
		return result, nil
	}
	if requestedPrincipal > result.Limit {
		result.Reason = fmt.Sprintf("requested principal exceeds your eligibility limit of KES %d", result.Limit)
		return result, nil
	}
	result.Eligible = true
	result.Reason = fmt.Sprintf("eligible up to KES %d", result.Limit)
	return result, nil
}

func GetLoanEligibilityPolicy() (*models.LoanEligibilityPolicy, error) {
	var policy models.LoanEligibilityPolicy
	if err := db.DB.Where("active = ?", true).Order("id desc").First(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("loan eligibility policy has not been configured")
		}
		return nil, err
	}
	if policy.SavingsMultiple <= 0 && policy.ShareCapitalMultiple <= 0 {
		return nil, errors.New("loan eligibility policy must include a savings or share capital multiple")
	}
	return &policy, nil
}

func SetLoanEligibilityPolicy(adminUsername string, savingsMultiple, shareMultiple float64, minimumShareCapital int64) error {
	if savingsMultiple < 0 || shareMultiple < 0 || minimumShareCapital < 0 {
		return errors.New("eligibility policy values cannot be negative")
	}
	if savingsMultiple == 0 && shareMultiple == 0 {
		return errors.New("at least one eligibility multiple must be greater than zero")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.LoanEligibilityPolicy{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}
		return tx.Create(&models.LoanEligibilityPolicy{SavingsMultiple: savingsMultiple, ShareCapitalMultiple: shareMultiple, MinimumShareCapital: minimumShareCapital, Active: true, UpdatedBy: adminUsername}).Error
	})
}

func GetMemberLoans(username string) ([]models.Loan, error) {
	var loans []models.Loan
	err := db.DB.Joins("JOIN users ON users.id = loans.user_id").Where("users.username = ?", strings.ToLower(strings.TrimSpace(username))).Preload("Installments", func(tx *gorm.DB) *gorm.DB { return tx.Order("installment_no asc") }).Order("loans.created_at desc").Find(&loans).Error
	return loans, err
}

func GetAllLoans() ([]models.Loan, error) {
	var loans []models.Loan
	err := db.DB.Preload("User").Preload("Installments", func(tx *gorm.DB) *gorm.DB { return tx.Order("installment_no asc") }).Order("created_at desc").Find(&loans).Error
	return loans, err
}

func DecideLoan(adminUsername string, loanID uint, action string, approvedPrincipal, totalRepayable int64, note string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var loan models.Loan
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&loan, loanID).Error; err != nil {
			return errors.New("loan not found")
		}
		if loan.Status != "pending" {
			return errors.New("only pending loans can be decided")
		}
		now := time.Now()
		loan.DecidedBy, loan.DecidedAt, loan.DecisionNote = adminUsername, &now, strings.TrimSpace(note)
		switch action {
		case "reject":
			loan.Status = "rejected"
		case "approve":
			if approvedPrincipal <= 0 || approvedPrincipal > loan.RequestedPrincipal || totalRepayable < approvedPrincipal {
				return errors.New("approved principal must be positive and total repayable cannot be below principal")
			}
			loan.Status, loan.ApprovedPrincipal, loan.TotalRepayable = "approved", approvedPrincipal, totalRepayable
		default:
			return errors.New("invalid loan decision")
		}
		return tx.Save(&loan).Error
	})
}

func DisburseLoan(adminUsername string, loanID uint) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var loan models.Loan
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&loan, loanID).Error; err != nil {
			return errors.New("loan not found")
		}
		if loan.Status != "approved" {
			return errors.New("only approved loans can be disbursed")
		}
		var account models.Account
		if err := tx.Where("user_id = ? AND account_type = ? AND active = ?", loan.UserID, "current", true).First(&account).Error; err != nil {
			return errors.New("member current account not found or inactive")
		}
		var member models.User
		if err := tx.First(&member, loan.UserID).Error; err != nil {
			return err
		}
		account.Balance += loan.ApprovedPrincipal
		if err := tx.Save(&account).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.Transaction{Username: member.Username, AccountNumber: account.Number, Type: "loan_disbursement", Amount: loan.ApprovedPrincipal, Balance: account.Balance, ReferenceNumber: GenerateReferenceNumber(), Status: "completed"}).Error; err != nil {
			return err
		}
		now := time.Now()
		loan.Status, loan.DisbursedBy, loan.DisbursedAt = "disbursed", adminUsername, &now
		loan.OutstandingPrincipal = loan.ApprovedPrincipal
		loan.OutstandingInterest = loan.TotalRepayable - loan.ApprovedPrincipal
		if err := tx.Save(&loan).Error; err != nil {
			return err
		}
		principalBase, principalRemainder := loan.ApprovedPrincipal/int64(loan.TermMonths), loan.ApprovedPrincipal%int64(loan.TermMonths)
		interestTotal := loan.TotalRepayable - loan.ApprovedPrincipal
		interestBase, interestRemainder := interestTotal/int64(loan.TermMonths), interestTotal%int64(loan.TermMonths)
		for i := 1; i <= loan.TermMonths; i++ {
			principalDue, interestDue := principalBase, interestBase
			if int64(i) <= principalRemainder {
				principalDue++
			}
			if int64(i) <= interestRemainder {
				interestDue++
			}
			installment := models.LoanInstallment{LoanID: loan.ID, InstallmentNo: i, DueDate: now.AddDate(0, i, 0), PrincipalDue: principalDue, InterestDue: interestDue, Status: "pending"}
			if err := tx.Create(&installment).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func RecordLoanRepayment(adminUsername string, loanID uint, amount int64) error {
	if amount <= 0 {
		return errors.New("repayment must be greater than zero")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var loan models.Loan
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&loan, loanID).Error; err != nil {
			return errors.New("loan not found")
		}
		if loan.Status != "disbursed" {
			return errors.New("only disbursed loans accept repayments")
		}
		outstanding := loan.OutstandingPrincipal + loan.OutstandingInterest
		if amount > outstanding {
			return errors.New("repayment exceeds outstanding balance")
		}
		remaining, interestPaid := amount, amount
		if interestPaid > loan.OutstandingInterest {
			interestPaid = loan.OutstandingInterest
		}
		loan.OutstandingInterest -= interestPaid
		remaining -= interestPaid
		principalPaid := remaining
		loan.OutstandingPrincipal -= principalPaid
		var installments []models.LoanInstallment
		if err := tx.Where("loan_id = ? AND status != ?", loan.ID, "paid").Order("installment_no asc").Find(&installments).Error; err != nil {
			return err
		}
		allocation := amount
		for i := range installments {
			due := installments[i].PrincipalDue + installments[i].InterestDue - installments[i].AmountPaid
			if due <= 0 {
				continue
			}
			paid := allocation
			if paid > due {
				paid = due
			}
			installments[i].AmountPaid += paid
			allocation -= paid
			if installments[i].AmountPaid == installments[i].PrincipalDue+installments[i].InterestDue {
				installments[i].Status = "paid"
			} else {
				installments[i].Status = "partial"
			}
			if err := tx.Save(&installments[i]).Error; err != nil {
				return err
			}
			if allocation == 0 {
				break
			}
		}
		var count int64
		tx.Model(&models.LoanRepayment{}).Count(&count)
		if err := tx.Create(&models.LoanRepayment{LoanID: loan.ID, Amount: amount, PrincipalPaid: principalPaid, InterestPaid: interestPaid, ReferenceNumber: fmt.Sprintf("LR-%08d", count+1), RecordedBy: adminUsername}).Error; err != nil {
			return err
		}
		if loan.OutstandingPrincipal+loan.OutstandingInterest == 0 {
			loan.Status = "completed"
		}
		return tx.Save(&loan).Error
	})
}

func ProcessDueLoanCollections(now time.Time) error {
	var installments []models.LoanInstallment
	if err := db.DB.Where("due_date < ? AND status != ?", now, "paid").Order("due_date asc").Find(&installments).Error; err != nil {
		return err
	}
	for _, installment := range installments {
		_ = collectInstallment(installment, now)
	}
	return nil
}

func collectInstallment(installment models.LoanInstallment, now time.Time) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var loan models.Loan
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&loan, installment.LoanID).Error; err != nil || loan.Status != "disbursed" {
			return err
		}
		due := installment.PrincipalDue + installment.InterestDue - installment.AmountPaid
		if due <= 0 {
			return nil
		}
		var accounts []models.Account
		if err := tx.Where("user_id = ? AND active = ? AND account_type IN ?", loan.UserID, true, []string{"savings", "current"}).Order("CASE account_type WHEN 'savings' THEN 0 ELSE 1 END").Find(&accounts).Error; err != nil {
			return err
		}
		available := int64(0)
		for _, account := range accounts {
			available += account.Balance
		}
		if available < due {
			installment.Status = "overdue"
			if now.Sub(installment.DueDate) >= 90*24*time.Hour {
				installment.Status, loan.Status = "defaulted", "defaulted"
			}
			if err := tx.Save(&installment).Error; err != nil {
				return err
			}
			if err := tx.Save(&loan).Error; err != nil {
				return err
			}
			noticeType, title := "loan_overdue", "Loan installment overdue"
			if loan.Status == "defaulted" {
				noticeType, title = "loan_defaulted", "Loan classified as defaulted"
			}
			notification := models.MemberNotification{UserID: loan.UserID, Type: noticeType, Title: title, Message: fmt.Sprintf("KES %d remains due on loan installment %d.", due, installment.InstallmentNo), DedupKey: fmt.Sprintf("%s-%d-%s", noticeType, installment.ID, now.Format("2006-01-02"))}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&notification).Error; err != nil {
				return err
			}
			return tx.Create(&models.LoanCollectionAttempt{LoanID: loan.ID, InstallmentID: installment.ID, AmountRequested: due, Status: "failed", Details: "insufficient savings and current account balances"}).Error
		}
		remaining := due
		var sources []string
		for i := range accounts {
			debit := remaining
			if debit > accounts[i].Balance {
				debit = accounts[i].Balance
			}
			if debit == 0 {
				continue
			}
			accounts[i].Balance -= debit
			remaining -= debit
			if err := tx.Save(&accounts[i]).Error; err != nil {
				return err
			}
			sources = append(sources, accounts[i].Number)
			if remaining == 0 {
				break
			}
		}
		interestPaid := due
		if interestPaid > loan.OutstandingInterest {
			interestPaid = loan.OutstandingInterest
		}
		principalPaid := due - interestPaid
		loan.OutstandingInterest -= interestPaid
		loan.OutstandingPrincipal -= principalPaid
		installment.AmountPaid += due
		installment.Status = "paid"
		if loan.OutstandingInterest+loan.OutstandingPrincipal == 0 {
			loan.Status = "completed"
		}
		var count int64
		tx.Model(&models.LoanRepayment{}).Count(&count)
		if err := tx.Create(&models.LoanRepayment{LoanID: loan.ID, Amount: due, PrincipalPaid: principalPaid, InterestPaid: interestPaid, ReferenceNumber: fmt.Sprintf("LR-%08d", count+1), RecordedBy: "system_scheduler", SourceAccountNumber: strings.Join(sources, ","), Method: "automatic_debit"}).Error; err != nil {
			return err
		}
		if err := tx.Save(&installment).Error; err != nil {
			return err
		}
		if err := tx.Save(&loan).Error; err != nil {
			return err
		}
		return tx.Create(&models.LoanCollectionAttempt{LoanID: loan.ID, InstallmentID: installment.ID, AmountRequested: due, AmountCollected: due, Status: "completed", Details: "collected from " + strings.Join(sources, ",")}).Error
	})
}

func GetMemberNotifications(username string) ([]models.MemberNotification, error) {
	var notifications []models.MemberNotification
	err := db.DB.Joins("JOIN users ON users.id = member_notifications.user_id").Where("users.username = ?", strings.ToLower(strings.TrimSpace(username))).Order("member_notifications.created_at desc").Limit(20).Find(&notifications).Error
	return notifications, err
}

func StartLoanCollectionScheduler() {
	go func() {
		_ = ProcessDueLoanCollections(time.Now())
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for now := range ticker.C {
			_ = ProcessDueLoanCollections(now)
		}
	}()
}
