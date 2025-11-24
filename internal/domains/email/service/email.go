package service

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// EmailService defines the interface for email operations
type EmailService interface {
	// SendEmail sends a plain text email
	SendEmail(to, subject, body string) error
	
	// SendTemplatedEmail sends an email using a template
	SendTemplatedEmail(to, subject, templateName string, data interface{}) error
	
	// SendOTP sends an OTP code via email
	SendOTP(to, otp string) error
	
	// SendPasswordReset sends a password reset link via email
	SendPasswordReset(to, token, resetLink string) error
	
	// SendWelcome sends a welcome email to new users
	SendWelcome(to, username string) error
	
	// SendVerification sends an account verification email
	SendVerification(to, username, verificationLink string) error
}

// emailService implements EmailService interface
type emailService struct {
	smtpService SMTPService
	templates   *template.Template
	logger      *zap.Logger
}

// EmailData represents common data for email templates
type EmailData struct {
	To        string
	Subject   string
	Username  string
	OTP       string
	ResetLink string
	Token     string
}

// NewEmailService creates a new email service instance
func NewEmailService(smtpService SMTPService) (EmailService, error) {
	logger := logging.GetComponentLogger("email-service")

	// Load all email templates
	templates, err := template.ParseGlob("internal/domains/email/templates/*.html")
	if err != nil {
		logger.Error("Failed to load email templates", zap.Error(err))
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return &emailService{
		smtpService: smtpService,
		templates:   templates,
		logger:      logger,
	}, nil
}

// SendEmail sends a plain text email
func (s *emailService) SendEmail(to, subject, body string) error {
	s.logger.Info("Sending plain text email",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	if err := s.smtpService.Send(to, subject, body); err != nil {
		s.logger.Error("Failed to send email",
			zap.String("to", to),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully", zap.String("to", to))
	return nil
}

// SendTemplatedEmail sends an email using a template
func (s *emailService) SendTemplatedEmail(to, subject, templateName string, data interface{}) error {
	s.logger.Info("Sending templated email",
		zap.String("to", to),
		zap.String("template", templateName),
		zap.String("subject", subject),
	)

	// Render template
	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		s.logger.Error("Failed to render email template",
			zap.String("template", templateName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	// Send email
	if err := s.smtpService.Send(to, subject, body.String()); err != nil {
		s.logger.Error("Failed to send templated email",
			zap.String("to", to),
			zap.String("template", templateName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send templated email: %w", err)
	}

	s.logger.Info("Templated email sent successfully",
		zap.String("to", to),
		zap.String("template", templateName),
	)
	return nil
}

// SendOTP sends an OTP code via email
func (s *emailService) SendOTP(to, otp string) error {
	s.logger.Info("Sending OTP email", zap.String("to", to))

	data := EmailData{
		To:  to,
		OTP: otp,
	}

	subject := "Your Refyne Verification Code"
	return s.SendTemplatedEmail(to, subject, "otp.html", data)
}

// SendPasswordReset sends a password reset link via email
func (s *emailService) SendPasswordReset(to, token, resetLink string) error {
	s.logger.Info("Sending password reset email", zap.String("to", to))

	data := EmailData{
		To:        to,
		Token:     token,
		ResetLink: resetLink,
	}

	subject := "Reset Your Refyne Password"
	return s.SendTemplatedEmail(to, subject, "password_reset.html", data)
}

// SendWelcome sends a welcome email to new users
func (s *emailService) SendWelcome(to, username string) error {
	s.logger.Info("Sending welcome email",
		zap.String("to", to),
		zap.String("username", username),
	)

	data := EmailData{
		To:       to,
		Username: username,
	}

	subject := "Welcome to Refyne!"
	return s.SendTemplatedEmail(to, subject, "welcome.html", data)
}

// SendVerification sends an account verification email
func (s *emailService) SendVerification(to, username, verificationLink string) error {
	s.logger.Info("Sending verification email",
		zap.String("to", to),
		zap.String("username", username),
	)

	data := map[string]string{
		"To":               to,
		"Username":         username,
		"VerificationLink": verificationLink,
	}

	subject := "Verify Your Refyne Account"
	return s.SendTemplatedEmail(to, subject, "verification.html", data)
}
