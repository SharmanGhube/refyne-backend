package service

import (
	"fmt"

	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

// resendService implements SMTPService interface using Resend API
type resendService struct {
	client *resend.Client
	from   string
	logger *zap.Logger
}

// NewResendService creates a new Resend service instance
func NewResendService(cfg *config.Config) SMTPService {
	logger := logging.GetComponentLogger("resend-service")

	client := resend.NewClient(cfg.ResendAPIKey)

	return &resendService{
		client: client,
		from:   "noreply@refyne.com",
		logger: logger,
	}
}

// Send sends an email via Resend API
func (r *resendService) Send(to, subject, body string) error {
	r.logger.Info("Sending email via Resend",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		r.logger.Error("Failed to send email via Resend",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	r.logger.Info("Email sent successfully via Resend",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}

// SendBatch sends multiple emails in a batch via Resend
func (r *resendService) SendBatch(recipients []string, subject, body string) error {
	r.logger.Info("Sending batch emails via Resend",
		zap.Int("count", len(recipients)),
		zap.String("subject", subject),
	)

	var errors []string
	successCount := 0

	for _, recipient := range recipients {
		if err := r.Send(recipient, subject, body); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", recipient, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		r.logger.Warn("Some batch emails failed to send",
			zap.Int("success", successCount),
			zap.Int("failed", len(errors)),
		)
		return fmt.Errorf("failed to send %d emails", len(errors))
	}

	r.logger.Info("All batch emails sent successfully via Resend",
		zap.Int("count", successCount),
	)
	return nil
}
