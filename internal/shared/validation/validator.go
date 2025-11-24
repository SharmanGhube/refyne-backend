package validation

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// Common validation patterns
var (
	EmailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	NameRegex     = regexp.MustCompile(`^[a-zA-Z\s\-']{1,50}$`)
	UUIDRegex     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// Dangerous patterns for XSS detection
	ScriptTagRegex    = regexp.MustCompile(`(?i)<script[\s\S]*?>[\s\S]*?</script>`)
	OnEventRegex      = regexp.MustCompile(`(?i)on\w+\s*=`)
	JavascriptRegex   = regexp.MustCompile(`(?i)javascript:`)
	DataURIRegex      = regexp.MustCompile(`(?i)data:text/html`)
	IframeRegex       = regexp.MustCompile(`(?i)<iframe[\s\S]*?>`)
	ObjectRegex       = regexp.MustCompile(`(?i)<object[\s\S]*?>`)
	EmbedRegex        = regexp.MustCompile(`(?i)<embed[\s\S]*?>`)
	SQLInjectionRegex = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|<script|onerror|onload)`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validator provides validation utilities
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) *ValidationError {
	email = strings.TrimSpace(email)

	if email == "" {
		return &ValidationError{Field: "email", Message: "Email is required"}
	}

	if len(email) > 255 {
		return &ValidationError{Field: "email", Message: "Email must not exceed 255 characters"}
	}

	if !EmailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "Invalid email format"}
	}

	return nil
}

// ValidateUsername validates username format
func (v *Validator) ValidateUsername(username string) *ValidationError {
	username = strings.TrimSpace(username)

	if username == "" {
		return &ValidationError{Field: "username", Message: "Username is required"}
	}

	if len(username) < 3 {
		return &ValidationError{Field: "username", Message: "Username must be at least 3 characters"}
	}

	if len(username) > 30 {
		return &ValidationError{Field: "username", Message: "Username must not exceed 30 characters"}
	}

	if !UsernameRegex.MatchString(username) {
		return &ValidationError{Field: "username", Message: "Username can only contain letters, numbers, and underscores"}
	}

	return nil
}

// ValidateName validates first/last name
func (v *Validator) ValidateName(name, fieldName string) *ValidationError {
	name = strings.TrimSpace(name)

	if name == "" {
		return &ValidationError{Field: fieldName, Message: fieldName + " is required"}
	}

	if len(name) < 1 {
		return &ValidationError{Field: fieldName, Message: fieldName + " must be at least 1 character"}
	}

	if len(name) > 50 {
		return &ValidationError{Field: fieldName, Message: fieldName + " must not exceed 50 characters"}
	}

	if !NameRegex.MatchString(name) {
		return &ValidationError{Field: fieldName, Message: fieldName + " contains invalid characters"}
	}

	return nil
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(password string) *ValidationError {
	if password == "" {
		return &ValidationError{Field: "password", Message: "Password is required"}
	}

	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "Password must be at least 8 characters"}
	}

	if len(password) > 128 {
		return &ValidationError{Field: "password", Message: "Password must not exceed 128 characters"}
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &ValidationError{Field: "password", Message: "Password must contain at least one uppercase letter"}
	}
	if !hasLower {
		return &ValidationError{Field: "password", Message: "Password must contain at least one lowercase letter"}
	}
	if !hasNumber {
		return &ValidationError{Field: "password", Message: "Password must contain at least one number"}
	}
	if !hasSpecial {
		return &ValidationError{Field: "password", Message: "Password must contain at least one special character"}
	}

	return nil
}

// ValidateUUID validates UUID format
func (v *Validator) ValidateUUID(id, fieldName string) *ValidationError {
	id = strings.TrimSpace(strings.ToLower(id))

	if id == "" {
		return &ValidationError{Field: fieldName, Message: fieldName + " is required"}
	}

	if !UUIDRegex.MatchString(id) {
		return &ValidationError{Field: fieldName, Message: "Invalid " + fieldName + " format"}
	}

	return nil
}

// SanitizeString removes potentially dangerous content from strings
func (v *Validator) SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// HTML escape
	input = html.EscapeString(input)

	return input
}

// SanitizeHTML sanitizes HTML content (more aggressive)
func (v *Validator) SanitizeHTML(input string) string {
	// Remove script tags
	input = ScriptTagRegex.ReplaceAllString(input, "")

	// Remove event handlers
	input = OnEventRegex.ReplaceAllString(input, "")

	// Remove javascript: protocol
	input = JavascriptRegex.ReplaceAllString(input, "")

	// Remove data URIs
	input = DataURIRegex.ReplaceAllString(input, "")

	// Remove dangerous tags
	input = IframeRegex.ReplaceAllString(input, "")
	input = ObjectRegex.ReplaceAllString(input, "")
	input = EmbedRegex.ReplaceAllString(input, "")

	// HTML escape remaining content
	input = html.EscapeString(input)

	return input
}

// DetectXSS checks for XSS attack patterns
func (v *Validator) DetectXSS(input string) bool {
	input = strings.ToLower(input)

	// Check for script tags
	if ScriptTagRegex.MatchString(input) {
		return true
	}

	// Check for event handlers
	if OnEventRegex.MatchString(input) {
		return true
	}

	// Check for javascript: protocol
	if JavascriptRegex.MatchString(input) {
		return true
	}

	// Check for data URIs
	if DataURIRegex.MatchString(input) {
		return true
	}

	// Check for dangerous tags
	if IframeRegex.MatchString(input) || ObjectRegex.MatchString(input) || EmbedRegex.MatchString(input) {
		return true
	}

	return false
}

// DetectSQLInjection checks for SQL injection patterns
func (v *Validator) DetectSQLInjection(input string) bool {
	return SQLInjectionRegex.MatchString(input)
}

// ValidateAndSanitizeInput performs comprehensive validation and sanitization
func (v *Validator) ValidateAndSanitizeInput(input string, maxLength int) (string, *ValidationError) {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check for XSS
	if v.DetectXSS(input) {
		return "", &ValidationError{Field: "input", Message: "Input contains potentially dangerous content"}
	}

	// Check for SQL injection
	if v.DetectSQLInjection(input) {
		return "", &ValidationError{Field: "input", Message: "Input contains potentially dangerous patterns"}
	}

	// Check length
	if len(input) > maxLength {
		return "", &ValidationError{Field: "input", Message: "Input exceeds maximum length"}
	}

	// Sanitize
	sanitized := v.SanitizeString(input)

	return sanitized, nil
}

// NewValidationAppError creates an AppError from validation errors
func NewValidationAppError(c *gin.Context, validationErrors []*ValidationError) *errors.AppError {
	details := make(map[string]interface{})
	for _, err := range validationErrors {
		details[err.Field] = err.Message
	}

	return &errors.AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		Type:       errors.ErrorTypeValidation,
		HTTPStatus: 400,
		Details:    details,
	}
}
