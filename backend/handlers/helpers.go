package handlers

import (
	"fmt"
	"net/http"
	"time"
)

// Shared by all handlers
func formatKES(amount int64) string {
	return fmt.Sprintf("KES %d", amount)
}

func formatDate(t time.Time) string {
	return t.Format("02 Jan 2006 15:04:05")
}

func getSessionUser(r *http.Request) string {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		return ""
	}
	return cookie.Value
}