package service

import (
	"context"
	"time"

	"temp_backend/internal/domain"
	"temp_backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoomService defines the interface for room business logic
type RoomService interface {
	CreateRoom(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error)
	AddUserToRoom(ctx context.Context, code string, userID primitive.ObjectID) (*domain.Room, error)
	GetRoom(ctx context.Context, code string) (*domain.Room, error)
	GetRoomByID(ctx context.Context, roomID primitive.ObjectID) (*domain.Room, error)
	LeaveRoom(ctx context.Context, code string, userID primitive.ObjectID) error
	DeleteRoom(ctx context.Context, code string, userID primitive.ObjectID) error
	ListUserRooms(ctx context.Context, userID primitive.ObjectID) ([]*domain.Room, error)
}

type roomService struct {
	repo repository.RoomRepository
}

// NewRoomService creates a new room service
func NewRoomService(repo repository.RoomRepository) RoomService {
	return &roomService{repo: repo}
}

// CreateRoom creates a new room with validation
func (s *roomService) CreateRoom(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error) {
	// Create room object
	now := time.Now()
	room := &domain.Room{
		Name:      name,
		Code:      code,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate room
	if err := room.Validate(); err != nil {
		return nil, err
	}

	// Check if code already exists
	existingRoom, err := s.repo.GetByCode(ctx, code)
	if err == nil && existingRoom != nil {
		return nil, domain.ErrCodeAlreadyExists
	}
	if err != nil && err != domain.ErrRoomNotFound {
		return nil, err
	}

	// Create room in repository
	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// AddUserToRoom adds a user to a room by room code
func (s *roomService) AddUserToRoom(ctx context.Context, code string, userID primitive.ObjectID) (*domain.Room, error) {
	// Get room by code
	room, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Add user to room members
	if err := s.repo.AddUserToRoom(ctx, room.ID, userID); err != nil {
		return nil, err
	}

	// Fetch updated room
	updatedRoom, err := s.repo.GetByID(ctx, room.ID)
	if err != nil {
		return nil, err
	}

	return updatedRoom, nil
}

// GetRoom retrieves a room by code
func (s *roomService) GetRoom(ctx context.Context, code string) (*domain.Room, error) {
	room, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoomByID retrieves a room by ID
func (s *roomService) GetRoomByID(ctx context.Context, roomID primitive.ObjectID) (*domain.Room, error) {
	room, err := s.repo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return room, nil
}

// LeaveRoom removes a user from a room
func (s *roomService) LeaveRoom(ctx context.Context, code string, userID primitive.ObjectID) error {
	// Get room by code
	room, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	// Check if user is a member (not the owner)
	isMember := false
	for _, m := range room.Members {
		if m == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return domain.ErrInvalidInput
	}

	// Remove user from room members
	if err := s.repo.RemoveUserFromRoom(ctx, room.ID, userID); err != nil {
		return err
	}

	return nil
}

// DeleteRoom soft-deletes a room (only owner can delete)
func (s *roomService) DeleteRoom(ctx context.Context, code string, userID primitive.ObjectID) error {
	// Get room by code
	room, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	// Check if user is the owner
	if room.CreatedBy != userID {
		return domain.ErrInvalidInput
	}

	// Soft delete the room
	if err := s.repo.DeleteRoom(ctx, room.ID); err != nil {
		return err
	}

	return nil
}

// ListUserRooms retrieves all rooms where the user is the creator or a member
func (s *roomService) ListUserRooms(ctx context.Context, userID primitive.ObjectID) ([]*domain.Room, error) {
	rooms, err := s.repo.ListUserRooms(ctx, userID)
	if err != nil {
		return nil, err
	}

	return rooms, nil
}
