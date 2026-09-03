package validator

import (
	"net/mail"
	"strings"
)

func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MinLength(value string, min int) bool {
	return len(strings.TrimSpace(value)) >= min
}

func IsEmail(value string) bool {
	value = strings.TrimSpace(value)

	if value == "" {
		return false
	}

	address, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}

	if address.Address != value {
		return false
	}

	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]

	return strings.Contains(domain, ".") &&
		!strings.HasPrefix(domain, ".") &&
		!strings.HasSuffix(domain, ".")
}
