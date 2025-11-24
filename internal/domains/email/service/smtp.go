package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// SMTPService defines the interface for SMTP operations
type SMTPService interface {
	// Send sends an email via SMTP
	Send(to, subject, body string) error
	
	// SendBatch sends multiple emails in a batch
	SendBatch(recipients []string, subject, body string) error
}

// smtpService implements SMTPService interface
type smtpService struct {
	config *config.SMTPConfig
	auth   smtp.Auth
	logger *zap.Logger
}

// NewSMTPService creates a new SMTP service instance
func NewSMTPService(cfg *config.Config) SMTPService {
	logger := logging.GetComponentLogger("smtp-service")

	// Create SMTP auth
	auth := smtp.PlainAuth(
		"",
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.Host,
	)

	return &smtpService{
		config: &cfg.SMTP,
		auth:   auth,
		logger: logger,
	}
}

// Send sends an email via SMTP
func (s *smtpService) Send(to, subject, body string) error {
	s.logger.Debug("Preparing to send email",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	// Build email message
	message := s.buildMessage(s.config.Username, to, subject, body)

	// Get SMTP address
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Send email based on TLS/SSL configuration
	var err error
	if s.config.UseSSL {
		err = s.sendWithSSL(addr, to, message)
	} else if s.config.UseTLS {
		err = s.sendWithTLS(addr, to, message)
	} else {
		err = s.sendPlain(addr, to, message)
	}

	if err != nil {
		s.logger.Error("Failed to send email via SMTP",
			zap.String("to", to),
			zap.String("host", s.config.Host),
			zap.Int("port", s.config.Port),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully via SMTP",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}

// SendBatch sends multiple emails in a batch
func (s *smtpService) SendBatch(recipients []string, subject, body string) error {
	s.logger.Info("Sending batch emails",
		zap.Int("count", len(recipients)),
		zap.String("subject", subject),
	)

	var errors []string
	successCount := 0

	for _, recipient := range recipients {
		if err := s.Send(recipient, subject, body); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", recipient, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		s.logger.Warn("Some emails failed to send",
			zap.Int("success", successCount),
			zap.Int("failed", len(errors)),
			zap.Strings("errors", errors),
		)
		return fmt.Errorf("failed to send %d emails: %s", len(errors), strings.Join(errors, "; "))
	}

	s.logger.Info("All batch emails sent successfully", zap.Int("count", successCount))
	return nil
}

// buildMessage constructs the email message with headers
func (s *smtpService) buildMessage(from, to, subject, body string) []byte {
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	return []byte(message)
}

// sendPlain sends email without encryption
func (s *smtpService) sendPlain(addr, to string, message []byte) error {
	return smtp.SendMail(addr, s.auth, s.config.Username, []string{to}, message)
}

// sendWithTLS sends email with STARTTLS
func (s *smtpService) sendWithTLS(addr, to string, message []byte) error {
	return smtp.SendMail(addr, s.auth, s.config.Username, []string{to}, message)
}

// sendWithSSL sends email with SSL/TLS
func (s *smtpService) sendWithSSL(addr, to string, message []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with SSL: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.Auth(s.auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(s.config.Username); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}
