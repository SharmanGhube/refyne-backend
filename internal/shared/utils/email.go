package utils

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func CheckValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)

	return regex.MatchString(email) && len(email) <= 254
}

func GenerateUUID() string {
	uuid := uuid.New()
	return strings.Replace(uuid.String(), "-", "", -1)
}
