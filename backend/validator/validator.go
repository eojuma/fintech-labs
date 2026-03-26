package validator

func ValidUsername(Username string) bool {
	if len(Username) == 0 {
		return false
	}
	for _, v := range Username {
		if !(v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z') {
			return false
		}
	}
	return true
}
