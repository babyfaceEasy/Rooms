package service

import (
	"context"
	"errors"
	"testing"

	"temp_backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	createFunc     func(ctx context.Context, user *domain.User) error
	getByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	getByIDFunc    func(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	getByIDsFunc   func(ctx context.Context, ids []primitive.ObjectID) ([]*domain.User, error)
	getByCodeFunc  func(ctx context.Context, code string) (*domain.User, error)
	updateFunc     func(ctx context.Context, user *domain.User) error
	deleteFunc     func(ctx context.Context, id primitive.ObjectID) error
	softDeleteFunc func(ctx context.Context, id primitive.ObjectID) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockUserRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.User, error) {
	if m.getByIDsFunc != nil {
		return m.getByIDsFunc(ctx, ids)
	}
	return []*domain.User{}, nil
}

func (m *MockUserRepository) GetByCode(ctx context.Context, code string) (*domain.User, error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	if m.softDeleteFunc != nil {
		return m.softDeleteFunc(ctx, id)
	}
	return nil
}

func TestRegister_Success(t *testing.T) {
	userRepoMock := &MockUserRepository{
		createFunc: func(ctx context.Context, user *domain.User) error {
			user.ID = primitive.NewObjectID()
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "SecurePass123!", true)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, true, user.IsAgeVerified)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "SecurePass123!", user.PasswordHash)
}

func TestRegister_PasswordHashedWithBcrypt(t *testing.T) {
	mock := &MockUserRepository{
		createFunc: func(ctx context.Context, user *domain.User) error {
			user.ID = primitive.NewObjectID()
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)
	password := "SecurePass123!"

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", password, true)

	assert.NoError(t, err)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.NoError(t, err)
}

func TestRegister_AgeVerificationRequired(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "SecurePass123!", false)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrAgeVerificationRequired))
}

func TestRegister_InvalidEmail_Empty(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "", "SecurePass123!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidEmail))
}

func TestRegister_InvalidEmail_Format(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	testCases := []string{
		"notanemail",
		"missing@domain",
		"@nodomain.com",
		"spaces in@email.com",
	}

	for _, email := range testCases {
		user, err := svc.Register(context.Background(), "John Doe", email, "SecurePass123!", true)
		assert.Error(t, err, "should reject email: %s", email)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, domain.ErrInvalidEmail))
	}
}

func TestRegister_InvalidPassword_TooShort(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "Short1!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestRegister_InvalidPassword_NoUppercase(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "lowercase1!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestRegister_InvalidPassword_NoLowercase(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "UPPERCASE1!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestRegister_InvalidPassword_NoDigit(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "NoDigits!Abc", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestRegister_InvalidPassword_NoSpecialChar(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "NoSpecial1Abc", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestRegister_InvalidName_Empty(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "", "john@example.com", "SecurePass123!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestRegister_InvalidName_TooShort(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "J", "john@example.com", "SecurePass123!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestRegister_InvalidName_TooLong(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	user, err := svc.Register(context.Background(), longName, "john@example.com", "SecurePass123!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mock := &MockUserRepository{
		createFunc: func(ctx context.Context, user *domain.User) error {
			return domain.ErrEmailAlreadyExists
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "john@example.com", "SecurePass123!", true)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrEmailAlreadyExists))
}

func TestRegister_EmailNormalization(t *testing.T) {
	var capturedEmail string
	mock := &MockUserRepository{
		createFunc: func(ctx context.Context, user *domain.User) error {
			capturedEmail = user.Email
			user.ID = primitive.NewObjectID()
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.Register(context.Background(), "John Doe", "John@EXAMPLE.COM", "SecurePass123!", true)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "john@example.com", capturedEmail)
}

func TestGetUserByID_Success(t *testing.T) {
	existingUser := &domain.User{
		ID:    primitive.NewObjectID(),
		Name:  "John Doe",
		Email: "john@example.com",
	}

	mock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return existingUser, nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.GetUserByID(context.Background(), existingUser.ID.Hex())

	assert.NoError(t, err)
	assert.Equal(t, existingUser, user)
}

func TestGetUserByID_InvalidID(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.GetUserByID(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestDeleteUser_Success(t *testing.T) {
	deletedID := primitive.NewObjectID()
	var capturedID primitive.ObjectID

	mock := &MockUserRepository{
		deleteFunc: func(ctx context.Context, id primitive.ObjectID) error {
			capturedID = id
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	err := svc.DeleteUser(context.Background(), deletedID.Hex())

	assert.NoError(t, err)
	assert.Equal(t, deletedID, capturedID)
}

func TestDeleteUser_InvalidID(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	err := svc.DeleteUser(context.Background(), "invalid")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestUpdateProfile_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	existingUser := &domain.User{
		ID:    userID,
		Name:  "Old Name",
		Email: "john@example.com",
	}

	var capturedUser *domain.User
	mock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			if id == userID {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
		updateFunc: func(ctx context.Context, user *domain.User) error {
			capturedUser = user
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.UpdateProfile(context.Background(), userID.Hex(), "New Name")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "New Name", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, userID, capturedUser.ID)
	assert.Equal(t, "New Name", capturedUser.Name)
}

func TestUpdateProfile_InvalidID(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.UpdateProfile(context.Background(), "invalid", "New Name")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	mock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.UpdateProfile(context.Background(), primitive.NewObjectID().Hex(), "New Name")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestUpdateProfile_InvalidName_Empty(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.UpdateProfile(context.Background(), primitive.NewObjectID().Hex(), "")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestUpdateProfile_InvalidName_TooShort(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	user, err := svc.UpdateProfile(context.Background(), primitive.NewObjectID().Hex(), "J")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestUpdateProfile_InvalidName_TooLong(t *testing.T) {
	mock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(mock, refreshTokenRepoMock)

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}

	user, err := svc.UpdateProfile(context.Background(), primitive.NewObjectID().Hex(), longName)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestChangePassword_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	currentPassword := "OldPass123!"
	newPassword := "NewPass456!"

	// Create a password hash for the existing password
	existingHash, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	assert.NoError(t, err)

	existingUser := &domain.User{
		ID:           userID,
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: string(existingHash),
	}

	var updatedUser *domain.User
	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			if id == userID {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
		updateFunc: func(ctx context.Context, user *domain.User) error {
			updatedUser = user
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{
		deleteByUserIDFunc: func(ctx context.Context, userID primitive.ObjectID) error {
			return nil
		},
	}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err = svc.ChangePassword(context.Background(), userID.Hex(), currentPassword, newPassword)

	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	// Verify new password hash is different from old one
	assert.NotEqual(t, string(existingHash), updatedUser.PasswordHash)
	// Verify new hash matches new password
	err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(newPassword))
	assert.NoError(t, err)
}

func TestChangePassword_InvalidCurrentPassword(t *testing.T) {
	userID := primitive.NewObjectID()
	currentPassword := "CorrectPass123!"

	// Create a password hash
	existingHash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:           userID,
		PasswordHash: string(existingHash),
	}

	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return existingUser, nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.ChangePassword(context.Background(), userID.Hex(), "WrongPass123!", "NewPass456!")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestChangePassword_InvalidNewPassword(t *testing.T) {
	userID := primitive.NewObjectID()
	currentPassword := "OldPass123!"

	existingHash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:           userID,
		PasswordHash: string(existingHash),
	}

	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return existingUser, nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	// Test with invalid new password (too short)
	err := svc.ChangePassword(context.Background(), userID.Hex(), currentPassword, "Short1!")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestChangePassword_SameAsCurrentPassword(t *testing.T) {
	userID := primitive.NewObjectID()
	password := "SamePass123!"

	existingHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:           userID,
		PasswordHash: string(existingHash),
	}

	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return existingUser, nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	// Test with new password same as current
	err := svc.ChangePassword(context.Background(), userID.Hex(), password, password)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPassword))
}

func TestChangePassword_UserNotFound(t *testing.T) {
	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.ChangePassword(context.Background(), primitive.NewObjectID().Hex(), "OldPass123!", "NewPass456!")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestChangePassword_InvalidID(t *testing.T) {
	userRepoMock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.ChangePassword(context.Background(), "invalid", "OldPass123!", "NewPass456!")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestChangePassword_InvalidatesRefreshTokens(t *testing.T) {
	userID := primitive.NewObjectID()
	currentPassword := "OldPass123!"
	newPassword := "NewPass456!"

	existingHash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:           userID,
		PasswordHash: string(existingHash),
	}

	var deletedByUserID primitive.ObjectID
	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return existingUser, nil
		},
		updateFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{
		deleteByUserIDFunc: func(ctx context.Context, uid primitive.ObjectID) error {
			deletedByUserID = uid
			return nil
		},
	}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.ChangePassword(context.Background(), userID.Hex(), currentPassword, newPassword)

	assert.NoError(t, err)
	// Verify refresh tokens were invalidated for this user
	assert.Equal(t, userID, deletedByUserID)
}

func TestDeleteAccount_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	existingUser := &domain.User{
		ID:    userID,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	var softDeletedID primitive.ObjectID
	var deletedByUserID primitive.ObjectID

	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			if id == userID {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
		softDeleteFunc: func(ctx context.Context, id primitive.ObjectID) error {
			softDeletedID = id
			return nil
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{
		deleteByUserIDFunc: func(ctx context.Context, uid primitive.ObjectID) error {
			deletedByUserID = uid
			return nil
		},
	}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.DeleteAccount(context.Background(), userID.Hex())

	assert.NoError(t, err)
	// Verify soft delete was called
	assert.Equal(t, userID, softDeletedID)
	// Verify refresh tokens were invalidated
	assert.Equal(t, userID, deletedByUserID)
}

func TestDeleteAccount_UserNotFound(t *testing.T) {
	userRepoMock := &MockUserRepository{
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.DeleteAccount(context.Background(), primitive.NewObjectID().Hex())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestDeleteAccount_InvalidID(t *testing.T) {
	userRepoMock := &MockUserRepository{}
	refreshTokenRepoMock := &MockRefreshTokenRepository{}
	svc := NewUserService(userRepoMock, refreshTokenRepoMock)

	err := svc.DeleteAccount(context.Background(), "invalid")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}
