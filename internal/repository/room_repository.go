package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// RoomRepository defines the interface for room persistence
type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Room, error)
	GetByCode(ctx context.Context, code string) (*domain.Room, error)
	AddUserToRoom(ctx context.Context, roomID, userID primitive.ObjectID) error
	RemoveUserFromRoom(ctx context.Context, roomID, userID primitive.ObjectID) error
}
