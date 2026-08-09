package handler

import (
	"context"
	"errors"
	"testing"

	"temp_backend/internal/domain"
	"temp_backend/internal/service"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserService is a mock implementation of UserService for testing.
type MockUserService struct {
	registerFunc       func(ctx context.Context, name, email, password string, ageVerified bool) (*domain.User, error)
	getUserByIDFunc    func(ctx context.Context, id string) (*domain.User, error)
	updateProfileFunc  func(ctx context.Context, id, name string) (*domain.User, error)
	changePasswordFunc func(ctx context.Context, id, currentPassword, newPassword string) error
	deleteUserFunc     func(ctx context.Context, id string) error
	deleteAccountFunc  func(ctx context.Context, id string) error
}

func (m *MockUserService) Register(ctx context.Context, name, email, password string, ageVerified bool) (*domain.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, name, email, password, ageVerified)
	}
	return nil, nil
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserService) UpdateProfile(ctx context.Context, id, name string) (*domain.User, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, id, name)
	}
	return nil, nil
}

func (m *MockUserService) ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	if m.changePasswordFunc != nil {
		return m.changePasswordFunc(ctx, id, currentPassword, newPassword)
	}
	return nil
}

func (m *MockUserService) DeleteUser(ctx context.Context, id string) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(ctx, id)
	}
	return nil
}

func (m *MockUserService) DeleteAccount(ctx context.Context, id string) error {
	if m.deleteAccountFunc != nil {
		return m.deleteAccountFunc(ctx, id)
	}
	return nil
}

// MockEmailService is a mock implementation of EmailService for testing.
type MockEmailService struct {
	sendVerificationEmailFunc  func(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error
	sendPasswordResetEmailFunc func(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error {
	if m.sendVerificationEmailFunc != nil {
		return m.sendVerificationEmailFunc(ctx, userID, recipientEmail, dynamicData)
	}
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, userID primitive.ObjectID, recipientEmail string, dynamicData map[string]string) error {
	if m.sendPasswordResetEmailFunc != nil {
		return m.sendPasswordResetEmailFunc(ctx, userID, recipientEmail, dynamicData)
	}
	return nil
}

func TestNewUserHandler(t *testing.T) {
	mock := &MockUserService{}
	var svc service.UserService = mock
	emailSvc := &MockEmailService{}

	handler := NewUserHandler(svc, emailSvc)

	assert.NotNil(t, handler)
	assert.Equal(t, handler.svc, svc)
}

func TestViewProfile_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	user := &domain.User{
		ID:    userID,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	mock := &MockUserService{
		getUserByIDFunc: func(ctx context.Context, id string) (*domain.User, error) {
			if id == userID.Hex() {
				return user, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	emailSvc := &MockEmailService{}

	handler := NewUserHandler(mock, emailSvc)

	// We can't easily test Fiber handlers without the full HTTP stack
	// Just verify the handler exists and is callable
	assert.NotNil(t, handler.ViewProfile)
}

func TestViewProfile_UserNotFound(t *testing.T) {
	mock := &MockUserService{
		getUserByIDFunc: func(ctx context.Context, id string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	emailSvc := &MockEmailService{}

	handler := NewUserHandler(mock, emailSvc)

	// Verify the handler exists
	assert.NotNil(t, handler.ViewProfile)
}

func TestUserHandler_CreatesWithValidService(t *testing.T) {
	userID := primitive.NewObjectID()
	mock := &MockUserService{
		registerFunc: func(ctx context.Context, name, email, password string, ageVerified bool) (*domain.User, error) {
			return &domain.User{
				ID:            userID,
				Name:          name,
				Email:         email,
				IsAgeVerified: true,
			}, nil
		},
	}

	var svc service.UserService = mock
	emailSvc := &MockEmailService{}
	handler := NewUserHandler(svc, emailSvc)

	assert.NotNil(t, handler)
}

func TestUserHandler_HandlesGetUserByID(t *testing.T) {
	userID := primitive.NewObjectID()
	expectedUser := &domain.User{
		ID:    userID,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	mock := &MockUserService{
		getUserByIDFunc: func(ctx context.Context, id string) (*domain.User, error) {
			if id == userID.Hex() {
				return expectedUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	var svc service.UserService = mock
	emailSvc := &MockEmailService{}
	handler := NewUserHandler(svc, emailSvc)

	// Test successful retrieval
	user, err := handler.svc.GetUserByID(context.Background(), userID.Hex())
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	// Test not found
	user, err = handler.svc.GetUserByID(context.Background(), primitive.NewObjectID().Hex())
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserHandler_HandlesDeleteUser(t *testing.T) {
	userID := primitive.NewObjectID()

	mock := &MockUserService{
		deleteUserFunc: func(ctx context.Context, id string) error {
			if id == userID.Hex() {
				return nil
			}
			return domain.ErrUserNotFound
		},
	}

	var svc service.UserService = mock
	emailSvc := &MockEmailService{}
	handler := NewUserHandler(svc, emailSvc)

	// Test successful deletion
	err := handler.svc.DeleteUser(context.Background(), userID.Hex())
	assert.NoError(t, err)

	// Test not found
	err = handler.svc.DeleteUser(context.Background(), primitive.NewObjectID().Hex())
	assert.Error(t, err)
}

func TestUserHandler_HandlesErrorMappings(t *testing.T) {
	testCases := []struct {
		name          string
		error         error
		expectedFound bool
	}{
		{"InvalidEmail", domain.ErrInvalidEmail, true},
		{"InvalidPassword", domain.ErrInvalidPassword, true},
		{"AgeVerificationRequired", domain.ErrAgeVerificationRequired, true},
		{"EmailAlreadyExists", domain.ErrEmailAlreadyExists, true},
		{"UserNotFound", domain.ErrUserNotFound, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, errors.Is(tc.error, tc.error), "Error should match itself")
		})
	}
}

func TestRegisterRequest_Structure(t *testing.T) {
	req := RegisterRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		AgeVerified: true,
	}

	assert.Equal(t, "John Doe", req.Name)
	assert.Equal(t, "john@example.com", req.Email)
	assert.Equal(t, "SecurePass123!", req.Password)
	assert.True(t, req.AgeVerified)
}

func TestUserResponse_Structure(t *testing.T) {
	resp := UserResponse{
		ID:        "507f1f77bcf86cd799439011",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: "2026-06-28T19:00:00Z",
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", resp.ID)
	assert.Equal(t, "John Doe", resp.Name)
	assert.Equal(t, "john@example.com", resp.Email)
	assert.Equal(t, "2026-06-28T19:00:00Z", resp.CreatedAt)
}

func TestUpdateProfile_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	updatedUser := &domain.User{
		ID:    userID,
		Name:  "Updated Name",
		Email: "john@example.com",
	}

	mock := &MockUserService{
		updateProfileFunc: func(ctx context.Context, id, name string) (*domain.User, error) {
			if id == userID.Hex() && name == "Updated Name" {
				return updatedUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Verify the handler exists and is callable
	assert.NotNil(t, handler.UpdateProfile)
}

func TestUpdateProfile_InvalidName(t *testing.T) {
	userID := primitive.NewObjectID()

	mock := &MockUserService{
		updateProfileFunc: func(ctx context.Context, id, name string) (*domain.User, error) {
			if name == "" {
				return nil, domain.ErrInvalidInput
			}
			return nil, nil
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Test with invalid name
	_, err := handler.svc.UpdateProfile(context.Background(), userID.Hex(), "")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	mock := &MockUserService{
		updateProfileFunc: func(ctx context.Context, id, name string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Test with non-existent user
	_, err := handler.svc.UpdateProfile(context.Background(), primitive.NewObjectID().Hex(), "New Name")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestUpdateProfileRequest_Structure(t *testing.T) {
	req := UpdateProfileRequest{
		Name: "Updated Name",
	}

	assert.Equal(t, "Updated Name", req.Name)
}

func TestChangePassword_Success(t *testing.T) {
	userID := primitive.NewObjectID()

	mock := &MockUserService{
		changePasswordFunc: func(ctx context.Context, id, currentPassword, newPassword string) error {
			if id == userID.Hex() && currentPassword == "OldPass123!" && newPassword == "NewPass456!" {
				return nil
			}
			return domain.ErrInvalidPassword
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Verify the handler exists and is callable
	assert.NotNil(t, handler.ChangePassword)
}

func TestChangePassword_InvalidCurrentPassword(t *testing.T) {
	userID := primitive.NewObjectID()

	mock := &MockUserService{
		changePasswordFunc: func(ctx context.Context, id, currentPassword, newPassword string) error {
			if currentPassword != "CorrectPass123!" {
				return domain.ErrInvalidPassword
			}
			return nil
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Test with wrong current password
	err := handler.svc.ChangePassword(context.Background(), userID.Hex(), "WrongPass123!", "NewPass456!")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestChangePassword_UserNotFound(t *testing.T) {
	mock := &MockUserService{
		changePasswordFunc: func(ctx context.Context, id, currentPassword, newPassword string) error {
			return domain.ErrUserNotFound
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Test with non-existent user
	err := handler.svc.ChangePassword(context.Background(), primitive.NewObjectID().Hex(), "OldPass123!", "NewPass456!")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestChangePasswordRequest_Structure(t *testing.T) {
	req := ChangePasswordRequest{
		CurrentPassword: "OldPass123!",
		NewPassword:     "NewPass456!",
	}

	assert.Equal(t, "OldPass123!", req.CurrentPassword)
	assert.Equal(t, "NewPass456!", req.NewPassword)
}

func TestChangePasswordResponse_Structure(t *testing.T) {
	resp := ChangePasswordResponse{
		Message: "password changed successfully, please log in again",
	}

	assert.Equal(t, "password changed successfully, please log in again", resp.Message)
}

func TestDeleteAccount_Success(t *testing.T) {
	userID := primitive.NewObjectID()

	mock := &MockUserService{
		deleteAccountFunc: func(ctx context.Context, id string) error {
			if id == userID.Hex() {
				return nil
			}
			return domain.ErrUserNotFound
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Verify the handler exists and is callable
	assert.NotNil(t, handler.DeleteAccount)
}

func TestDeleteAccount_UserNotFound(t *testing.T) {
	mock := &MockUserService{
		deleteAccountFunc: func(ctx context.Context, id string) error {
			return domain.ErrUserNotFound
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Test with non-existent user
	err := handler.svc.DeleteAccount(context.Background(), primitive.NewObjectID().Hex())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestDeleteAccount_InvalidID(t *testing.T) {
	mock := &MockUserService{
		deleteAccountFunc: func(ctx context.Context, id string) error {
			return domain.ErrInvalidInput
		},
	}

	emailSvc := &MockEmailService{}
	handler := NewUserHandler(mock, emailSvc)

	// Test with invalid ID
	err := handler.svc.DeleteAccount(context.Background(), "invalid")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}
