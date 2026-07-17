package utils

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"fintech-labs/internal/db"
	"fintech-labs/internal/models"
)

func ValidUsername(username string) bool {
	username = strings.ToLower(strings.TrimSpace(username))

	if len(username) < 3 || len(username) > 30 {
		return false
	}

	for _, v := range username {
		if !(v >= 'a' && v <= 'z' || v >= '0' && v <= '9' || v == '.' || v == '-' || v == '_') {
			return false
		}
	}
	return true
}

func ValidFullName(fullname string) bool {
	fullname = strings.TrimSpace(fullname)
	if len(fullname) < 4 || len(fullname) > 100 {
		return false
	}

	for _, name := range fullname {
		if !(name >= 'a' && name <= 'z' || name >= 'A' && name <= 'Z' || name == ' ') {
			return false
		}
	}
	return true
}

func ValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}

func ValidPhoneNumber(phone string) bool {
	phone = strings.TrimSpace(phone)

	phoneRegex := regexp.MustCompile(`^(?:254|\+254|0)?(7|1|2)\d{8}$`)

	return phoneRegex.MatchString(phone)
}

func ValidNationalID(Id string) bool {
	Id = strings.TrimSpace(Id)
	if len(Id) < 7 || len(Id) > 8 {
		return false
	}
	for _, ch := range Id {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func FormatKES(amount int64) string {
	if amount < 0 {
		return fmt.Sprintf("-KES %d", -amount)
	}
	return fmt.Sprintf("KES %d", amount)
}

func FormatDate(t time.Time) string {
	loc, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return t.Format("02 Jan 2006 15:04:05")
	}
	return t.In(loc).Format("02 Jan 2006 15:04:05")
}

func GetSessionUser(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		return ""
	}

	token := cookie.Value
	if token == "" {
		return ""
	}

	var session models.Session

	result := db.DB.Preload("User").Where("token = ?", token).First(&session)
	if result.Error != nil {
		return ""
	}

	if time.Since(session.LastActivityAt) > 10*time.Minute {
		db.DB.Delete(&session)
		return ""
	}

	db.DB.Model(&session).Update("last_activity_at", time.Now())
	http.SetCookie(w, &http.Cookie{
		Name:     "session_user",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   600,
		Secure:   os.Getenv("RENDER") == "true",
		SameSite: http.SameSiteStrictMode,
	})
	return session.User.Username
}


func FormatPhoneForSMS(phone string)string{
	phone = strings.TrimSpace(phone)

	if strings.HasPrefix(phone,"+"){
		return phone
	}

	if strings.HasPrefix(phone,"0"){
		return "+254"+phone[1:]
	}
	if strings.HasPrefix(phone,"254"){
		return "+"+phone
	}
	return phone
}