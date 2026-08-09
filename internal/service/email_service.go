package service

import (
	"context"
	"fmt"
	"log/slog"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmailService defines methods for sending emails.
type EmailService interface {
	// SendVerificationEmail sends a verification email to the user.
	SendVerificationEmail(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error

	// SendPasswordResetEmail sends a password reset email to the user.
	SendPasswordResetEmail(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error
}

// emailService implements EmailService.
type emailService struct {
	emailRepo          repository.EmailRepository
	sendgridClient     *sendgrid.Client
	senderEmail        string
	enabled            bool
	logger             *slog.Logger
	verificationTplID  string
	passwordResetTplID string
}

// NewEmailService creates and returns a new emailService.
func NewEmailService(
	emailRepo repository.EmailRepository,
	sendgridClient *sendgrid.Client,
	senderEmail string,
	enabled bool,
	verificationTplID string,
	passwordResetTplID string,
	logger *slog.Logger,
) EmailService {
	return &emailService{
		emailRepo:          emailRepo,
		sendgridClient:     sendgridClient,
		senderEmail:        senderEmail,
		enabled:            enabled,
		logger:             logger,
		verificationTplID:  verificationTplID,
		passwordResetTplID: passwordResetTplID,
	}
}

// SendVerificationEmail sends a verification email to the user.
func (s *emailService) SendVerificationEmail(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error {
	if !s.enabled {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "email sending is disabled", "user_id", userID.Hex(), "recipient", recipientEmail)
		}
		return nil
	}

	// Create email log
	email := &domain.Email{
		UserID:         userID,
		RecipientEmail: recipientEmail,
		TemplateID:     s.verificationTplID,
		DynamicData:    dynamicData,
		Status:         domain.EmailStatusPending,
	}

	savedEmail, err := s.emailRepo.SaveEmailLog(ctx, email)
	if err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to save email log", "user_id", userID.Hex(), "error", err)
		}
		return nil // Gracefully degrade - don't fail if logging fails
	}

	// Send email via SendGrid
	msgID, err := s.sendEmail(ctx, recipientEmail, s.verificationTplID, dynamicData)
	if err != nil {
		// Update status to failed in database
		errMsg := err.Error()
		updateErr := s.emailRepo.UpdateEmailStatus(ctx, savedEmail.ID.Hex(), domain.EmailStatusFailed, &errMsg)
		if updateErr != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to update email status", "email_id", savedEmail.ID.Hex(), "error", updateErr)
		}
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to send verification email", "user_id", userID.Hex(), "error", err)
		}
		return nil // Gracefully degrade - don't fail the parent operation
	}

	// Update status to sent in database
	updateErr := s.emailRepo.UpdateEmailStatus(ctx, savedEmail.ID.Hex(), domain.EmailStatusSent, nil)
	if updateErr != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "failed to update email status to sent", "email_id", savedEmail.ID.Hex(), "error", updateErr)
	}

	// Store sendgrid message ID if available
	if msgID != "" {
		updateErr := s.emailRepo.UpdateEmailStatus(ctx, savedEmail.ID.Hex(), domain.EmailStatusSent, nil)
		if updateErr != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to update sendgrid message id", "email_id", savedEmail.ID.Hex(), "error", updateErr)
		}
	}

	if s.logger != nil {
		s.logger.InfoContext(ctx, "verification email sent", "user_id", userID.Hex(), "recipient", recipientEmail)
	}
	return nil
}

// SendPasswordResetEmail sends a password reset email to the user.
func (s *emailService) SendPasswordResetEmail(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error {
	if !s.enabled {
		if s.logger != nil {
			s.logger.InfoContext(ctx, "email sending is disabled", "user_id", userID.Hex(), "recipient", recipientEmail)
		}
		return nil
	}

	// Create email log
	email := &domain.Email{
		UserID:         userID,
		RecipientEmail: recipientEmail,
		TemplateID:     s.passwordResetTplID,
		DynamicData:    dynamicData,
		Status:         domain.EmailStatusPending,
	}

	savedEmail, err := s.emailRepo.SaveEmailLog(ctx, email)
	if err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to save email log", "user_id", userID.Hex(), "error", err)
		}
		return nil // Gracefully degrade - don't fail if logging fails
	}

	// Send email via SendGrid
	msgID, err := s.sendEmail(ctx, recipientEmail, s.passwordResetTplID, dynamicData)
	if err != nil {
		// Update status to failed in database
		errMsg := err.Error()
		updateErr := s.emailRepo.UpdateEmailStatus(ctx, savedEmail.ID.Hex(), domain.EmailStatusFailed, &errMsg)
		if updateErr != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to update email status", "email_id", savedEmail.ID.Hex(), "error", updateErr)
		}
		if s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to send password reset email", "user_id", userID.Hex(), "error", err)
		}
		return nil // Gracefully degrade - don't fail the parent operation
	}

	// Update status to sent in database
	updateErr := s.emailRepo.UpdateEmailStatus(ctx, savedEmail.ID.Hex(), domain.EmailStatusSent, nil)
	if updateErr != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "failed to update email status to sent", "email_id", savedEmail.ID.Hex(), "error", updateErr)
	}

	// Store sendgrid message ID if available
	if msgID != "" {
		updateErr := s.emailRepo.UpdateEmailStatus(ctx, savedEmail.ID.Hex(), domain.EmailStatusSent, nil)
		if updateErr != nil && s.logger != nil {
			s.logger.ErrorContext(ctx, "failed to update sendgrid message id", "email_id", savedEmail.ID.Hex(), "error", updateErr)
		}
	}

	if s.logger != nil {
		s.logger.InfoContext(ctx, "password reset email sent", "user_id", userID.Hex(), "recipient", recipientEmail)
	}
	return nil
}

// sendEmail is a helper that sends an email via SendGrid using dynamic templates.
func (s *emailService) sendEmail(ctx context.Context, recipientEmail, templateID string, dynamicData map[string]string) (string, error) {
	from := mail.NewEmail("Temp Backend", s.senderEmail)
	to := mail.NewEmail("User", recipientEmail)

	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.Subject = "Email Notification"
	m.SetTemplateID(templateID)

	// Create personalization with recipient and dynamic data
	p := mail.NewPersonalization()
	p.AddTos(to)

	// Convert map[string]string to map[string]interface{} for dynamic template data
	dynamicDataInterface := make(map[string]interface{})
	for k, v := range dynamicData {
		dynamicDataInterface[k] = v
	}
	p.DynamicTemplateData = dynamicDataInterface
	m.AddPersonalizations(p)

	response, err := s.sendgridClient.SendWithContext(ctx, m)
	if err != nil {
		return "", fmt.Errorf("sendgrid send failed: %w", err)
	}

	// Check response status code
	if response.StatusCode >= 400 {
		return "", fmt.Errorf("sendgrid returned status %d: %s", response.StatusCode, response.Body)
	}

	// Extract message ID from response headers (X-Message-Id header)
	var messageID string
	if headers, ok := response.Headers["X-Message-Id"]; ok && len(headers) > 0 {
		messageID = headers[0]
	}

	return messageID, nil
}
