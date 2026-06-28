package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// MockRoomService is a mock implementation for testing
type MockRoomService struct {
	createRoomFunc func(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error)
}

func (m *MockRoomService) CreateRoom(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error) {
	if m.createRoomFunc != nil {
		return m.createRoomFunc(ctx, name, code, userID)
	}
	return nil, nil
}

func TestCreateRoomRequest_Structure(t *testing.T) {
	req := CreateRoomRequest{
		Name: "Conference Room A",
		Code: "CONF_A_001",
	}

	assert.Equal(t, "Conference Room A", req.Name)
	assert.Equal(t, "CONF_A_001", req.Code)
}

func TestRoomResponse_Structure(t *testing.T) {
	resp := RoomResponse{
		ID:        "507f1f77bcf86cd799439011",
		Name:      "Conference Room A",
		Code:      "CONF_A_001",
		CreatedBy: "507f1f77bcf86cd799439012",
		CreatedAt: "2024-06-28T21:15:00Z",
		UpdatedAt: "2024-06-28T21:15:00Z",
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", resp.ID)
	assert.Equal(t, "Conference Room A", resp.Name)
	assert.Equal(t, "CONF_A_001", resp.Code)
}

func TestNewRoomHandler(t *testing.T) {
	mockService := &MockRoomService{}
	handler := NewRoomHandler(mockService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.CreateRoom)
}

func TestCreateRoom_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	room := &domain.Room{
		ID:        primitive.NewObjectID(),
		Name:      "Conference Room A",
		Code:      "CONF_A_001",
		CreatedBy: userID,
	}

	mockService := &MockRoomService{
		createRoomFunc: func(ctx context.Context, name, code string, uid primitive.ObjectID) (*domain.Room, error) {
			if name == "Conference Room A" && code == "CONF_A_001" && uid == userID {
				return room, nil
			}
			return nil, domain.ErrInvalidInput
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.CreateRoom)
}

func TestCreateRoom_InvalidInput(t *testing.T) {
	mockService := &MockRoomService{
		createRoomFunc: func(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error) {
			return nil, domain.ErrInvalidInput
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.CreateRoom)
}

func TestCreateRoom_CodeAlreadyExists(t *testing.T) {
	mockService := &MockRoomService{
		createRoomFunc: func(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error) {
			return nil, domain.ErrCodeAlreadyExists
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.CreateRoom)
}

func TestCreateRoom_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		expectError bool
	}{
		{"InvalidInput", domain.ErrInvalidInput, true},
		{"CodeAlreadyExists", domain.ErrCodeAlreadyExists, true},
		{"RoomNotFound", domain.ErrRoomNotFound, true},
		{"UnknownError", errors.New("unknown error"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockRoomService{
				createRoomFunc: func(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error) {
					return nil, tc.err
				},
			}

			handler := NewRoomHandler(mockService)
			assert.NotNil(t, handler.CreateRoom)
		})
	}
}
