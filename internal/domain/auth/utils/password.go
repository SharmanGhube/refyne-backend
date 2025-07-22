package auth

import (
	"errors"
	"fmt"
	"regexp"
)

type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}
}

// ValidatePassword checks if a password meets the policy requirements
func (p *PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters long", p.MinLength)
	}

	if p.RequireUpper && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	if p.RequireLower && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	if p.RequireDigit && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one digit")
	}

	if p.RequireSpecial && !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
