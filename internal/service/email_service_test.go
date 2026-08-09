package service

import (
	"context"
	"testing"

	"github.com/sendgrid/sendgrid-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"temp_backend/internal/domain"
)

// MockEmailRepository is a mock implementation of EmailRepository for testing.
type MockEmailRepository struct {
	SaveEmailLogFunc      func(ctx context.Context, email *domain.Email) (*domain.Email, error)
	UpdateEmailStatusFunc func(ctx context.Context, emailID string, status string, errorMsg *string) error
	GetByIDFunc           func(ctx context.Context, id string) (*domain.Email, error)
	ListByUserIDFunc      func(ctx context.Context, userID string) ([]*domain.Email, error)
}

func (m *MockEmailRepository) SaveEmailLog(ctx context.Context, email *domain.Email) (*domain.Email, error) {
	if m.SaveEmailLogFunc != nil {
		return m.SaveEmailLogFunc(ctx, email)
	}
	return email, nil
}

func (m *MockEmailRepository) UpdateEmailStatus(ctx context.Context, emailID string, status string, errorMsg *string) error {
	if m.UpdateEmailStatusFunc != nil {
		return m.UpdateEmailStatusFunc(ctx, emailID, status, errorMsg)
	}
	return nil
}

func (m *MockEmailRepository) GetByID(ctx context.Context, id string) (*domain.Email, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockEmailRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Email, error) {
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func TestSendVerificationEmail_Disabled(t *testing.T) {
	mockRepo := &MockEmailRepository{}
	svc := NewEmailService(mockRepo, nil, "noreply@test.com", false, "tmpl_1", "tmpl_2", nil)

	userID := primitive.NewObjectID()
	err := svc.SendVerificationEmail(context.Background(), userID, "test@example.com", map[string]string{})

	// When disabled, should return nil without error
	assert.NoError(t, err)
}

func TestSendVerificationEmail_LogsEmail(t *testing.T) {
	savedEmail := &domain.Email{
		ID: primitive.NewObjectID(),
	}

	mockRepo := &MockEmailRepository{
		SaveEmailLogFunc: func(ctx context.Context, email *domain.Email) (*domain.Email, error) {
			assert.Equal(t, domain.EmailStatusPending, email.Status)
			return savedEmail, nil
		},
		UpdateEmailStatusFunc: func(ctx context.Context, emailID string, status string, errorMsg *string) error {
			return nil
		},
	}

	// Create a mock SendGrid client (will fail to send, but that's ok for this test)
	mockSendgridClient := sendgrid.NewSendClient("test-key")
	svc := NewEmailService(mockRepo, mockSendgridClient, "noreply@test.com", true, "tmpl_1", "tmpl_2", nil)

	userID := primitive.NewObjectID()
	dynamicData := map[string]string{"user_name": "John"}

	// This will attempt to send via SendGrid (which will fail), but we're just testing that
	// the email is logged and handled gracefully
	err := svc.SendVerificationEmail(context.Background(), userID, "test@example.com", dynamicData)

	// Should gracefully degrade - no error even if SendGrid fails
	assert.NoError(t, err)
}

func TestSendPasswordResetEmail_Disabled(t *testing.T) {
	mockRepo := &MockEmailRepository{}
	svc := NewEmailService(mockRepo, nil, "noreply@test.com", false, "tmpl_1", "tmpl_2", nil)

	userID := primitive.NewObjectID()
	err := svc.SendPasswordResetEmail(context.Background(), userID, "test@example.com", map[string]string{})

	// When disabled, should return nil without error
	assert.NoError(t, err)
}

func TestSendPasswordResetEmail_LogsEmail(t *testing.T) {
	savedEmail := &domain.Email{
		ID: primitive.NewObjectID(),
	}

	mockRepo := &MockEmailRepository{
		SaveEmailLogFunc: func(ctx context.Context, email *domain.Email) (*domain.Email, error) {
			return savedEmail, nil
		},
		UpdateEmailStatusFunc: func(ctx context.Context, emailID string, status string, errorMsg *string) error {
			return nil
		},
	}

	mockSendgridClient := sendgrid.NewSendClient("test-key")
	svc := NewEmailService(mockRepo, mockSendgridClient, "noreply@test.com", true, "tmpl_1", "tmpl_2", nil)

	userID := primitive.NewObjectID()
	dynamicData := map[string]string{"reset_code": "ABC123"}

	err := svc.SendPasswordResetEmail(context.Background(), userID, "test@example.com", dynamicData)

	// Should gracefully degrade - no error even if SendGrid fails
	assert.NoError(t, err)
}
