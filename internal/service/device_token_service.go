package service

import (
	"context"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeviceTokenService defines operations for managing device tokens.
type DeviceTokenService interface {
	// RegisterDevice registers a new device token for a user.
	RegisterDevice(ctx context.Context, userID primitive.ObjectID, token string, platform domain.DeviceTokenPlatform) (*domain.DeviceToken, error)

	// UnregisterDevice removes a device token.
	UnregisterDevice(ctx context.Context, tokenID primitive.ObjectID) error

	// GetDeviceToken retrieves a device token by ID.
	GetDeviceToken(ctx context.Context, tokenID primitive.ObjectID) (*domain.DeviceToken, error)

	// GetActiveDeviceTokens retrieves all active device tokens for a user.
	GetActiveDeviceTokens(ctx context.Context, userID primitive.ObjectID) ([]*domain.DeviceToken, error)

	// DeleteStaleTokens removes device tokens older than the specified number of days.
	// This is typically called by a cleanup job.
	DeleteStaleTokens(ctx context.Context, expiryDays int) error
}

// deviceTokenService is the concrete implementation of DeviceTokenService.
type deviceTokenService struct {
	deviceTokenRepo repository.DeviceTokenRepository
}

// NewDeviceTokenService creates a new DeviceTokenService.
func NewDeviceTokenService(deviceTokenRepo repository.DeviceTokenRepository) DeviceTokenService {
	return &deviceTokenService{
		deviceTokenRepo: deviceTokenRepo,
	}
}

func (s *deviceTokenService) RegisterDevice(
	ctx context.Context,
	userID primitive.ObjectID,
	token string,
	platform domain.DeviceTokenPlatform,
) (*domain.DeviceToken, error) {
	deviceToken := &domain.DeviceToken{
		ID:       primitive.NewObjectID(),
		UserID:   userID,
		Token:    token,
		Platform: platform,
		IsActive: true,
	}

	if err := s.deviceTokenRepo.Create(ctx, deviceToken); err != nil {
		return nil, err
	}

	return deviceToken, nil
}

func (s *deviceTokenService) UnregisterDevice(ctx context.Context, tokenID primitive.ObjectID) error {
	return s.deviceTokenRepo.Delete(ctx, tokenID)
}

func (s *deviceTokenService) GetDeviceToken(ctx context.Context, tokenID primitive.ObjectID) (*domain.DeviceToken, error) {
	return s.deviceTokenRepo.GetByID(ctx, tokenID)
}

func (s *deviceTokenService) GetActiveDeviceTokens(ctx context.Context, userID primitive.ObjectID) ([]*domain.DeviceToken, error) {
	return s.deviceTokenRepo.GetActiveByUserID(ctx, userID)
}

func (s *deviceTokenService) DeleteStaleTokens(ctx context.Context, expiryDays int) error {
	return s.deviceTokenRepo.DeleteStaleTokens(ctx, expiryDays)
}
