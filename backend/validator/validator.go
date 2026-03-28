package validator

import "strings"

func ValidUsername(username string) bool {
	username = strings.TrimSpace(username)

	if len(username) < 3 || len(username) > 20 {
		return false
	}

	for _, v := range username {
	
		if !(v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z') {
			return false
		}
	}

	reserved := []string{"admin", "root", "support", "system", "fintech", "official"}
	for _, word := range reserved {
		if strings.ToLower(username) == word {
			return false
		}
	}

	return true
}
