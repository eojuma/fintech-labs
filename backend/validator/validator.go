package validator

import (
	"regexp"
	"strings"
)

func ValidUsername(username string) bool {
	username = strings.ToLower(strings.TrimSpace(username))

	if len(username) < 3 || len(username) > 30 {
		return false
	}

	for _, v := range username {
		if !(v >= 'a' && v <= 'z'|| v >= '0' && v <= '9' || v == '.' || v == '-' || v == '_') {
			return false
		}
	}
	return true
}

func ValidFullName(fullname string)bool{
	fullname=strings.TrimSpace(fullname)
if len(fullname)<4 || len(fullname)>100{
	return false
}

for _,name:=range fullname{
	if !(name >= 'a' && name <= 'z' || name >= 'A' && name <= 'Z' ||  name==' ') {
			return false
		}
}
return true
}


func ValidEmail(email string)bool{
email=strings.TrimSpace(email)
if len(email) <5 || len(email)>254{
	return false
}
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
return emailRegex.MatchString(strings.ToLower(email))
}
