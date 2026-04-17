package handlers

import (
	"fmt"
	"net/http"
	"time"
)

func formatKES(amount int64) string {
	if amount < 0 {
		return fmt.Sprintf("-KES %d", -amount)
	}
	return fmt.Sprintf("KES %d", amount)
}

func formatDate(t time.Time) string {
	// Adding 3 hours for East Africa Time if needed
	return t.Format("02 Jan 2006 15:04:05")
}

func getSessionUser(r *http.Request) string {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		return ""
	}
	return cookie.Value
}