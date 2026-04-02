package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/refynehq/refyne-backend/internal/domains/email/service"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// EmailJobArgs represents the arguments for sending an email job
type EmailJobArgs struct {
	To           string                 `json:"to"`
	Subject      string                 `json:"subject"`
	TemplateName string                 `json:"template_name"`
	Data         map[string]interface{} `json:"data"`
}

// Kind returns the unique name for this job type
func (EmailJobArgs) Kind() string {
	return "send_email"
}

// EmailWorker processes email sending jobs
type EmailWorker struct {
	river.WorkerDefaults[EmailJobArgs]
	emailService service.EmailService
	logger       *zap.Logger
}

// NewEmailWorker creates a new email worker
func NewEmailWorker(emailService service.EmailService) *EmailWorker {
	return &EmailWorker{
		emailService: emailService,
		logger:       logging.GetComponentLogger("email-worker"),
	}
}

// Work processes the email sending job
func (w *EmailWorker) Work(ctx context.Context, job *river.Job[EmailJobArgs]) error {
	args := job.Args

	w.logger.Info("Processing email job",
		zap.Int64("job_id", job.ID),
		zap.String("to", args.To),
		zap.String("subject", args.Subject),
		zap.String("template", args.TemplateName),
		zap.Int("attempt", job.Attempt),
	)

	// Convert map to EmailData struct
	emailData := convertToEmailData(args.Data)

	// Send templated email
	if err := w.emailService.SendTemplatedEmail(
		args.To,
		args.Subject,
		args.TemplateName,
		emailData,
	); err != nil {
		w.logger.Error("Failed to send email",
			zap.Int64("job_id", job.ID),
			zap.String("to", args.To),
			zap.Int("attempt", job.Attempt),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	w.logger.Info("Email sent successfully",
		zap.Int64("job_id", job.ID),
		zap.String("to", args.To),
	)

	return nil
}

// convertToEmailData converts a map to EmailData struct
func convertToEmailData(data map[string]interface{}) service.EmailData {
	emailData := service.EmailData{}

	if to, ok := data["to"].(string); ok {
		emailData.To = to
	}
	if subject, ok := data["subject"].(string); ok {
		emailData.Subject = subject
	}
	if username, ok := data["username"].(string); ok {
		emailData.Username = username
	}
	if otp, ok := data["otp"].(string); ok {
		emailData.OTP = otp
	}
	if resetLink, ok := data["reset_link"].(string); ok {
		emailData.ResetLink = resetLink
	}
	if token, ok := data["token"].(string); ok {
		emailData.Token = token
	}

	return emailData
}

// QueueEmailJob queues an email job to be processed asynchronously
func QueueEmailJob(
	ctx context.Context,
	client *river.Client[any],
	to, subject, templateName string,
	data map[string]interface{},
) error {
	logger := logging.GetComponentLogger("email-queue")

	args := EmailJobArgs{
		To:           to,
		Subject:      subject,
		TemplateName: templateName,
		Data:         data,
	}

	// Marshal args for logging
	argsJSON, _ := json.Marshal(args)
	logger.Info("Queueing email job",
		zap.String("to", to),
		zap.String("template", templateName),
		zap.String("args", string(argsJSON)),
	)

	// Insert job with retry configuration
	job, err := client.Insert(ctx, args, &river.InsertOpts{
		MaxAttempts: 5,
		Priority:    2, // Medium priority
		Queue:       "email",
	})

	if err != nil {
		logger.Error("Failed to queue email job",
			zap.String("to", to),
			zap.Error(err),
		)
		return fmt.Errorf("failed to queue email job: %w", err)
	}

	logger.Info("Email job queued successfully",
		zap.Int64("job_id", job.Job.ID),
		zap.String("to", to),
	)

	return nil
}

// QueueOTPEmail queues an OTP email job
func QueueOTPEmail(ctx context.Context, client *river.Client[any], to, otp string) error {
	data := map[string]interface{}{
		"to":  to,
		"otp": otp,
	}

	return QueueEmailJob(ctx, client, to, "Your Refyne Verification Code", "otp.html", data)
}

// QueuePasswordResetEmail queues a password reset email job
func QueuePasswordResetEmail(ctx context.Context, client *river.Client[any], to, token, resetLink string) error {
	data := map[string]interface{}{
		"to":         to,
		"token":      token,
		"reset_link": resetLink,
	}

	return QueueEmailJob(ctx, client, to, "Reset Your Refyne Password", "password_reset.html", data)
}

// QueueWelcomeEmail queues a welcome email job
func QueueWelcomeEmail(ctx context.Context, client *river.Client[any], to, username string) error {
	data := map[string]interface{}{
		"to":       to,
		"username": username,
	}

	return QueueEmailJob(ctx, client, to, "Welcome to Refyne!", "welcome.html", data)
}

// QueueWorkspaceMemberInvitation queues a workspace member invitation email job
func QueueWorkspaceMemberInvitation(ctx context.Context, client *river.Client[any], to, workspaceName, invitedBy, invitationLink string) error {
	data := map[string]interface{}{
		"to":               to,
		"workspace_name":   workspaceName,
		"invited_by":       invitedBy,
		"invitation_link":  invitationLink,
	}

	return QueueEmailJob(ctx, client, to, "You're invited to "+workspaceName+" on Refyne", "workspace_invitation.html", data)
}
