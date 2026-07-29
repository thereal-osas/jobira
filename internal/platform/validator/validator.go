package validator

import "strings"

func Required (value string) bool {
	return strings.TrimSpace(value) != ""
}

func MinLength(value string, min int) bool {
	return len(strings.TrimSpace(value)) >= min
}

func IsEmail(value string) bool {
	value = strings.TrimSpace(value)

	return strings.Contains(value, "@") && strings.Contains(value, ".")
}