package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"temp_backend/internal/domain"
)

// MockRoomRepository is a mock implementation for testing
type MockRoomRepository struct {
	createFunc        func(ctx context.Context, room *domain.Room) error
	getByIDFunc       func(ctx context.Context, id primitive.ObjectID) (*domain.Room, error)
	getByCodeFunc     func(ctx context.Context, code string) (*domain.Room, error)
	addUserToRoomFunc func(ctx context.Context, roomID, userID primitive.ObjectID) error
}

func (m *MockRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, room)
	}
	return nil
}

func (m *MockRoomRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Room, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrRoomNotFound
}

func (m *MockRoomRepository) GetByCode(ctx context.Context, code string) (*domain.Room, error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	return nil, domain.ErrRoomNotFound
}

func (m *MockRoomRepository) AddUserToRoom(ctx context.Context, roomID, userID primitive.ObjectID) error {
	if m.addUserToRoomFunc != nil {
		return m.addUserToRoomFunc(ctx, roomID, userID)
	}
	return nil
}

func TestCreateRoom_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
		createFunc: func(ctx context.Context, room *domain.Room) error {
			room.ID = primitive.NewObjectID()
			return nil
		},
	}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "Conference Room A", "CONF_A_001", userID)

	assert.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, "Conference Room A", room.Name)
	assert.Equal(t, "CONF_A_001", room.Code)
	assert.Equal(t, userID, room.CreatedBy)
}

func TestCreateRoom_InvalidName_Empty(t *testing.T) {
	userID := primitive.NewObjectID()
	repoMock := &MockRoomRepository{}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "", "CONF_A_001", userID)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestCreateRoom_InvalidName_TooLong(t *testing.T) {
	userID := primitive.NewObjectID()
	repoMock := &MockRoomRepository{}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "This is a very long room name that exceeds the fifty character limit", "CONF_A_001", userID)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestCreateRoom_InvalidCode_Empty(t *testing.T) {
	userID := primitive.NewObjectID()
	repoMock := &MockRoomRepository{}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "Conference Room A", "", userID)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestCreateRoom_InvalidCode_InvalidCharacters(t *testing.T) {
	userID := primitive.NewObjectID()
	repoMock := &MockRoomRepository{}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "Conference Room A", "CONF@A#1", userID)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrInvalidInput))
}

func TestCreateRoom_CodeAlreadyExists(t *testing.T) {
	userID := primitive.NewObjectID()
	existingRoom := &domain.Room{
		ID:        primitive.NewObjectID(),
		Name:      "Existing Room",
		Code:      "CONF_A_001",
		CreatedBy: userID,
	}

	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == "CONF_A_001" {
				return existingRoom, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "Conference Room A", "CONF_A_001", userID)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrCodeAlreadyExists))
}

func TestCreateRoom_RepositoryError(t *testing.T) {
	userID := primitive.NewObjectID()
	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
		createFunc: func(ctx context.Context, room *domain.Room) error {
			return errors.New("database error")
		},
	}

	svc := NewRoomService(repoMock)
	room, err := svc.CreateRoom(context.Background(), "Conference Room A", "CONF_A_001", userID)

	assert.Error(t, err)
	assert.Nil(t, room)
}

func TestCreateRoom_ValidateCodeFormatWithValidInputs(t *testing.T) {
	userID := primitive.NewObjectID()
	
	testCases := []struct {
		name    string
		code    string
		shouldFail bool
	}{
		{"Alphanumeric", "ABC123", false},
		{"WithUnderscore", "CONF_A_001", false},
		{"WithHyphen", "CONF-A-001", false},
		{"WithBoth", "CONF_A-001", false},
		{"AllNumbers", "123456", false},
		{"AllLetters", "ABCDEF", false},
		{"WithSpace", "CONF A", true},
		{"WithSpecialChar", "CONF@A", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoMock := &MockRoomRepository{
				getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
					return nil, domain.ErrRoomNotFound
				},
				createFunc: func(ctx context.Context, room *domain.Room) error {
					room.ID = primitive.NewObjectID()
					return nil
				},
			}

			svc := NewRoomService(repoMock)
			room, err := svc.CreateRoom(context.Background(), "Test Room", tc.code, userID)

			if tc.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, room)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, room)
			}
		})
	}
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
		Members:   []primitive.ObjectID{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
		addUserToRoomFunc: func(ctx context.Context, rID, uID primitive.ObjectID) error {
			room.Members = append(room.Members, uID)
			return nil
		},
		getByIDFunc: func(ctx context.Context, id primitive.ObjectID) (*domain.Room, error) {
			if id == roomID {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	svc := NewRoomService(repoMock)
	updatedRoom, err := svc.AddUserToRoom(context.Background(), roomCode, userID)

	assert.NoError(t, err)
	assert.NotNil(t, updatedRoom)
	assert.Equal(t, roomID, updatedRoom.ID)
	assert.Contains(t, updatedRoom.Members, userID)
}

func TestAddUserToRoom_RoomNotFound(t *testing.T) {
	userID := primitive.NewObjectID()
	roomCode := "NONEXISTENT"

	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
	}

	svc := NewRoomService(repoMock)
	room, err := svc.AddUserToRoom(context.Background(), roomCode, userID)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrRoomNotFound))
}

func TestAddUserToRoom_RepositoryAddError(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: primitive.NewObjectID(),
		Members:   []primitive.ObjectID{},
	}

	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
		addUserToRoomFunc: func(ctx context.Context, rID, uID primitive.ObjectID) error {
			return errors.New("database error")
		},
	}

	svc := NewRoomService(repoMock)
	updatedRoom, err := svc.AddUserToRoom(context.Background(), roomCode, userID)

	assert.Error(t, err)
	assert.Nil(t, updatedRoom)
}

func TestGetRoom_Success(t *testing.T) {
	roomID := primitive.NewObjectID()
	roomCode := "CONF_A_001"
	now := time.Now()

	room := &domain.Room{
		ID:        roomID,
		Name:      "Conference Room A",
		Code:      roomCode,
		CreatedBy: primitive.NewObjectID(),
		Members:   []primitive.ObjectID{primitive.NewObjectID()},
		CreatedAt: now,
		UpdatedAt: now,
	}

	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			if code == roomCode {
				return room, nil
			}
			return nil, domain.ErrRoomNotFound
		},
	}

	svc := NewRoomService(repoMock)
	retrievedRoom, err := svc.GetRoom(context.Background(), roomCode)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedRoom)
	assert.Equal(t, roomCode, retrievedRoom.Code)
}

func TestGetRoom_NotFound(t *testing.T) {
	roomCode := "NONEXISTENT"

	repoMock := &MockRoomRepository{
		getByCodeFunc: func(ctx context.Context, code string) (*domain.Room, error) {
			return nil, domain.ErrRoomNotFound
		},
	}

	svc := NewRoomService(repoMock)
	room, err := svc.GetRoom(context.Background(), roomCode)

	assert.Error(t, err)
	assert.Nil(t, room)
	assert.True(t, errors.Is(err, domain.ErrRoomNotFound))
}
