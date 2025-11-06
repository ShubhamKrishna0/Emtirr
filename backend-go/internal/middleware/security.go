package middleware

import (
	"regexp"
	"strings"
)

type ValidationResult struct {
	Valid    bool
	Error    string
	Username string
	Column   int
}

func ValidateUsername(username string) ValidationResult {
	if username == "" {
		return ValidationResult{Valid: false, Error: "Username is required"}
	}

	trimmed := strings.TrimSpace(username)
	if len(trimmed) < 2 || len(trimmed) > 20 {
		return ValidationResult{Valid: false, Error: "Username must be 2-20 characters"}
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", trimmed)
	if !matched {
		return ValidationResult{Valid: false, Error: "Username can only contain letters, numbers, _ and -"}
	}

	return ValidationResult{Valid: true, Username: trimmed}
}

func ValidateMove(column int) ValidationResult {
	if column < 0 || column > 6 {
		return ValidationResult{Valid: false, Error: "Invalid column"}
	}
	return ValidationResult{Valid: true, Column: column}
}