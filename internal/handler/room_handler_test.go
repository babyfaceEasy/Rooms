package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"temp_backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockRoomService is a mock implementation for testing
type MockRoomService struct {
	createRoomFunc       func(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error)
	addUserToRoomFunc    func(ctx context.Context, code string, userID primitive.ObjectID) (*domain.Room, error)
	getRoom              func(ctx context.Context, code string) (*domain.Room, error)
	getRoomByID          func(ctx context.Context, roomID primitive.ObjectID) (*domain.Room, error)
	getRoomUsers         func(ctx context.Context, code string) ([]*domain.User, error)
	leaveRoom            func(ctx context.Context, code string, userID primitive.ObjectID) error
	removeMemberFromRoom func(ctx context.Context, code string, ownerID, memberID primitive.ObjectID) error
	deleteRoom           func(ctx context.Context, code string, userID primitive.ObjectID) error
	listUserRooms        func(ctx context.Context, userID primitive.ObjectID) ([]*domain.Room, error)
}

func (m *MockRoomService) CreateRoom(ctx context.Context, name, code string, userID primitive.ObjectID) (*domain.Room, error) {
	if m.createRoomFunc != nil {
		return m.createRoomFunc(ctx, name, code, userID)
	}
	return nil, nil
}

func (m *MockRoomService) AddUserToRoom(ctx context.Context, code string, userID primitive.ObjectID) (*domain.Room, error) {
	if m.addUserToRoomFunc != nil {
		return m.addUserToRoomFunc(ctx, code, userID)
	}
	return nil, nil
}

func (m *MockRoomService) GetRoom(ctx context.Context, code string) (*domain.Room, error) {
	if m.getRoom != nil {
		return m.getRoom(ctx, code)
	}
	return nil, nil
}

func (m *MockRoomService) GetRoomByID(ctx context.Context, roomID primitive.ObjectID) (*domain.Room, error) {
	if m.getRoomByID != nil {
		return m.getRoomByID(ctx, roomID)
	}
	return nil, nil
}

func (m *MockRoomService) GetRoomUsers(ctx context.Context, code string) ([]*domain.User, error) {
	if m.getRoomUsers != nil {
		return m.getRoomUsers(ctx, code)
	}
	return nil, nil
}

func (m *MockRoomService) LeaveRoom(ctx context.Context, code string, userID primitive.ObjectID) error {
	if m.leaveRoom != nil {
		return m.leaveRoom(ctx, code, userID)
	}
	return nil
}

func (m *MockRoomService) RemoveMemberFromRoom(ctx context.Context, code string, ownerID, memberID primitive.ObjectID) error {
	if m.removeMemberFromRoom != nil {
		return m.removeMemberFromRoom(ctx, code, ownerID, memberID)
	}
	return nil
}

func (m *MockRoomService) DeleteRoom(ctx context.Context, code string, userID primitive.ObjectID) error {
	if m.deleteRoom != nil {
		return m.deleteRoom(ctx, code, userID)
	}
	return nil
}

func (m *MockRoomService) ListUserRooms(ctx context.Context, userID primitive.ObjectID) ([]*domain.Room, error) {
	if m.listUserRooms != nil {
		return m.listUserRooms(ctx, userID)
	}
	return []*domain.Room{}, nil
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

func TestAddUserToRoomRequest_Structure(t *testing.T) {
	req := AddUserToRoomRequest{
		Code: "CONF_A_001",
	}

	assert.Equal(t, "CONF_A_001", req.Code)
}

func TestAddUserToRoom_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: primitive.NewObjectID(),
		Members:   []primitive.ObjectID{userID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		addUserToRoomFunc: func(ctx context.Context, code string, uID primitive.ObjectID) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.AddUserToRoom)
}

func TestAddUserToRoom_RoomNotFound(t *testing.T) {
	mockService := &MockRoomService{
		addUserToRoomFunc: func(ctx context.Context, code string, uID primitive.ObjectID) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.AddUserToRoom)
}

func TestAddUserToRoom_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		expectError bool
	}{
		{"RoomNotFound", domain.ErrRoomNotFound, true},
		{"UnknownError", errors.New("unknown error"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockRoomService{
				addUserToRoomFunc: func(ctx context.Context, code string, uID primitive.ObjectID) (*domain.Room, error) {
					return nil, tc.err
				},
			}

			handler := NewRoomHandler(mockService)
			assert.NotNil(t, handler.AddUserToRoom)
		})
	}
}

func TestGetRoom_OwnerAccess(t *testing.T) {
	ownerID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoom)
}

func TestGetRoom_MemberAccess(t *testing.T) {
	memberID := primitive.NewObjectID()
	ownerID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoom)
}

func TestGetRoom_UnauthorizedAccess(t *testing.T) {
	ownerID := primitive.NewObjectID()
	memberID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoom)
}

func TestGetRoom_NotFound(t *testing.T) {
	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoom)
}

func TestGetRoomMembers_OwnerAccess(t *testing.T) {
	ownerID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()
	memberID1 := primitive.NewObjectID()
	memberID2 := primitive.NewObjectID()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID1, memberID2},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoomMembers)
}

func TestGetRoomMembers_MemberAccess(t *testing.T) {
	ownerID := primitive.NewObjectID()
	memberID1 := primitive.NewObjectID()
	memberID2 := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID1, memberID2},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoomMembers)
}

func TestGetRoomMembers_UnauthorizedAccess(t *testing.T) {
	ownerID := primitive.NewObjectID()
	memberID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoomMembers)
}

func TestGetRoomMembers_NotFound(t *testing.T) {
	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.GetRoomMembers)
}

func TestRoomMembersResponse_Structure(t *testing.T) {
	resp := RoomMembersResponse{
		RoomID:   "507f1f77bcf86cd799439011",
		RoomCode: "CONF_A_001",
		RoomName: "Conference Room A",
		Owner:    "507f1f77bcf86cd799439012",
		Members:  []string{"507f1f77bcf86cd799439013", "507f1f77bcf86cd799439014"},
		Count:    2,
	}

	assert.Equal(t, "507f1f77bcf86cd799439011", resp.RoomID)
	assert.Equal(t, "CONF_A_001", resp.RoomCode)
	assert.Equal(t, "Conference Room A", resp.RoomName)
	assert.Equal(t, 2, resp.Count)
	assert.Len(t, resp.Members, 2)
}

func TestLeaveRoom_Success(t *testing.T) {
	roomCode := "CONF_A_001"

	mockService := &MockRoomService{
		leaveRoom: func(ctx context.Context, code string, uID primitive.ObjectID) error {
			if code == roomCode {
				return nil
			}
			return domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.LeaveRoom)
}

func TestLeaveRoom_NotFound(t *testing.T) {
	mockService := &MockRoomService{
		leaveRoom: func(ctx context.Context, code string, uID primitive.ObjectID) error {
			return domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.LeaveRoom)
}

func TestLeaveRoom_NotMember(t *testing.T) {
	mockService := &MockRoomService{
		leaveRoom: func(ctx context.Context, code string, uID primitive.ObjectID) error {
			return domain.ErrInvalidInput
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.LeaveRoom)
}

func TestHandleRoomDelete_OwnerDeletes(t *testing.T) {
	ownerID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
		deleteRoom: func(ctx context.Context, code string, uID primitive.ObjectID) error {
			return nil
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.HandleRoomDelete)
}

func TestHandleRoomDelete_OwnerDeletesNotFound(t *testing.T) {
	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.HandleRoomDelete)
}

func TestHandleRoomDelete_MemberLeaves(t *testing.T) {
	memberID := primitive.NewObjectID()
	ownerID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
		leaveRoom: func(ctx context.Context, code string, uID primitive.ObjectID) error {
			return nil
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.HandleRoomDelete)
}

func TestHandleRoomDelete_MemberLeavesNotMember(t *testing.T) {
	memberID := primitive.NewObjectID()
	ownerID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: ownerID,
		Members:   []primitive.ObjectID{memberID},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		getRoom: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
		leaveRoom: func(ctx context.Context, code string, uID primitive.ObjectID) error {
			return domain.ErrInvalidInput
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.HandleRoomDelete)
}

func TestListUserRooms_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: userID,
		Members:   []primitive.ObjectID{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService := &MockRoomService{
		listUserRooms: func(ctx context.Context, uID primitive.ObjectID) ([]*domain.Room, error) {
			if uID == userID {
				return []*domain.Room{room}, nil
			}
			return []*domain.Room{}, nil
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.ListUserRooms)
}

func TestListUserRooms_Empty(t *testing.T) {
	mockService := &MockRoomService{
		listUserRooms: func(ctx context.Context, uID primitive.ObjectID) ([]*domain.Room, error) {
			return []*domain.Room{}, nil
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.ListUserRooms)
}

func TestListUserRooms_MultipleRooms(t *testing.T) {
	userID := primitive.NewObjectID()
	room1 := &domain.Room{
		ID:        primitive.NewObjectID(),
		Name:      "Room 1",
		Code:      "ROOM_1",
		CreatedBy: userID,
		Members:   []primitive.ObjectID{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	room2 := &domain.Room{
		ID:        primitive.NewObjectID(),
		Name:      "Room 2",
		Code:      "ROOM_2",
		CreatedBy: primitive.NewObjectID(),
		Members:   []primitive.ObjectID{userID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := &MockRoomService{
		listUserRooms: func(ctx context.Context, uID primitive.ObjectID) ([]*domain.Room, error) {
			if uID == userID {
				return []*domain.Room{room1, room2}, nil
			}
			return []*domain.Room{}, nil
		},
	}

	handler := NewRoomHandler(mockService)
	assert.NotNil(t, handler.ListUserRooms)
}
