package validator

import "strings"

func ValidUsername(username string) bool {
	username = strings.TrimSpace(username)

	if len(username) < 4 || len(username) > 64 {
		return false
	}

	for _, v := range username {
		if !(v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9' || v == '.' || v == '-' || v == '_' || v==' ') {
			return false
		}
	}
	return true
}
