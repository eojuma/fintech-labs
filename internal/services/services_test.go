package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "fintech-labs-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	conn, err := gorm.Open(sqlite.Open(filepath.Join(dir, "test.db")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// Allow concurrent background goroutines (email/SMS/suspicious-activity) to
	// share the test DB without immediately hitting SQLITE_BUSY.
	conn.Exec("PRAGMA busy_timeout = 5000")

	if err := conn.AutoMigrate(&models.User{}, &models.Account{}, &models.Transaction{}, &models.Session{}); err != nil {
		panic(err)
	}

	db.DB = conn

	os.Exit(m.Run())
}

var testSeq int

// newTestUser creates a user with a unique username/email/phone/national-ID,
// opens a current account for them, and sets a transaction PIN of "1234".
func newTestUser(t *testing.T) (*models.User, *models.Account) {
	t.Helper()
	testSeq++
	digits := fmt.Sprintf("%08d", testSeq)
	username := fmt.Sprintf("user%d", testSeq)

	user, err := CreateUser("Test Member", username, username+"@example.com", "2547"+digits, digits, "secret123", "customer")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	account, err := CreateAccountForUser(user.ID)
	if err != nil {
		t.Fatalf("CreateAccountForUser: %v", err)
	}
	if err := SetTransactionPin(user.Username, "1234"); err != nil {
		t.Fatalf("SetTransactionPin: %v", err)
	}
	return user, account
}

func setBalance(t *testing.T, accountNumber string, balance int64) {
	t.Helper()
	if err := db.DB.Model(&models.Account{}).Where("number = ?", accountNumber).Update("balance", balance).Error; err != nil {
		t.Fatalf("setBalance: %v", err)
	}
}

func accountBalance(t *testing.T, accountNumber string) int64 {
	t.Helper()
	var account models.Account
	if err := db.DB.Where("number = ?", accountNumber).First(&account).Error; err != nil {
		t.Fatalf("accountBalance: %v", err)
	}
	return account.Balance
}

func TestDepositOwnershipRejected(t *testing.T) {
	attacker, _ := newTestUser(t)
	_, victim := newTestUser(t)

	if _, err := Deposit(attacker.Username, victim.Number, 500); err == nil {
		t.Fatal("expected deposit into another user's account to fail")
	}
	if got := accountBalance(t, victim.Number); got != 0 {
		t.Fatalf("victim balance changed: got %d, want 0", got)
	}
}

func TestWithdrawOwnershipRejected(t *testing.T) {
	attacker, _ := newTestUser(t)
	_, victim := newTestUser(t)
	setBalance(t, victim.Number, 100000)

	if _, err := Withdraw(attacker.Username, victim.Number, 200); err == nil {
		t.Fatal("expected withdrawal from another user's account to fail")
	}
	if got := accountBalance(t, victim.Number); got != 100000 {
		t.Fatalf("victim balance changed: got %d, want 100000", got)
	}
}

func TestTransferOwnershipRejected(t *testing.T) {
	attacker, attackerAccount := newTestUser(t)
	_, victim := newTestUser(t)
	setBalance(t, victim.Number, 100000)

	if _, err := SendMoney(attacker.Username, victim.Number, attackerAccount.Number, 100); err == nil {
		t.Fatal("expected transfer from another user's account to fail")
	}
	if got := accountBalance(t, victim.Number); got != 100000 {
		t.Fatalf("victim balance changed: got %d, want 100000", got)
	}
}

func TestDepositHappyPath(t *testing.T) {
	user, account := newTestUser(t)

	ref, err := Deposit(user.Username, account.Number, 500)
	if err != nil {
		t.Fatalf("Deposit: %v", err)
	}
	if ref == "" {
		t.Fatal("expected a reference number")
	}
	if got := accountBalance(t, account.Number); got != 500 {
		t.Fatalf("balance after deposit: got %d, want 500", got)
	}
}

func TestWithdrawHappyPathAndInsufficientFunds(t *testing.T) {
	user, account := newTestUser(t)
	setBalance(t, account.Number, 1000)

	if _, err := Withdraw(user.Username, account.Number, 400); err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
	if got := accountBalance(t, account.Number); got != 600 {
		t.Fatalf("balance after withdrawal: got %d, want 600", got)
	}

	if _, err := Withdraw(user.Username, account.Number, 99999); err == nil {
		t.Fatal("expected insufficient funds error")
	}
	if got := accountBalance(t, account.Number); got != 600 {
		t.Fatalf("balance changed after failed withdrawal: got %d, want 600", got)
	}
}

func TestTransferHappyPath(t *testing.T) {
	sender, senderAccount := newTestUser(t)
	_, recipientAccount := newTestUser(t)
	setBalance(t, senderAccount.Number, 1000)

	if _, err := SendMoney(sender.Username, senderAccount.Number, recipientAccount.Number, 300); err != nil {
		t.Fatalf("SendMoney: %v", err)
	}
	if got := accountBalance(t, senderAccount.Number); got != 700 {
		t.Fatalf("sender balance: got %d, want 700", got)
	}
	if got := accountBalance(t, recipientAccount.Number); got != 300 {
		t.Fatalf("recipient balance: got %d, want 300", got)
	}
}

func TestCreateUserRejectsShortPassword(t *testing.T) {
	testSeq++
	username := fmt.Sprintf("user%d", testSeq)
	_, err := CreateUser("Test Member", username, username+"@example.com", "2547"+fmt.Sprintf("%08d", testSeq), fmt.Sprintf("%08d", testSeq), "abc", "customer")
	if err == nil {
		t.Fatal("expected short password to be rejected")
	}
}

func TestVerifyTransactionPinLockout(t *testing.T) {
	user, _ := newTestUser(t)

	for i := 0; i < 4; i++ {
		if err := VerifyTransactionPin(user.Username, "0000"); err == nil {
			t.Fatal("expected incorrect PIN to fail")
		}
	}

	// Fifth failure locks the PIN.
	if err := VerifyTransactionPin(user.Username, "0000"); err == nil {
		t.Fatal("expected fifth failure to lock the PIN")
	}

	// Even the correct PIN must be rejected while locked.
	if err := VerifyTransactionPin(user.Username, "1234"); err == nil {
		t.Fatal("expected correct PIN to be rejected while locked")
	}
}
