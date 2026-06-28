package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
	"temp_backend/internal/repository"
)

// RoomService defines the interface for room business logic
type RoomService interface {
	CreateRoom(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error)
	AddUserToRoom(ctx context.Context, code string, userID primitive.ObjectID) (*domain.Room, error)
	GetRoom(ctx context.Context, code string) (*domain.Room, error)
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
